package clair

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/Sirupsen/logrus"
	"github.com/coreos/clair/api/v1"
	"github.com/docker/distribution"
	"github.com/docker/distribution/manifest/schema1"
	"github.com/docker/distribution/manifest/schema2"
	"github.com/docker/docker/reference"
	"github.com/jgsqware/clairctl/config"
	"github.com/jgsqware/clairctl/xstrings"
)

// ErrUnanalizedLayer is returned when the layer was not correctly analyzed
var ErrUnanalizedLayer = errors.New("layer cannot be analyzed")

var registryMapping map[string]string

func Push(image reference.Named, manifest distribution.Manifest) error {
	layers, err := newLayering(image)
	if err != nil {
		return err
	}

	switch manifest.(type) {
	case *schema1.SignedManifest:
		for _, l := range manifest.(*schema1.SignedManifest).FSLayers {
			layers.digests = append(layers.digests, l.BlobSum.String())
		}
		return layers.pushAll()
	case *schema2.DeserializedManifest:
		for _, l := range manifest.(*schema2.DeserializedManifest).Layers {
			layers.digests = append(layers.digests, l.Digest.String())
		}
		return layers.pushAll()
	default:
		return nil
	}
}

type layering struct {
	image          reference.Named
	digests        []string
	parentID, hURL string
}

func (layer *layering) pushAll() error {
	layerCount := len(layer.digests)

	if layerCount == 0 {
		logrus.Warningln("there is no layer to push")
	}
	for index, digest := range layer.digests {

		if config.IsLocal {
			digest = strings.TrimPrefix(digest, "sha256:")
		}

		lUID := xstrings.Substr(digest, 0, 12)
		logrus.Infof("Pushing Layer %d/%d [%v]", index+1, layerCount, lUID)

		insertRegistryMapping(digest, layer.image.Hostname())
		payload := v1.LayerEnvelope{Layer: &v1.Layer{
			Name:       digest,
			Path:       blobsURI(layer.image.Hostname(), layer.image.RemoteName(), digest),
			ParentName: layer.parentID,
			Format:     "Docker",
		}}

		//FIXME Update to TLS
		if config.IsLocal {
			payload.Layer.Path += "/layer.tar"
		}
		payload.Layer.Path = strings.Replace(payload.Layer.Path, layer.image.Hostname(), layer.hURL, 1)
		if err := pushLayer(payload); err != nil {
			logrus.Infof("adding layer %d/%d [%v]: %v", index+1, layerCount, lUID, err)
			if err != ErrUnanalizedLayer {
				return err
			}
			layer.parentID = ""
		} else {
			layer.parentID = payload.Layer.Name
		}
	}
	return nil
}

func newLayering(image reference.Named) (*layering, error) {
	layer := layering{
		parentID: "",
		image:    image,
	}

	localIP, err := config.LocalServerIP()
	if err != nil {
		return nil, err
	}
	layer.hURL = fmt.Sprintf("http://%v/v2", localIP)
	if config.IsLocal {
		layer.hURL = strings.Replace(layer.hURL, "/v2", "/local", -1)
		logrus.Infof("using %v as local url", layer.hURL)
	}
	return &layer, nil
}

func pushLayer(layer v1.LayerEnvelope) error {
	lJSON, err := json.Marshal(layer)
	if err != nil {
		return fmt.Errorf("marshalling layer: %v", err)
	}

	lURI := fmt.Sprintf("%v/layers", uri)
	request, err := http.NewRequest("POST", lURI, bytes.NewBuffer(lJSON))
	if err != nil {
		return fmt.Errorf("creating 'add layer' request: %v", err)
	}
	request.Header.Set("Content-Type", "application/json")

	response, err := (&http.Client{}).Do(request)

	if err != nil {
		return fmt.Errorf("pushing layer to clair: %v", err)
	}
	defer response.Body.Close()

	if response.StatusCode != 201 {
		if response.StatusCode == 422 {
			return ErrUnanalizedLayer
		}
		return fmt.Errorf("receiving http error: %d", response.StatusCode)
	}

	return nil
}

func blobsURI(registry string, name string, digest string) string {
	return strings.Join([]string{registry, name, "blobs", digest}, "/")
}

func insertRegistryMapping(layerDigest string, registryURI string) {

	if registryURI == "docker.io" {
		registryURI = "registry-1." + registryURI
	}
	if strings.Contains(registryURI, "docker") {
		registryURI = "https://" + registryURI + "/v2"

	} else {
		registryURI = "http://" + registryURI + "/v2"
	}
	logrus.Debugf("Saving %s[%s]", layerDigest, registryURI)
	registryMapping[layerDigest] = registryURI
}

//GetRegistryMapping return the registryURI corresponding to the layerID passed as parameter
func GetRegistryMapping(layerDigest string) (string, error) {
	registryURI, present := registryMapping[layerDigest]
	if !present {
		return "", fmt.Errorf("%v mapping not found", layerDigest)
	}
	return registryURI, nil
}

func init() {
	registryMapping = map[string]string{}
}
