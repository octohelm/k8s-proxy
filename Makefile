PKG=$(shell cat go.mod | grep module | sed -E s/module\ //g)
VERSION=$(shell cat .version)
APP=k8s-proxy

up:
	go run cmd/k8s-proxy/main.go

test:
	go test ./...

build:
	cd cmd/k8s-proxy && ../../scripts/build.sh

dockerx:
	docker buildx build \
		--push \
		--build-arg=GOPROXY=${GOPROXY} \
		--build-arg=PKG=${PKG} \
		--build-arg=VERSION=${VERSION} \
		--platform=linux/amd64,linux/arm64 \
		-f Dockerfile \
		-t octohelm/k8s-proxy:${VERSION} .

lint:
	husky hook pre-commit
	husky hook commit-msg

include Makefile.apply