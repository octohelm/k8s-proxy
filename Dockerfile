FROM golang:1.15-buster as builder

ARG GOPROXY=https://proxy.golang.org,direct
ENV GOBIN=/go/bin

WORKDIR /go/src
COPY ./ ./

RUN make build

FROM debian:buster-slim
COPY --from=builder /go/bin/k8s-proxy /go/bin/k8s-proxy
EXPOSE 80
ENTRYPOINT ["/go/bin/k8s-proxy"]