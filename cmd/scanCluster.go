package cmd

import (
	"fmt"
	"os"
	"text/template"
	"strings"

	"github.com/jgsqware/clairctl/clair"
	"github.com/jgsqware/clairctl/config"
	"github.com/jgsqware/clairctl/docker"
	"github.com/jgsqware/clairctl/xstrings"
	"github.com/spf13/cobra"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/pkg/api/v1"
	"github.com/spf13/viper"
)

var scanCmd = &cobra.Command{
	Use:   "scanCluster",
	Short: "Scan and analyze all Docker images in cluster",
	Long:  `Scan and analyze all Docker images in cluster, against Ubuntu, Red hat and Debian vulnerabilities databases`,
	Run: func(cmd *cobra.Command, args []string) {

		clusterConfig, err := rest.InClusterConfig()
		if err != nil {
			log.Fatalf("Error scanCluster must be run from inside a cluster %v", err)
		}
		// creates the clientset
		clientset, err := kubernetes.NewForConfig(clusterConfig)
		if err != nil {
			panic(err.Error())
		}
		pods, err := clientset.CoreV1().Pods("").List(v1.ListOptions{})
		if err != nil {
			panic(err.Error())
		}
		fmt.Printf("There are %d pods in the cluster\n", len(pods.Items))

		images := make(map[string]struct{})

		for _, pod := range pods.Items {
			for _, container := range pod.Spec.Containers {
				images[container.Image] = struct{}{}
			}
		}

		fmt.Printf("list of images %v", images)

		for imageName := range (images) {
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

			analysis := clair.Analyze(image, manifest)
			err = template.Must(template.New("analysis").Parse(analyzeTplt)).Execute(os.Stdout, analysis)
			if err != nil {
				fmt.Println(errInternalError)
				log.Fatalf("rendering analysis: %v", err)
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
		}
	},
}

func init() {
	RootCmd.AddCommand(scanCmd)
	reportCmd.Flags().StringP("format", "f", "html", "Format for Report [html,json]")
	viper.BindPFlag("clair.report.format", reportCmd.Flags().Lookup("format"))
}
