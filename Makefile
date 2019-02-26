SHELL:=$(PREFIX)/bin/bash

VERSION=`git describe --always`
MODULE=github.com/elmerbulthuis/shell-go

GO_SRC=*.go */*.go
BIN=\
	gameye-shell-linux-amd64 \
	gameye-shell-darwin-amd64 \

BIN_TARGET=$(addprefix bin/,${BIN})

all: rebuild

rebuild: clean build

clean:
	rm -rf bin

build: ${BIN_TARGET}

bin/gameye-shell-linux-amd64: export GOOS=linux
bin/gameye-shell-linux-amd64: export GOARCH=amd64
bin/gameye-shell-darwin-amd64: export GOOS=darwin
bin/gameye-shell-darwin-amd64: export GOARCH=amd64
bin/gameye-shell-%: $(GO_SRC)
	go build \
		-o $@ \
		-ldflags=" \
			-X ${MODULE}/resource.Version=${VERSION} -extldflags '-static' \
		"

.PHONY: all clean build rebuild di
