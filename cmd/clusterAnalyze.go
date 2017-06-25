package cmd

import (
	"fmt"
	"html/template"
	"os"

	"github.com/jgsqware/clairctl/clair"
	"github.com/jgsqware/clairctl/config"
	"github.com/jgsqware/clairctl/docker"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var analyzeClusterCmd = &cobra.Command{
	Use:   "analyze",
	Short: "Analyze Docker images",
	Long:  `Analyze Docker images with Clair, against Ubuntu, Red hat and Debian vulnerabilities databases`,
	Run: func(cmd *cobra.Command, args []string) {
		images := config.ClusterImages
		for imageName := range images {
			config.ImageName = imageName
			image, manifest, err := docker.RetrieveManifest(imageName, true)
			if err != nil {
				fmt.Println(errInternalError)
				log.Fatalf("retrieving manifest for %q: %v", imageName, err)
			}

			if err := clair.Push(image, manifest); err != nil {
				if err != nil {
					fmt.Println(errInternalError)
					log.Fatalf("pushing image %q: %v", image.String(), err)
				}
			}

			analyzes := clair.Analyze(image, manifest)
			err = template.Must(template.New("analysis").Parse(analyzeTplt)).Execute(os.Stdout, analyzes)
			if err != nil {
				fmt.Println(errInternalError)
				log.Fatalf("rendering analysis: %v", err)
			}
			if viper.GetString("notifier.endpoint") != "" {
				clair.Notify(analyzes)
			}
		}
	},
}

func init() {
	clusterCmd.AddCommand(analyzeClusterCmd)
}
