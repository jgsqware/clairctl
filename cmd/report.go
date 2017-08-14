package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/jgsqware/clairctl/clair"
	"github.com/jgsqware/clairctl/config"
	"github.com/jgsqware/clairctl/docker"
	"github.com/jgsqware/clairctl/xstrings"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var reportCmd = &cobra.Command{
	Use:   "report IMAGE",
	Short: "Generate Docker Image vulnerabilities report",
	Long:  `Generate Docker Image vulnerabilities report as HTML or JSON`,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) != 1 {
			fmt.Printf("clairctl: \"report\" requires a minimum of 1 argument")
			os.Exit(1)
		}

		config.ImageName = args[0]
		image, manifest, err := docker.RetrieveManifest(config.ImageName, true)

		if err != nil {
			fmt.Println(errInternalError)
			log.Fatalf("retrieving manifest for %q: %v", config.ImageName, err)
		}

		analyzes := clair.Analyze(image, manifest)

		if whitelistConfig != "" {
			whiteListProcessor := NewWhiteList(whitelistConfig)
			whiteListProcessor.filter(analyzes)
		}

		imageName := strings.Replace(analyzes.ImageName, "/", "-", -1)
		if analyzes.Tag != "" {
			imageName += "-" + analyzes.Tag
		}

		switch clair.Report.Format {
		case "html":
			html, err := clair.ReportAsHTML(analyzes)
			if err != nil {
				fmt.Println(errInternalError)
				log.Fatalf("generating HTML report: %v", err)
			}
			err = clair.SaveReport(imageName, string(html))
			if err != nil {
				fmt.Println(errInternalError)
				log.Fatalf("saving HTML report: %v", err)
			}

		case "json":
			json, err := xstrings.ToIndentJSON(analyzes)

			if err != nil {
				fmt.Println(errInternalError)
				log.Fatalf("indenting JSON: %v", err)
			}
			err = clair.SaveReport(imageName, string(json))
			if err != nil {
				fmt.Println(errInternalError)
				log.Fatalf("saving JSON report: %v", err)
			}

		default:
			fmt.Printf("Unsupported Report format: %v", clair.Report.Format)
			log.Fatalf("Unsupported Report format: %v", clair.Report.Format)
		}
	},
}

func init() {
	RootCmd.AddCommand(reportCmd)
	reportCmd.Flags().BoolVarP(&config.IsLocal, "local", "l", false, "Use local images")
	reportCmd.Flags().StringP("format", "f", "html", "Format for Report [html,json]")
	reportCmd.Flags().StringVarP(&whitelistConfig, "whitelist", "w", "", "YAML Configuration file for severity whitelisting")
	viper.BindPFlag("clair.report.format", reportCmd.Flags().Lookup("format"))
}
