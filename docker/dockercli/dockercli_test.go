package dockercli

import (
	"testing"

	"github.com/docker/docker/reference"
)

func TestImageParsing(t *testing.T) {
	images := map[string]string{
		"ubuntu:14.04":                          "docker.io/library/ubuntu/14_04",
		"ubuntu/ubuntu:14.04":                   "docker.io/ubuntu/ubuntu/14_04",
		"registry.com/ubuntu:14.04":             "registry.com/ubuntu/14_04",
		"registry.com/ubuntu/ubuntu:14.04":      "registry.com/ubuntu/ubuntu/14_04",
		"registry.com:5000/ubuntu:14.04":        "registry.com:5000/ubuntu/14_04",
		"registry.com:5000/ubuntu/ubuntu:14.04": "registry.com:5000/ubuntu/ubuntu/14_04",
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

		result := tempImagePath(image)

		if result != expected {
			t.Errorf("Expecting %s, got %s", expected, result)
		}
	}
}
