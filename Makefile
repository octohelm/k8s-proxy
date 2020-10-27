PKG = $(shell cat go.mod | grep "^module " | sed -e "s/module //g")
VERSION = $(shell cat .version)
COMMIT_SHA ?= $(shell git describe --always)-devel
TAG ?= $(VERSION)

GOBUILD = CGO_ENABLED=0 go build -ldflags "-X ${PKG}/version.Version=${VERSION}+sha.${COMMIT_SHA}"

GOBIN ?= ./bin
GOOS ?= $(shell go env GOOS)
GOARCH ?= $(shell go env GOARCH)
NAME ?= k8s-proxy

up:
	go run cmd/$(NAME)/main.go

test:
	go test ./...

build:
	$(GOBUILD) -o $(GOBIN)/$(NAME)-$(GOOS)-$(GOARCH) ./cmd/$(NAME)/main.go

prepare:
	@echo ::set-output name=image::$(NAME):$(TAG)
	@echo ::set-output name=build_args::VERSION=$(VERSION)

build.dockerx:
	docker buildx build \
		--push \
		--build-arg=GOPROXY=${GOPROXY} \
		--platform=linux/amd64,linux/arm64 \
		-t octohelm/$(NAME):$(TAG) \
		-f hack/Dockerfile .

lint:
	husky hook pre-commit
	husky hook commit-msg

include Makefile.apply