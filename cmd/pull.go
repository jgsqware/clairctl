package cmd

import (
	"fmt"
	"html/template"
	"os"

	"github.com/Sirupsen/logrus"
	"github.com/docker/distribution/digest"
	"github.com/docker/docker/reference"
	"github.com/jgsqware/clairctl/config"
	"github.com/jgsqware/clairctl/docker"
	"github.com/spf13/cobra"
)

// const pullTplt = `
// Image: {{.Named.FullName}}
//  {{.V1Manifest.FSLayers | len}} layers found
//  {{range .V1Manifest.FSLayers}} ➜ {{.BlobSum}}
//  {{end}}
// `
const pullTplt = `
Image: {{.Named.FullName}}
 {{.Layers | len}} layers found
 {{range .Layers}} ➜ {{.}}
 {{end}}
`

var pullCmd = &cobra.Command{
	Use:   "pull IMAGE",
	Short: "Pull Docker image to Clair",
	Long:  `Upload a Docker image to Clair for further analysis`,
	Run: func(cmd *cobra.Command, args []string) {

		if len(args) != 1 {
			fmt.Printf("clairctl: \"pull\" requires a minimum of 1 argument\n")
			os.Exit(1)
		}

		config.ImageName = args[0]
		image, manifest, err := docker.RetrieveManifest(config.ImageName, true)
		if err != nil {
			fmt.Println(errInternalError)
			logrus.Fatalf("retrieving manifest for %q: %v", config.ImageName, err)
		}

		layers, err := docker.GetLayerDigests(manifest)
		if err != nil {
			fmt.Println(errInternalError)
			logrus.Fatalf("retrieving layers for %q: %v", config.ImageName, err)
		}
		data := struct {
			Layers []digest.Digest
			Named  reference.Named
		}{
			Layers: layers,
			Named:  image,
		}

		err = template.Must(template.New("pull").Parse(pullTplt)).Execute(os.Stdout, data)
		if err != nil {
			fmt.Println(errInternalError)
			logrus.Fatalf("rendering image: %v", err)
		}
	},
}

func init() {
	RootCmd.AddCommand(pullCmd)
	pullCmd.Flags().BoolVarP(&config.IsLocal, "local", "l", false, "Use local images")
}
