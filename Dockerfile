FROM alpine:3.5

ENV GOPATH=/go
ENV PATH=${GOPATH}/bin:${PATH}
ENV DOCKER_VERSION=17.05.0-ce
ENV DOCKER_API_VERSION=1.24
ENV CLAIRCTL_VERSION=${DOCKER_TAG:-master}
ARG CLAIRCTL_COMMIT=master

WORKDIR /root

RUN apk add --update curl \
 && apk add --virtual build-dependencies go gcc build-base glide git \
 && adduser clairctl -D \
 && addgroup docker -S -g 50 \
 && adduser clairctl docker \
 && curl https://get.docker.com/builds/Linux/x86_64/docker-${DOCKER_VERSION}.tgz -o docker.tgz \
 && tar xfvz docker.tgz --strip 1 -C /usr/bin/ docker/docker \
 && rm -f docker.tgz \
 && go get -u github.com/jteeuwen/go-bindata/... \
 && curl -sL https://github.com/jgsqware/clairctl/archive/${CLAIRCTL_VERSION}.zip -o clairctl.zip \
 && mkdir -p ${GOPATH}/src/github.com/jgsqware/ \
 && unzip clairctl.zip -d ${GOPATH}/src/github.com/jgsqware/ \
 && rm -f clairctl.zip \
 && mv ${GOPATH}/src/github.com/jgsqware/clairctl-${CLAIRCTL_COMMIT}* ${GOPATH}/src/github.com/jgsqware/clairctl \
 && cd ${GOPATH}/src/github.com/jgsqware/clairctl \
 && glide install -v \
 && go generate ./clair \
 && go build -o /usr/local/bin/clairctl -ldflags "-X github.com/jgsqware/clairctl/cmd.version=${CLAIRCTL_VERSION}" \
 && apk del build-dependencies \
 && rm -rf /var/cache/apk/* \
 && rm -rf /root/.glide/ \
 && rm -rf /go \
 && echo $'clair:\n\
  port: 6060\n\
  healthPort: 6061\n\
  uri: http://clair\n\
  priority: Low\n\
  report:\n\
    path: /reports\n\
    format: html\n\
  clairctl:\n\
    port: 44480\n\
    tempfolder: /tmp'\
    > /home/clairctl/clairctl.yml

EXPOSE 44480

USER clairctl
WORKDIR /home/clairctl/

VOLUME ["/tmp/", "/reports/"]
 
CMD ["/usr/sbin/crond", "-f"]