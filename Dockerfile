FROM golang:1.14-alpine as builder

ARG GOPROXY=https://proxy.golang.org,direct
ARG PKG=github.com/octohelm/k8s-proxy
ARG VERSION=0.0.1

COPY ./ /go/src/${PKG}

WORKDIR /go/src/${PKG}/cmd/k8s-proxy

RUN ../../scripts/build.sh && cp k8s-proxy /go/bin/

FROM alpine
COPY --from=builder /go/bin/k8s-proxy /go/bin/k8s-proxy
EXPOSE 80
ENTRYPOINT ["/go/bin/k8s-proxy"]