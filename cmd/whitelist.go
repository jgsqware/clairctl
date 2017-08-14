package cmd

import (
	"io/ioutil"
	"gopkg.in/yaml.v2"
	"github.com/jgsqware/clairctl/clair"
	"github.com/coreos/clair/api/v1"
	"strings"
)

func NewWhiteList(path string) *WhiteList {
	w := &WhiteList{path: path}
	w.init()
	return w
}

type vulnerabilitiesWhitelist struct {
	GeneralWhitelist map[string]string
	Images           map[string]map[string]string
}

type WhiteList struct {
	path        string
	whitelisted vulnerabilitiesWhitelist
}

func (v *WhiteList) init() {
	v.whitelisted = vulnerabilitiesWhitelist{}
	whitelistBytes, err := ioutil.ReadFile(v.path)
	if err != nil {
		log.Fatalf("file error %s", err)
	}
	err = yaml.Unmarshal(whitelistBytes, &v.whitelisted)
	if err != nil {
		log.Fatalf("unmarsh error %s : %v", v.path, err)
	}
}

func (v *WhiteList) filter(analysis clair.ImageAnalysis) {

	//access by ref to not reconstruct the whole struc

	for indexLayerEnvelope := range analysis.Layers {
		for indexFeature := range analysis.Layers[indexLayerEnvelope].Layer.Features {

			filteredVulnerabilities := []v1.Vulnerability{}
			for _, vulnerability := range analysis.Layers[indexLayerEnvelope].Layer.Features[indexFeature].Vulnerabilities {
				if ! v.isWhiteListed(vulnerability.NamespaceName, vulnerability.Name) {
					filteredVulnerabilities = append(filteredVulnerabilities, vulnerability)
				} else {
					log.Debugf("Whitelisted vulnerability %s:%s", vulnerability.NamespaceName, vulnerability.Name)
				}
			}

			analysis.Layers[indexLayerEnvelope].Layer.Features[indexFeature].Vulnerabilities = filteredVulnerabilities
		}
	}
}

func (v *WhiteList) isWhiteListed(namespace string, vulnerability string) bool {
	imagesWhitelist := v.getImageWhiteList(namespace)

	if _, exists := imagesWhitelist[vulnerability]; exists {
		return true
	}

	if _, exists := v.whitelisted.GeneralWhitelist[vulnerability]; exists {
		return true
	}

	return false
}

func (v *WhiteList) getImageWhiteList(namespace string) map[string]string {
	var vulnerabilities map[string]string
	imageWithoutVersion := strings.Split(namespace, ":")

	if val, exists := v.whitelisted.Images[imageWithoutVersion[0]]; exists {
		vulnerabilities = val
	}

	return vulnerabilities
}
