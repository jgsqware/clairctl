// Copyright 2016 CoreOS, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Package dockerdist provides helper methods for retrieving and parsing a
// information from a remote Docker repository.
package dockerdist

import (
	"errors"
	"reflect"

	"github.com/Sirupsen/logrus"
	distlib "github.com/docker/distribution"
	"github.com/docker/distribution/manifest/schema1"
	"github.com/docker/distribution/manifest/schema2"
	"github.com/docker/distribution/registry/api/v2"
	"github.com/docker/distribution/registry/client"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/cli/config"
	"github.com/docker/docker/distribution"
	"github.com/docker/docker/dockerversion"
	"github.com/docker/docker/reference"
	"github.com/docker/docker/registry"
	"github.com/opencontainers/go-digest"
	"github.com/spf13/viper"

	"strings"

	"golang.org/x/net/context"
)

var ErrTagNotFound = errors.New("this image or tag is not found")

func isInsecureRegistry(registryHostname string) bool {
	for _, r := range viper.GetStringSlice("docker.insecure-registries") {
		if r == registryHostname {
			return true
		}
	}

	return false
}

// getRepositoryClient returns a client for performing registry operations against the given named
// image.
func getRepositoryClient(image reference.Named, insecure bool, scopes ...string) (distlib.Repository, error) {

	logrus.Debugf("Retrieving repository client")

	serviceOptions := registry.ServiceOptions{
		InsecureRegistries: viper.GetStringSlice("docker.insecure-registries"),
	}
	service := registry.NewService(serviceOptions)
	ctx := context.Background()
	authConfig, err := GetAuthCredentials(image.String())
	if err != nil {
		logrus.Debugf("GetAuthCredentials error: %v", err)
		return nil, err
	}

	if (types.AuthConfig{}) != authConfig {

		userAgent := dockerversion.DockerUserAgent(ctx)
		_, _, err = service.Auth(ctx, &authConfig, userAgent)
		if err != nil {
			logrus.Debugf("Auth: err: %v", err)
			return nil, err
		}
	}

	repoInfo, err := service.ResolveRepository(image)
	if err != nil {
		logrus.Debugf("ResolveRepository err: %v", err)
		return nil, err
	}

	metaHeaders := map[string][]string{}

	endpoints, err := service.LookupPullEndpoints(image.Hostname())

	if err != nil {
		logrus.Debugf("registry.LookupPullEndpoints error: %v", err)
		return nil, err
	}

	var confirmedV2 bool
	var repository distlib.Repository

	for _, endpoint := range endpoints {
		if confirmedV2 && endpoint.Version == registry.APIVersion1 {
			logrus.Debugf("Skipping v1 endpoint %s because v2 registry was detected", endpoint.URL)
			continue
		}

		if isInsecureRegistry(endpoint.URL.Host) {
			//endpoint.TLSConfig.InsecureSkipVerify = viper.GetBool("auth.insecureSkipVerify")
			endpoint.URL.Scheme = "http"
		}

		repository, confirmedV2, err = distribution.NewV2Repository(ctx, repoInfo, endpoint, metaHeaders, &authConfig, scopes...)
		if err != nil {
			logrus.Debugf("cannot instanciate new v2 repository on %v", endpoint.URL)
			return nil, err
		}

		if !confirmedV2 {
			return nil, errors.New("Only V2 repository are supported")
		}
		break
	}

	return repository, nil
}

// getDigest returns the digest for the given image.
func getDigest(ctx context.Context, repo distlib.Repository, image reference.Named) (digest.Digest, error) {
	if withDigest, ok := image.(reference.Canonical); ok {
		return withDigest.Digest(), nil
	}
	// Get TagService.
	tagSvc := repo.Tags(ctx)

	// Get Tag name.
	tag := "latest"
	if withTag, ok := image.(reference.NamedTagged); ok {
		tag = withTag.Tag()
	}

	// Get Tag's Descriptor.
	descriptor, err := tagSvc.Get(ctx, tag)
	if err != nil {

		// Docker returns an UnexpectedHTTPResponseError if it cannot parse the JSON body of an
		// unexpected error. Unfortunately, HEAD requests *by definition* don't have bodies, so
		// Docker will return this error for non-200 HEAD requests. We therefore have to hack
		// around it... *sigh*.
		if _, ok := err.(*client.UnexpectedHTTPResponseError); ok {
			return "", errors.New("Received error when trying to fetch the specified tag: it might not exist or you do not have access")
		}

		if strings.Contains(err.Error(), v2.ErrorCodeManifestUnknown.Message()) {
			return "", ErrTagNotFound
		}

		return "", err
	}

	return descriptor.Digest, nil
}

// GetAuthCredentials returns the auth credentials (if any found) for the given repository, as found
// in the user's docker config.
func GetAuthCredentials(image string) (types.AuthConfig, error) {
	// Lookup the index information for the name.
	indexInfo, err := registry.ParseSearchIndexInfo(image)
	if err != nil {
		return types.AuthConfig{}, err
	}
	// Retrieve the user's Docker configuration file (if any).
	configFile, err := config.Load(config.Dir())
	if err != nil {
		return types.AuthConfig{}, err
	}

	// Resolve the authentication information for the registry specified, via the config file.
	return registry.ResolveAuthConfig(configFile.AuthConfigs, indexInfo), nil
}

// DownloadManifest the manifest for the given image, using the given credentials.
func DownloadManifest(image string, insecure bool) (reference.NamedTagged, distlib.Manifest, error) {
	logrus.Debugf("Downloading manifest for %v", image)
	// Parse the image name as a docker image reference.
	n, err := reference.ParseNamed(image)
	if err != nil {
		return nil, nil, err
	}
	if reference.IsNameOnly(n) {
		n, _ = reference.ParseNamed(image + ":" + reference.DefaultTag)
	}

	named := n.(reference.NamedTagged)

	// Create a reference to a repository client for the repo.
	repo, err := getRepositoryClient(named, insecure, "pull")
	if err != nil {
		return nil, nil, err
	}
	// Get the digest.
	ctx := context.Background()

	digest, err := getDigest(ctx, repo, named)
	if err != nil {
		return nil, nil, err
	}

	// Retrieve the manifest for the tag.
	manSvc, err := repo.Manifests(ctx)
	if err != nil {
		return nil, nil, err
	}
	manifest, err := manSvc.Get(ctx, digest)
	if err != nil {
		return nil, nil, err
	}

	// Verify the manifest if it's signed.
	logrus.Debugf("manifest type: %v", reflect.TypeOf(manifest))

	switch manifest.(type) {
	case *schema1.SignedManifest:
		_, verr := schema1.Verify(manifest.(*schema1.SignedManifest))
		if verr != nil {
			return nil, nil, verr
		}
	case *schema2.DeserializedManifest:
		logrus.Debugf("retrieved schema2 manifest, no verification")
	default:
		logrus.Printf("Could not verify manifest for image %v: not signed", image)
	}

	return named, manifest, nil
}

// DownloadV1Manifest the manifest for the given image in v1 schema format, using the given credentials.
func DownloadV1Manifest(imageName string, insecure bool) (reference.NamedTagged, schema1.SignedManifest, error) {
	image, manifest, err := DownloadManifest(imageName, insecure)

	if err != nil {
		return nil, schema1.SignedManifest{}, err
	}
	// Ensure that the manifest type is supported.
	switch manifest.(type) {
	case *schema1.SignedManifest:
		return image, *manifest.(*schema1.SignedManifest), nil
	default:
		return nil, schema1.SignedManifest{}, errors.New("only v1 manifests are currently supported")
	}
}
