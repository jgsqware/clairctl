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
	w.whitelisted = vulnerabilitiesWhitelist{}

	whitelistBytes, err := ioutil.ReadFile(w.path)
	if err != nil {
		log.Fatalf("file error %s", err)
	}

	err = yaml.Unmarshal(whitelistBytes, &w.whitelisted)
	if err != nil {
		log.Fatalf("unmarsh error %s : %v", w.path, err)
	}

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

func (v *WhiteList) filter(analysis clair.ImageAnalysis) {

	//access by ref to not reconstruct the whole struc

	for i := range analysis.Layers {
		for f := range analysis.Layers[i].Layer.Features {

			filteredVulnerabilities := []v1.Vulnerability{}
			for _, vulnerability := range analysis.Layers[i].Layer.Features[f].Vulnerabilities {
				if ! v.isWhiteListed(vulnerability.NamespaceName, vulnerability.Name) {
					filteredVulnerabilities = append(filteredVulnerabilities, vulnerability)
				} else {
					log.Debugf("Whitelisted vulnerability %s:%s", vulnerability.NamespaceName, vulnerability.Name)
				}
			}

			analysis.Layers[i].Layer.Features[f].Vulnerabilities = filteredVulnerabilities
		}
	}
}

func (v *WhiteList) isWhiteListed(namespace string, vulnerability string) bool {

	if _, exists := v.getImageWhiteList(namespace)[vulnerability]; exists {
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
