SHELL:=$(PREFIX)/bin/bash

VERSION=$(shell git describe --always)
MODULE=github.com/Gameye/igniter-shell-go

GO_SRC=*.go */*.go
BIN=\
	igniter-shell-linux-amd64 \
	igniter-shell-darwin-amd64 \

BIN_TARGET=$(addprefix bin/,${BIN})

all: rebuild

rebuild: clean build

clean:
	rm -rf bin

build: ${BIN_TARGET} ${PACKAGE_TARGET}

bin/igniter-shell-linux-%: export GOOS=linux
bin/igniter-shell-darwin-%: export GOOS=darwin
bin/igniter-shell-%-amd64: export GOARCH=amd64
bin/igniter-shell-%: $(GO_SRC)
	go build \
		-o $@ \
		-ldflags=" \
			-X ${MODULE}/resource.Version=${VERSION} \
			-extldflags '-static' \
		"

.PHONY: all clean build rebuild
