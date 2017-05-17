# clairctl
[![Build Status](https://travis-ci.org/jgsqware/clairctl.svg?branch=master)](https://travis-ci.org/jgsqware/clairctl)

> Tracking container vulnerabilities with Clair Control

Clairctl is a lightweight command-line tool doing the bridge between Registries as Docker Hub, Docker Registry or Quay.io, and the CoreOS vulnerability tracker, Clair.
Clairctl will play as reverse proxy for authentication.

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

**clairctl** requires at least Go 1.7.

Install Glide:
```
curl https://glide.sh/get | sh
```

Clone and build:
```
git clone github.com/jgsqware/clairctl  $GOPATH/src/github.com/jgsqware/clairctl
cd $GOPATH/src/github.com/jgsqware/clairctl
glide install -v
go build
```

This will result in a `clairctl` executable in the `$GOPATH/src/github.com/jgsqware/clairctl` folder.

# Contribution and Test

Go to /contrib folder
