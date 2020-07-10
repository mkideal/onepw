PKG=github.com/mkideal/pkg/build
TAGS=$(shell git rev-list --tags --max-count=1)
VERSION=$(shell git describe --tags ${TAGS})
BRANCH=$(shell git symbolic-ref --short HEAD)
COMMIT=$(shell git rev-parse HEAD)
DATE=$(shell date "+%Y/%m/%d")
TIME=$(shell date "+%H:%M:%S")

build:
	go build -ldflags "-X ${PKG}.version=${VERSION} -X ${PKG}.branch=${BRANCH} -X ${PKG}.commit=${COMMIT} -X ${PKG}.date=${DATE} -X ${PKG}.time=${TIME}"

install:
	go install -ldflags "-X ${PKG}.version=${VERSION} -X ${PKG}.branch=${BRANCH} -X ${PKG}.commit=${COMMIT} -X ${PKG}.date=${DATE} -X ${PKG}.time=${TIME}"
