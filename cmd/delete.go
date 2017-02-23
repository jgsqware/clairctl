package cmd

import (
	"fmt"
	"os"

	"github.com/Sirupsen/logrus"
	"github.com/jgsqware/clairctl/clair"
	"github.com/jgsqware/clairctl/config"
	"github.com/jgsqware/clairctl/docker"
	"github.com/spf13/cobra"
)

var deleteCmd = &cobra.Command{
	Use:   "delete IMAGE",
	Short: "Delete Docker image",
	Long:  `Delete a Docker image from Clair`,
	Run: func(cmd *cobra.Command, args []string) {

		if len(args) != 1 {
			fmt.Printf("clairctl: \"delete\" requires a minimum of 1 argument")
			os.Exit(1)
		}

		config.ImageName = args[0]
		image, manifest, err := docker.RetrieveManifest(config.ImageName, true)
		if err != nil {
			fmt.Println(errInternalError)
			logrus.Fatalf("retrieving manifest for %q: %v", config.ImageName, err)
		}

		err = clair.Delete(image, manifest)
		if err != nil {
			fmt.Println(errInternalError)
			logrus.Fatalf("deleting layers for %q: %v", config.ImageName, err)
		}
		fmt.Printf("%v has been deleted from Clair\n", image.String())
	},
}

func init() {
	RootCmd.AddCommand(deleteCmd)
	deleteCmd.Flags().BoolVarP(&config.IsLocal, "local", "l", false, "Use local images")
}
