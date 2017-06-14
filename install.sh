#!/bin/bash -ue
PLATFORM=$(uname -s | tr '[:upper:]' '[:lower:]')
ARCH=$(uname -m)
if [[ "$(uname -m)" == *"64"* ]]; then ARCH="amd64"; else ARCH="386"; fi
echo $PLATFORM
echo $ARCH

curl -o /usr/local/bin/clairctl "https://s3.amazonaws.com/clairctl/latest/clairctl-$PLATFORM-$ARCH"
chmod +x /usr/local/bin/clairctl