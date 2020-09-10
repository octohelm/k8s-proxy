PKG = $(shell cat go.mod | grep "^module " | sed -e "s/module //g")
VERSION = $(shell cat .version)
COMMIT_SHA ?= $(shell git describe --always)-devel

GOBUILD = CGO_ENABLED=0 go build -ldflags "-X ${PKG}/version.Version=${VERSION}+sha.${COMMIT_SHA}"

GOBIN ?= ./bin
GOOS ?= $(shell go env GOOS)
GOARCH ?= $(shell go env GOARCH)


up:
	go run cmd/k8s-proxy/main.go

test:
	go test ./...

build:
	$(GOBUILD) -o $(GOBIN)/k8s-proxy ./cmd/k8s-proxy/main.go

build.dockerx:
	docker buildx build \
		--push \
		--build-arg=GOPROXY=${GOPROXY} \
		--platform=linux/amd64,linux/arm64 \
		-t hub-dev.demo.querycap.com/octohelm/k8s-proxy:${VERSION} \
		-f Dockerfile .

lint:
	husky hook pre-commit
	husky hook commit-msg

release:
	git push
	git push origin v${VERSION}

include Makefile.apply