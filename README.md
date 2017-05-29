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
