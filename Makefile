MODULE=github.com/oars-sigs/oars-cloud
VERSION_PKG=$(MODULE)/pkg/version
BINARY=bin/oars
GO_VERSION=$(shell go version)
BUILD_TIME=$(shell date +%FT%T%z)
GIT_COMMIT=$(shell git rev-parse HEAD)
APP_VERSION=$(shell git rev-parse --abbrev-ref HEAD | grep -v HEAD || git describe --exact-match HEAD 2> /dev/null || git rev-parse HEAD)

LDFLAGS=-ldflags "-w -s -extldflags '-static' -X '$(VERSION_PKG).AppVersion=${APP_VERSION}' -X '$(VERSION_PKG).BuildTime=${BUILD_TIME}' -X '$(VERSION_PKG).GoVersion=${GO_VERSION}' -X '$(VERSION_PKG).GitCommit=${GIT_COMMIT}'"

.PHONY: bin
all: bin

bin:
	go build  $(LDFLAGS) -tags static -o $(BINARY) cmd/oars/main.go



