include vars.mk

NULL := /dev/null

ifeq ($(COMMIT),)
  COMMIT := $(shell git rev-parse --short HEAD 2> $(NULL))
endif

ifeq ($(TAG_NAME),)
  TAG_NAME := $(shell git describe --always --dirty --abbrev=10 2> $(NULL))
endif

PKG_NAME=github.com/docker/hub-tool
STATIC_FLAGS= CGO_ENABLED=0
LDFLAGS := "-s -w \
  -X $(PKG_NAME)/internal.GitCommit=$(COMMIT) \
  -X $(PKG_NAME)/internal.Version=$(TAG_NAME)"
GO_BUILD = $(STATIC_FLAGS) go build -trimpath -ldflags=$(LDFLAGS)
VARS:= BINARY_NAME=${BINARY_NAME} \
       BINARY=${BINARY}

ifneq ($(strip $(E2E_TEST_NAME)),)
	RUN_TEST=-test.run $(E2E_TEST_NAME)
endif

.PHONY: lint
lint:
	golangci-lint run --timeout 10m0s ./...

.PHONY: e2e
e2e:
	# TODO: gotestsum doesn't forward ldflags to go test with golang 1.15.0, so moving back to go test temporarily
	$(VARS) go test ./e2e $(RUN_TEST) -ldflags=$(LDFLAGS)

.PHONY: test-unit
test-unit:
	gotestsum $(shell go list ./... | grep -vE '/e2e')

cross:
	GOOS=linux   GOARCH=amd64 $(GO_BUILD) -o dist/$(BINARY_NAME)_linux_amd64 ./cmd/$(BINARY_NAME)
	GOOS=darwin  GOARCH=amd64 $(GO_BUILD) -o dist/$(BINARY_NAME)_darwin_amd64 ./cmd/$(BINARY_NAME)
	GOOS=windows GOARCH=amd64 $(GO_BUILD) -o dist/$(BINARY_NAME)_windows_amd64.exe ./cmd/$(BINARY_NAME)

.PHONY: build
build:
	mkdir -p bin
	$(GO_BUILD) -o bin/$(PLATFORM_BINARY) ./cmd/$(BINARY_NAME)
	cp bin/$(PLATFORM_BINARY) bin/$(BINARY)

# For multi-platform (windows,macos,linux) github actions
.PHONY: download
download:
	GO111MODULE=on go get gotest.tools/gotestsum@v${GOTESTSUM_VERSION}
