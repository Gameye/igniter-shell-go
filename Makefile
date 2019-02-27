SHELL:=$(PREFIX)/bin/bash

VERSION=$(shell git describe --always)
MODULE=github.com/Gameye/igniter-shell-go

GO_SRC=*.go */*.go
BIN=\
	igniter-shell-linux-amd64 \
	igniter-shell-darwin-amd64 \

BIN_TARGET=$(addprefix bin/,${BIN})

PACKAGE=$(addsuffix .deb,$(notdir $(wildcard package/deb/*)))
PACKAGE_TARGET=$(addprefix out/,${PACKAGE})

all: rebuild

rebuild: clean build

clean:
	rm -rf bin out .package_tmp

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

.package_tmp/deb/%: bin/igniter-shell-linux package/deb/%
	@mkdir -p $@
	cp -r package/deb/$*/* $@

	@mkdir -p $@/usr/local/bin
	cp $< $@/usr/local/bin/igniter-shell

	sed -i 's/Version:.*/Version: '$(VERSION:v%=%)'/' $@/DEBIAN/control

out/%.deb:	.package_tmp/deb/%
	@mkdir -p $(@D)
	dpkg-deb --build $< $@

.PHONY: all clean build rebuild

.PRECIOUS: .package_tmp/deb/% bin/igniter-shell-%
