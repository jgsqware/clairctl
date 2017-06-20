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
	"bytes"
	"net/http"
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

		images := make(map[string]struct{})

		for _, pod := range pods.Items {
			for _, container := range pod.Spec.Containers {
				images[container.Image] = struct{}{}
			}
		}

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
				checkAndNotify(analyzes)
			}

		}
	},
}

func init() {
	RootCmd.AddCommand(scanCmd)
}

func checkAndNotify(analyzes clair.ImageAnalysis) {
	vulnerabilities := clair.AllVulnerabilities(analyzes)

	endpoint := viper.GetString("notifier.endpoint")
	severity := viper.GetString("notifier.severity")

	if vulnerabilities.Count("Critical") != 0 {
		postNotification(severity, analyzes, endpoint)
	}

	if severity == "Critical" {
		return
	}

	if vulnerabilities.Count("High") != 0 {
		postNotification("High", analyzes, endpoint)
	}

	if severity == "High" {
		return
	}

	if vulnerabilities.Count("Medium") != 0 {
		postNotification("Medium", analyzes, endpoint)
	}

	if severity == "Medium" {
		return
	}
	if vulnerabilities.Count("Low") != 0 {
		postNotification("Low", analyzes, endpoint)
	}

	if severity == "Low" {
		return
	}
	if vulnerabilities.Count("Negligible") != 0 {
		postNotification("Negligible", analyzes, endpoint)
	}
}

func postNotification(severity string, analyzes clair.ImageAnalysis, endpoint string) {
	var jsonStr = []byte(fmt.Sprintf(`{"text":"There are some %s vulnerabilites in the image %s"}`,
		severity, analyzes.ImageName))
	req, err := http.NewRequest("POST", endpoint, bytes.NewBuffer(jsonStr))
	if err != nil {
		log.Infof("Err posting notification' request: %v", err)
	}
	response, err := (&http.Client{}).Do(req)
	if err != nil {
		log.Infof("Err posting notification: %v", err)
	}
	defer response.Body.Close()
	if response.StatusCode != 200 {
		log.Infof("Err posting notification: returned %v statusCode", response.StatusCode)
	}
}
