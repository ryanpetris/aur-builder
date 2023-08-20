export GO111MODULE=on
GOPROXY ?= https://proxy.golang.org,direct
export GOPROXY

BUILD_TAG = devel
ARCH ?= $(shell uname -m)
BIN := aur-builder
DESTDIR :=
GO ?= go
PKGNAME := aur-builder
PREFIX := /usr/local

MAJORVERSION := 1
MINORVERSION := 0
PATCHVERSION := 0
VERSION ?= ${MAJORVERSION}.${MINORVERSION}.${PATCHVERSION}

FLAGS ?= -trimpath -mod=readonly -modcacherw
EXTRA_FLAGS ?= -buildmode=pie
LDFLAGS := -linkmode=external

SOURCES ?= $(shell find . -name "*.go" -type f)

.PHONY: default
default: build

.PHONY: all
all: | clean build

.PHONY: clean
clean:
	$(GO) clean $(FLAGS) -i ./...

.PHONY: build
build: $(BIN)

$(BIN): $(SOURCES)
	$(GO) build $(FLAGS) -ldflags '$(LDFLAGS)' $(EXTRA_FLAGS) -o "$@"
