package cmd

import (
	"fmt"

	"github.com/jgsqware/clairctl/clair"
	"github.com/jgsqware/clairctl/config"
	"github.com/jgsqware/clairctl/docker"
	"github.com/spf13/cobra"
)

var clusterPushCmd = &cobra.Command{
	Use:   "push",
	Short: "Push Docker images to Clair",
	Long:  `Upload a Docker images to Clair for further analysis`,
	Run: func(cmd *cobra.Command, args []string) {
		images := config.ClusterImages
		for imageName := range images {
			if config.IsLocal {
				startLocalServer()
			}
			config.ImageName = imageName
			image, manifest, err := docker.RetrieveManifest(config.ImageName, true)
			if err != nil {
				fmt.Println(errInternalError)
				log.Fatalf("retrieving manifest for %q: %v", config.ImageName, err)
			}

			if err := clair.Push(image, manifest); err != nil {
				if err != nil {
					fmt.Println(errInternalError)
					log.Fatalf("pushing image %q: %v", image.String(), err)
				}
			}
			fmt.Printf("%v has been pushed to Clair\n", image.String())
		}
	},
}

func init() {
	clusterCmd.AddCommand(clusterPushCmd)
	clusterPushCmd.Flags().BoolVarP(&config.IsLocal, "local", "l", false, "Use local images")
}
