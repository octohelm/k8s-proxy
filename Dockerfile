FROM --platform=${BUILDPLATFORM} golang:1.15-buster as builder

ARG GOPROXY=https://proxy.golang.org,direct
ENV GOBIN=/go/bin

WORKDIR /go/src
COPY ./ ./

ARG TARGETARCH
RUN GOARCH=${TARGETARCH} make build

FROM ghcr.io/querycap/distroless/static-debian10:latest

ARG TARGETARCH
COPY --from=builder /go/bin/k8s-proxy-${TARGETARCH} /go/bin/k8s-proxy
EXPOSE 80
ENTRYPOINT ["/go/bin/k8s-proxy"]