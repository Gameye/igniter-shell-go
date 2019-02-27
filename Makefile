SHELL:=$(PREFIX)/bin/bash

VERSION=`git describe --always`
MODULE=github.com/elmerbulthuis/shell-go

GO_SRC=*.go */*.go
BIN=\
	gameye-shell-linux-amd64 \
	gameye-shell-darwin-amd64 \

BIN_TARGET=$(addprefix bin/,${BIN})

PACKAGE=$(addsuffix .deb,$(notdir $(wildcard package/deb/*)))
PACKAGE_TARGET=$(addprefix out/,${PACKAGE})

all: rebuild

rebuild: clean build

clean:
	rm -rf bin out .package_tmp

build: ${BIN_TARGET} ${PACKAGE_TARGET}

bin/gameye-shell-linux-%: export GOOS=linux
bin/gameye-shell-darwin-%: export GOOS=darwin
bin/gameye-shell-%-amd64: export GOARCH=amd64
bin/gameye-shell-%: $(GO_SRC)
	go build \
		-o $@ \
		-ldflags=" \
			-X ${MODULE}/resource.Version=${VERSION} \
			-extldflags '-static' \
		"

.package_tmp/deb/%: bin/gameye-shell-linux package/deb/%
	@mkdir -p $@
	cp -r package/deb/$*/* $@

	@mkdir -p $@/usr/local/bin
	cp $< $@/usr/local/bin/gameye-game-igniter

	sed -i 's/Version:.*/Version: '${VERSION}'/' $@/DEBIAN/control

out/%.deb:	.package_tmp/deb/%
	@mkdir -p $(@D)
	dpkg-deb --build $< $@

.PHONY: all clean build rebuild
