package cmd

import (
	"fmt"
	"strings"

	"github.com/jgsqware/clairctl/clair"
	"github.com/jgsqware/clairctl/config"
	"github.com/jgsqware/clairctl/docker"
	"github.com/jgsqware/clairctl/xstrings"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var clusterReportCmd = &cobra.Command{
	Use:   "report",
	Short: "Generate Docker Images vulnerabilities report",
	Long:  `Generate Docker Images vulnerabilities report as HTML or JSON`,
	Run: func(cmd *cobra.Command, args []string) {
		images := config.ClusterImages
		for imageName := range images {
			config.ImageName = imageName
			image, manifest, err := docker.RetrieveManifest(config.ImageName, true)

			if err != nil {
				fmt.Println(errInternalError)
				log.Fatalf("retrieving manifest for %q: %v", config.ImageName, err)
			}

			analyzes := clair.Analyze(image, manifest)
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
			if viper.GetString("notifier.endpoint") != "" {
				clair.Notify(analyzes)
			}
		}
	},
}

func init() {
	clusterCmd.AddCommand(clusterReportCmd)
	viper.BindPFlag("clair.report.format", clusterReportCmd.Flags().Lookup("format"))
}
