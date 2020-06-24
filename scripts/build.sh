#!/bin/sh

CGO_ENABLED=0 go build -v -ldflags "-X ${PKG}/version.Version=${VERSION}" -o "${APP}"