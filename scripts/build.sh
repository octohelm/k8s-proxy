#!/bin/sh

go build -v -ldflags "-X ${PKG}/version.Version=${VERSION}" -o "${APP}"