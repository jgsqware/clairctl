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

```bash
curl -L https://raw.githubusercontent.com/jgsqware/clairctl/master/install.sh | sh
``` 

# Docker-compose

```bash
$ git clone git@github.com:jgsqware/clairctl.git $GOPATH/src/github.com/jgsqware/clairctl
$ cd $GOPATH/src/github.com/jgsqware/clairctl
$ docker-compose up -d postgres
Creating network "clairctl_default" with the default driver
Creating clairctl_postgres_1 ...
Creating clairctl_clair_1 ...
Creating clairctl_clairctl_1 ...
```

The above commands will check out the `clairctl` repo and start the complete postgres/clair/clairctl stack.

```bash
$ docker-compose exec clairctl clairctl health

Clair: ✔
```

The above command will make sure clairctl can reach clair.

If you wish to serve local images to clair, the user inside the clairctl container will need read access to `/var/run/docker.sock`.

Give the user access by:
  - Running the container as root (`--user root` with `docker run` or `user: root` with `docker-compose`)
  - Add the container user to the docker group (`----group-add group_id` with `docker run` or `group_add: group_id` with `docker-compose`)

To get the group name or id, simply execute :

```bash
$ docker-compose exec clairctl ls -alh /var/run/docker.sock
srw-rw----    1 root     50             0 Jul 18 09:48 /var/run/docker.sock
```

In the example above, 50 is the required group.

# Usage

[![asciicast](https://asciinema.org/a/41461.png)](https://asciinema.org/a/41461)

# Reporting

**clairctl** get vulnerabilities report from Clair and generate HTML report

clairctl can be used for Docker Hub and self-hosted Registry

# Commands

```
Analyze your docker image with Clair, directly from your registry.

Usage:
  clairctl [command]

Available Commands:
  analyze     Analyze Docker image
  health      Get Health of clairctl and underlying services
  pull        Pull Docker image information (This will not pull the image !)
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

# FAQ

## I get 400 errors !

If you get 400 errors, check out clair's logs. The usual reasons are :
  
  - You are serving a local image, and clair cannot connect to clairctl.
  - You are trying to analyze an official image from docker hub and you have not done a docker login first.
  
Please try these two things before opening an Issue.

## I get access denied on /var/run/docker.sock

If you are running the stack with the provided `docker-compose.yml`, don't forget to grant the user from the clairctl container access to `/var/run/docker.sock`. 

All steps are detailed in the Docker-compose section above.

## I get a "client is newer than server" error

When working with local images, you can experience the following erro : 

```
Error response from daemon: client is newer than server (client API version: 1.26, server API version: 1.23)
```

Don't panic, this is due to your Docker daemon running an API version older than the one from the Docker client API bundled with clairctl.

First, check out your docker version:

```bash
$ docker version

Client:
 Version:      1.11.2
 API version:  1.23
 Go version:   go1.5.4
 Git commit:   b9f10c9
 Built:        Wed Jun  1 21:20:08 2016
 OS/Arch:      linux/amd64

Server:
 Version:      1.11.2
 API version:  1.23
 Go version:   go1.7.3
 Git commit:   4dc5990
 Built:        
 OS/Arch:      linux/amd64
```

Simply run clairctl with the corresponding `DOCKER_API_VERSION` environment variable to enable compatibility with your Docker API version.

```bash
$ DOCKER_API_VERSION=1.23 clairctl push -l mynamespace/myimage:mytag
```

Bear in mind that this might not work with very old versions of Docker daemon.

# Contribution and Test

Go to /contrib folder
