SHELL:=$(PREFIX)/bin/bash

VERSION=$(shell git describe --always)
MODULE=github.com/Gameye/igniter-shell-go

GO_SRC=*.go */*.go
BIN=\
	amd64/linux/igniter-shell \
	amd64/darwin/igniter-shell \

BIN_TARGET=$(addprefix bin/,${BIN})
PACKAGE_TARGET=$(patsubst %,out/%-${VERSION}.tar.gz,${BIN})

all: build

rebuild: clean build

clean:
	rm -rf bin out

build: ${BIN_TARGET} ${PACKAGE_TARGET}

bin/%/linux/igniter-shell: export GOOS=linux
bin/%/darwin/igniter-shell: export GOOS=darwin
bin/amd64/%/igniter-shell: export GOARCH=amd64
bin/%: $(GO_SRC)
	go build \
		-o $@ \
		-ldflags=" \
			-X ${MODULE}/resource.Version=${VERSION} \
			-extldflags '-static' \
		"

out/%-${VERSION}.tar.gz: bin/%
	@mkdir --parents $(@D)
	tar --create --gzip --file $@ --directory $(<D) $(<F)

.PHONY: all clean build rebuild
