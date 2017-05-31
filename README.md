# clairctl
[![Build Status](https://travis-ci.org/jgsqware/clairctl.svg?branch=master)](https://travis-ci.org/jgsqware/clairctl)
[![Gitter chat](https://badges.gitter.im/gitterHQ/gitter.png)](https://gitter.im/clairctl/Lobby?utm_source=share-link&utm_medium=link&utm_campaign=share-link)

> Tracking container vulnerabilities with Clair Control

Clairctl is a lightweight command-line tool doing the bridge between Registries as Docker Hub, Docker Registry or Quay.io, and the CoreOS vulnerability tracker, Clair.
Clairctl will play as reverse proxy for authentication.

Clairctl version is align with the [CoreOS Clair](https://github.com/coreos/clair) supported version.

# Installation

  ## Released version:
  Go to [Release](https://github.com/jgsqware/clairctl/releases) and download your corresponding version

  ## Master branch version

    curl -L https://raw.githubusercontent.com/jgsqware/clairctl/master/install.sh | sh
    

# Usage

[![asciicast](https://asciinema.org/a/41461.png)](https://asciinema.org/a/41461)

# Reporting

**clairctl** get vulnerabilities report from Clair and generate HTML report

clairctl can be used for Docker Hub and self-hosted Registry

# Build from source

## Cross compile from a mac

Install go with cross compilers options:
`$ brew install go --with-cc-common`

The above `--with-cc-common` can be replaced with `--with-cc-all` if you need to compile for other archs.

Install glide
`$ brew install glide`

Set my GOPATH
`GOPATH=/Users/username/dev/go/`

Then the steps you recommended above:
```
git clone http://github.com/jgsqware/clairctl $GOPATH/src/github.com/jgsqware/clairctl
[...TRUNCATED...]
username@macbook $ cd $GOPATH/src/github.com/jgsqware/clairctl
username@macbook $ glide install -v
[INFO] 	Downloading dependencies. Please wait...
[INFO] 	--> Fetching github.com/fsouza/go-dockerclient.
[...TRUNCATED...]
[INFO] 	--> Fetching gopkg.in/yaml.v2.
[INFO] 	Setting references.
[INFO] 	--> Setting version for github.com/coreos/pkg to 2c77715c4df99b5420ffcae14ead08f52104065d.
[...TRUNCATED...]
[INFO] 	--> Setting version for gopkg.in/yaml.v2 to f7716cbe52baa25d2e9b0d0da546fcf909fc16b4.
[INFO] 	Exporting resolved dependencies...
[INFO] 	--> Exporting github.com/artyom/untar
[...TRUNCATED...]
[INFO] 	--> Exporting gopkg.in/yaml.v2
[INFO] 	Replacing existing vendor dependencies
[INFO] 	Removing nested vendor and Godeps/_workspace directories...
[INFO] 	Removing: /Users/username/dev/go/src/github.com/jgsqware/clairctl/vendor/github.com/coreos/clair/contrib/analyze-local-images/vendor
[INFO] 	Removing: /Users/username/dev/go/src/github.com/jgsqware/clairctl/vendor/github.com/coreos/clair/vendor
[INFO] 	Removing: /Users/username/dev/go/src/github.com/jgsqware/clairctl/vendor/github.com/docker/distribution/vendor
[INFO] 	Removing: /Users/username/dev/go/src/github.com/jgsqware/clairctl/vendor/github.com/docker/docker/vendor
[INFO] 	Removing: /Users/username/dev/go/src/github.com/jgsqware/clairctl/vendor/github.com/opencontainers/runc/Godeps/_workspace/src/github.com/docker/docker/vendor
[INFO] 	Removing: /Users/username/dev/go/src/github.com/jgsqware/clairctl/vendor/github.com/opencontainers/runc/Godeps/_workspace
[INFO] 	Removing Godep rewrites for /Users/username/dev/go/src/github.com/jgsqware/clairctl/vendor/github.com/opencontainers/runc
username@macbook $ go build
username@macbook $ GOOS=linux GOARCH=386 CGO_ENABLED=0 go build -o clairctl.linux.i386
username@macbook $ GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o clairctl.linux.amd64
```

# Command

```
Analyze your docker image with Clair, directly from your registry.

Usage:
  clairctl [command]

Available Commands:
  analyze     Analyze Docker image
  health      Get Health of clairctl and underlying services
  login       Log in to a Docker registry
  logout      Log out from a Docker registry
  pull        Pull Docker image information
  push        Push Docker image to Clair
  report      Generate Docker Image vulnerabilities report
  version     Get Versions of clairctl and underlying services

Flags:
      --config string      config file (default is ./.clairctl.yml)
      --log-level string   log level [Panic,Fatal,Error,Warn,Info,Debug]

Use "clairctl [command] --help" for more information about a command.
```

# Optional Configuration

```yaml
clair:
  port: 6060
  healthPort: 6061
  uri: http://clair
  report:
    path: ./reports
    format: html
```

# Building the latest binaries

**clairctl** requires Go 1.8+.

Install Glide:
```
curl https://glide.sh/get | sh
```

Clone and build:
```
git clone git@github.com:jgsqware/clairctl.git $GOPATH/src/github.com/jgsqware/clairctl
cd $GOPATH/src/github.com/jgsqware/clairctl
glide install -v
go generate ./clair
go build
```

This will result in a `clairctl` executable in the `$GOPATH/src/github.com/jgsqware/clairctl` folder.

# Contribution and Test

Go to /contrib folder
