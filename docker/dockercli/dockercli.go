package dockercli

import (
	"compress/bzip2"
	"compress/gzip"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"
	"syscall"

	"github.com/artyom/untar"
	"github.com/coreos/pkg/capnslog"
	"github.com/docker/distribution"
	"github.com/docker/distribution/manifest/schema1"
	"github.com/docker/docker/client"
	"github.com/docker/docker/image"
	"github.com/docker/docker/layer"
	"github.com/docker/docker/reference"
	"github.com/jgsqware/clairctl/config"
	"github.com/opencontainers/go-digest"
)

var log = capnslog.NewPackageLogger("github.com/jgsqware/clairctl", "dockercli")

func parseImage(imageName string) (reference.NamedTagged, error) {

	n, err := reference.ParseNamed(imageName)
	if err != nil {
		return nil, err
	}
	var image reference.NamedTagged
	if reference.IsNameOnly(n) {
		image = reference.WithDefaultTag(n).(reference.NamedTagged)
	} else {
		image = n.(reference.NamedTagged)
	}

	return image, nil

}

func tempImagePath(image reference.NamedTagged) string {

	return strings.Replace(strings.Replace(fmt.Sprintf("%s/%s", image.FullName(), image.Tag()), ".", "_", -1), ":", "_", -1)
}

//GetLocalManifest retrieve manifest for local image
func GetLocalManifest(imageName string, withExport bool) (reference.NamedTagged, distribution.Manifest, error) {
	image, err := parseImage(imageName)
	if err != nil {
		return nil, nil, err
	}
	var manifest distribution.Manifest
	if withExport {
		manifest, err = save(image)
	} else {
		manifest, err = historyFromCommand(image)
	}

	if err != nil {
		return nil, schema1.SignedManifest{}, err
	}
	m := manifest.(schema1.SignedManifest)
	m.Name = image.Name()
	m.Tag = image.Tag()
	return image, m, err
}

func save(image reference.NamedTagged) (distribution.Manifest, error) {

	path := config.TmpLocal() + "/" + tempImagePath(image) + "/blobs"

	if _, err := os.Stat(path); os.IsExist(err) {
		err := os.RemoveAll(path)
		if err != nil {
			return nil, err
		}
	}

	err := os.MkdirAll(path, 0755)
	if err != nil {
		return nil, err
	}

	log.Debug("docker image to save: ", image.String())
	log.Debug("saving in: ", path)

	cli, err := client.NewEnvClient()
	if err != nil {
		panic(err)
	}

	img, err := cli.ImageSave(context.Background(), []string{image.String()})
	if err != nil {
		return nil, fmt.Errorf("cannot save image %s: %s", image.String(), err)
	}
	defer img.Close()

	fo, err := os.Create(path + "/output.tar")
	// close fo on exit and check for its returned error
	defer func() {
		if err := fo.Close(); err != nil {
			panic(err)
		}
	}()

	if err != nil {
		log.Error(err)
		return nil, err
	}

	// make a buffer to keep chunks that are read
	buf := make([]byte, 1024)
	for {
		// read a chunk
		n, err := img.Read(buf)
		if err != nil && err != io.EOF {
			log.Error(err)
			return nil, err
		}
		if n == 0 {
			break
		}

		// write a chunk
		if _, err := fo.Write(buf[:n]); err != nil {
			log.Error(err)
			return nil, err
		}
	}

	err = openAndUntar(path+"/output.tar", path)
	if err != nil {
		log.Error(err)
		return nil, err
	}

	err = os.Remove(path + "/output.tar")
	if err != nil {
		log.Error(err)
		return nil, err
	}
	return historyFromManifest(path)
}

func historyFromManifest(path string) (distribution.Manifest, error) {
	mf, err := os.Open(path + "/manifest.json")
	defer mf.Close()

	if err != nil {
		return schema1.SignedManifest{}, err
	}

	// https://github.com/docker/docker/blob/master/image/tarexport/tarexport.go#L17
	type manifestItem struct {
		Config       string
		RepoTags     []string
		Layers       []string
		Parent       image.ID                                 `json:",omitempty"`
		LayerSources map[layer.DiffID]distribution.Descriptor `json:",omitempty"`
	}

	var manifest []manifestItem
	if err = json.NewDecoder(mf).Decode(&manifest); err != nil {
		return schema1.SignedManifest{}, err
	} else if len(manifest) != 1 {
		return schema1.SignedManifest{}, err
	}
	var layers []string
	for _, layer := range manifest[0].Layers {
		layers = append(layers, strings.TrimSuffix(layer, "/layer.tar"))
	}
	var m schema1.SignedManifest

	for _, layer := range manifest[0].Layers {
		var d digest.Digest
		d, err := digest.Parse("sha256:" + strings.TrimSuffix(layer, "/layer.tar"))
		if err != nil {
			return schema1.SignedManifest{}, err
		}
		m.FSLayers = append(m.FSLayers, schema1.FSLayer{BlobSum: d})
	}

	return m, nil
}

func historyFromCommand(image reference.NamedTagged) (schema1.SignedManifest, error) {

	client, err := client.NewEnvClient()
	if err != nil {
		return schema1.SignedManifest{}, err
	}
	histories, err := client.ImageHistory(context.Background(), image.String())

	manifest := schema1.SignedManifest{}
	for _, history := range histories {
		var d digest.Digest
		d, err := digest.Parse(history.ID)
		if err != nil {
			return schema1.SignedManifest{}, err
		}
		manifest.FSLayers = append(manifest.FSLayers, schema1.FSLayer{BlobSum: d})
	}
	return manifest, nil
}

func openAndUntar(name, dst string) error {
	var rd io.Reader
	f, err := os.Open(name)
	defer f.Close()

	if err != nil {
		return err
	}
	rd = f
	if strings.HasSuffix(name, ".gz") || strings.HasSuffix(name, ".tgz") {
		gr, err := gzip.NewReader(f)
		if err != nil {
			return err
		}
		defer gr.Close()
		rd = gr
	} else if strings.HasSuffix(name, ".bz2") {
		rd = bzip2.NewReader(f)
	}
	if err := os.MkdirAll(dst, os.ModeDir|os.ModePerm); err != nil {
		return err
	}
	// resetting umask is essential to have exact permissions on unpacked
	// files; it's not not put inside untar function as it changes
	// process-wide umask
	mask := syscall.Umask(0)
	defer syscall.Umask(mask)
	return untar.Untar(rd, dst)
}
