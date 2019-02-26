SHELL:=$(PREFIX)/bin/bash

VERSION=`git describe --always`
MODULE=github.com/elmerbulthuis/shell-go

GO_SRC=*.go */*.go
BIN=\
	gameye-shell \

BIN_TARGET=$(addprefix bin/,${BIN})

all: rebuild

rebuild: clean build

clean:
	rm -rf bin

build: ${BIN_TARGET}

${BIN_TARGET}: $(GO_SRC)
	go build \
		-o $@ \
		-ldflags=" \
			-X ${MODULE}/resource.Version=${VERSION} -extldflags '-static' \
		"

.PHONY: all clean build rebuild di
