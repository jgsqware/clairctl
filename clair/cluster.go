package clair

import (
	"bytes"
	"fmt"
	"net/http"

	"github.com/spf13/viper"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/pkg/api/v1"
	"k8s.io/client-go/rest"
)

func ClusterScan() map[string]struct{} {
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
	return images
}

func Notify(analyzes ImageAnalysis) {
	vulnerabilities := AllVulnerabilities(analyzes)

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

func postNotification(severity string, analyzes ImageAnalysis, endpoint string) {
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
