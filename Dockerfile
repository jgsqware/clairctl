FROM ubuntu:16.04

LABEL maintainer "jgsqware"

ENV DEBIAN_FRONTEND=noninteractive
ENV GOPATH /go
ENV PATH $GOPATH/bin:/usr/local/go/bin:$GOPATH/src/github.com/jgsqware/clairctl:$PATH
ENV GLIDE_VERSION 0.12.3
ENV GLIDE_DOWNLOAD_URL https://github.com/Masterminds/glide/releases/download/v${GLIDE_VERSION}/glide-v${GLIDE_VERSION}-linux-amd64.tar.gz

RUN apt update -q \
 && apt -y install software-properties-common \
 && add-apt-repository ppa:longsleep/golang-backports \
 && apt update -q \
 && apt -y -q install --no-install-recommends --fix-missing \
 	  ca-certificates \
      git \
      curl \
      build-essential \
      golang-go \
 && curl -fsSL "$GLIDE_DOWNLOAD_URL" -o glide.tar.gz \
 && tar -xzf glide.tar.gz \
 && mv linux-amd64/glide /usr/bin/ \
 && rm -r linux-amd64 \
 && rm glide.tar.gz \
 && go get -u github.com/jteeuwen/go-bindata/... \
 && git clone https://github.com/jgsqware/clairctl.git  $GOPATH/src/github.com/jgsqware/clairctl \
 && cd $GOPATH/src/github.com/jgsqware/clairctl \
 && glide install -v \
 && go generate ./clair \
 && go build \
 && apt-get -y remove software-properties-common \
 && apt-get -y autoremove --purge \
 && apt-get -y clean \
 && rm -rf /var/lib/apt/lists/* \
 && clairctl version
