package dockercli

import (
	"fmt"
	"testing"

	"github.com/docker/docker/reference"
)

func TestImageParsing(t *testing.T) {
	images := map[string]string{
		"ubuntu:14.04":                          "dockerio/library/ubuntu/1404",
		"ubuntu/ubuntu:14.04":                   "dockerio/ubuntu/ubuntu/1404",
		"registry.com/ubuntu:14.04":             "registrycom/ubuntu/1404",
		"registry.com/ubuntu/ubuntu:14.04":      "registrycom/ubuntu/ubuntu/1404",
		"registry.com:5000/ubuntu:14.04":        "registrycom5000/ubuntu/1404",
		"registry.com:5000/ubuntu/ubuntu:14.04": "registrycom5000/ubuntu/ubuntu/1404",
	}

	for value, expected := range images {

		n, err := reference.ParseNamed(value)
		if err != nil {
			t.Error("Error:", err, expected)
		}
		var image reference.NamedTagged
		if reference.IsNameOnly(n) {
			image = reference.WithDefaultTag(n).(reference.NamedTagged)
		} else {
			image = n.(reference.NamedTagged)
		}
		fmt.Println("Name(): ", image.Name())
		fmt.Println("String(): ", image.String())
		fmt.Println("FullName(): ", image.FullName())
		fmt.Println("Hostname(): ", image.Hostname())
		fmt.Println("RemoteName(): ", image.RemoteName())
		fmt.Println("Tag(): ", image.Tag())

		result := tempImagePath(image)

		if result != expected {
			t.Errorf("Expecting %s, got %s", expected, result)
		}
	}
}
