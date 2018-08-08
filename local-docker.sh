#!/usr/bin/env sh

# Purpose: This script allows building the clairctl from the local source files. Not downloading the source from github as a tar as done in Dockerfile


TAG_VERSION=$1

if test -n "$TAG_VERSION"; then
    TAG_VERSION="-t ${TAG_VERSION}";
fi

# Zip up all of the Go source files that are needed for the Dockerfile
# This clairctl.zip will be used inside the Dockerfile to create the binary of clairctl

rm clairctl.zip
zip -r clairctl.zip cmd clair config contrib docker docker-compose-data hooks server xstrings DockerFile glide* main.go VERSION LICENSE -x *.idea* -x vendor -x clairctl.zip -x clairctl -x docker.tgz

docker build . -f LocalDockerfile ${TAG_VERSION}