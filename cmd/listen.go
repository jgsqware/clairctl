package cmd

import (
	"fmt"

	"github.com/jgsqware/clairctl/config"
	"github.com/jgsqware/clairctl/server"
	"github.com/spf13/cobra"
)

var listenCmd = &cobra.Command{
	Use:   "listen",
	Short: "Starts the server to help troubleshooting communication issues",
	Long:  `Starts the server to help troubleshooting communication issues`,
	Run: func(cmd *cobra.Command, args []string) {
		sURL, err := config.LocalServerIP()
		if err != nil {
			fmt.Println(errInternalError)
			log.Fatalf("retrieving internal server IP: %v", err)
		}
		log.Fatal(server.ServeOK(sURL))
	},
}

func init() {
	RootCmd.AddCommand(listenCmd)
}
