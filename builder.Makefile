include vars.mk

NULL := /dev/null

ifeq ($(COMMIT),)
  COMMIT := $(shell git rev-parse --short HEAD 2> $(NULL))
endif

ifeq ($(TAG_NAME),)
  TAG_NAME := $(shell git describe --always --dirty --abbrev=10 2> $(NULL))
endif

PKG_NAME=github.com/docker/hub-cli-plugin
STATIC_FLAGS= CGO_ENABLED=0
LDFLAGS := "-s -w \
  -X $(PKG_NAME)/internal.GitCommit=$(COMMIT) \
  -X $(PKG_NAME)/internal.Version=$(TAG_NAME)"
GO_BUILD = $(STATIC_FLAGS) go build -trimpath -ldflags=$(LDFLAGS)

SNYK_DOWNLOAD_NAME:=snyk-linux
SNYK_BINARY:=snyk
PWD:=$(shell pwd)
ifeq ($(GOOS),windows)
	SNYK_DOWNLOAD_NAME:=snyk-win.exe
	SNYK_BINARY=snyk.exe
	PWD=$(shell dirname $(realpath $(firstword $(MAKEFILE_LIST))))
endif
ifeq ($(GOOS),darwin)
	SNYK_DOWNLOAD_NAME:=snyk-macos
endif

ifneq ($(strip $(E2E_TEST_NAME)),)
	RUN_TEST=-test.run $(E2E_TEST_NAME)
endif

VARS:= SNYK_DESKTOP_VERSION=${SNYK_DESKTOP_VERSION}\
	SNYK_USER_VERSION=${SNYK_USER_VERSION}\
	DOCKER_CONFIG=$(PWD)/docker-config\
	SNYK_OLD_PATH=$(PWD)/docker-config/snyk-old\
	SNYK_USER_PATH=$(PWD)/docker-config/snyk-user\
	SNYK_DESKTOP_PATH=$(PWD)/docker-config/snyk-desktop

.PHONY: lint
lint:
	golangci-lint run --timeout 10m0s ./...

.PHONY: e2e
e2e:
	mkdir -p docker-config/hub
	mkdir -p docker-config/cli-plugins
	cp ./bin/${PLATFORM_BINARY} docker-config/cli-plugins/${BINARY}
	# TODO: gotestsum doesn't forward ldflags to go test with golang 1.15.0, so moving back to go test temporarily
	$(VARS) go test ./e2e $(RUN_TEST) -ldflags=$(LDFLAGS)

.PHONY: test-unit
test-unit:
	gotestsum $(shell go list ./... | grep -vE '/e2e')

cross:
	GOOS=linux   GOARCH=amd64 $(GO_BUILD) -o dist/docker-hub_linux_amd64 ./cmd/docker-hub
	GOOS=darwin  GOARCH=amd64 $(GO_BUILD) -o dist/docker-hub_darwin_amd64 ./cmd/docker-hub
	GOOS=windows GOARCH=amd64 $(GO_BUILD) -o dist/docker-hub_windows_amd64.exe ./cmd/docker-hub

.PHONY: build
build:
	mkdir -p bin
	$(GO_BUILD) -o bin/$(PLATFORM_BINARY) ./cmd/docker-hub

# For multi-platform (windows,macos,linux) github actions
.PHONY: download
download:
	GO111MODULE=on go get gotest.tools/gotestsum@v${GOTESTSUM_VERSION}
