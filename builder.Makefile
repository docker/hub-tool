include vars.mk

NULL:=/dev/null

ifeq ($(COMMIT),)
    COMMIT:=$(shell git rev-parse HEAD 2> $(NULL))
endif
ifeq ($(TAG_NAME),)
    TAG_NAME:=$(shell git describe --tags --match "v[0-9]*" 2> $(NULL))
endif

PKG_NAME:=github.com/docker/hub-tool
STATIC_FLAGS:=CGO_ENABLED=0
LDFLAGS:="-s -w \
    -X $(PKG_NAME)/internal.GitCommit=$(COMMIT) \
    -X $(PKG_NAME)/internal.Version=$(TAG_NAME)"
GO_BUILD:=go build -trimpath -ldflags=$(LDFLAGS)
VARS:=BINARY_NAME=${BINARY_NAME} \
    BINARY=${BINARY}

ifneq ($(strip $(E2E_TEST_NAME)),)
    RUN_TEST=-test.run $(E2E_TEST_NAME)
endif

TAR_TRANSFORM:=--transform s/packaging/${BINARY_NAME}/ --transform s/bin/${BINARY_NAME}/ --transform s/${PLATFORM_BINARY}/${BINARY_NAME}/
ifneq ($(findstring bsd,$(shell tar --version)),)
    TAR_TRANSFORM=-s /packaging/${BINARY_NAME}/ -s /bin/${BINARY_NAME}/ -s /${PLATFORM_BINARY}/${BINARY_NAME}/
endif
TMPDIR_WIN_PKG:=$(shell mktemp -d)

.PHONY: lint
lint:
	golangci-lint run --timeout 10m0s ./...

.PHONY: e2e
e2e:
	# TODO: gotestsum doesn't forward ldflags to go test with golang 1.15.0, so moving back to go test temporarily
	$(VARS) go test ./e2e $(RUN_TEST) -ldflags=$(LDFLAGS)

.PHONY: test-unit
test-unit:
	$(STATIC_FLAGS) gotestsum $(shell go list ./... | grep -vE '/e2e')

cross:
	GOOS=linux   GOARCH=amd64 $(STATIC_FLAGS) $(GO_BUILD) -o bin/$(BINARY_NAME)_linux_amd64 ./cmd/$(BINARY_NAME)
	GOOS=linux   GOARCH=arm64 $(STATIC_FLAGS) $(GO_BUILD) -o bin/$(BINARY_NAME)_linux_arm64 ./cmd/$(BINARY_NAME)
	GOOS=darwin  GOARCH=amd64 $(STATIC_FLAGS) $(GO_BUILD) -o bin/$(BINARY_NAME)_darwin_amd64 ./cmd/$(BINARY_NAME)
	GOOS=windows GOARCH=amd64 $(STATIC_FLAGS) $(GO_BUILD) -o bin/$(BINARY_NAME)_windows_amd64.exe ./cmd/$(BINARY_NAME)

# Note we're building statically for now to simplify releases. We can
# investigate dynamic builds later.
.PHONY: build
build:
	mkdir -p bin
	$(STATIC_FLAGS) $(GO_BUILD) -o bin/$(PLATFORM_BINARY) ./cmd/$(BINARY_NAME)
	cp bin/$(PLATFORM_BINARY) bin/$(BINARY)

.PHONY: package
package: build
	mkdir -p dist
ifeq ($(GOOS),windows)
	cp bin/$(PLATFORM_BINARY) $(TMPDIR_WIN_PKG)/$(BINARY_NAME).exe
	cp packaging/LICENSE $(TMPDIR_WIN_PKG)/LICENSE
	rm -f dist/$(BINARY_NAME)-windows-$(GOARCH).zip && 7z a dist/$(BINARY_NAME)-windows-$(GOARCH).zip $(TMPDIR_WIN_PKG)/*
	rm -r $(TMPDIR_WIN_PKG)
else
	tar -czf dist/$(BINARY_NAME)-$(GOOS)-$(GOARCH).tar.gz $(TAR_TRANSFORM) packaging/LICENSE bin/$(PLATFORM_BINARY)
endif

.PHONY: ci-extract
ci-extract:
	mkdir -p bin
ifeq ($(GOOS),windows)
	7z e -obin/ dist/$(BINARY_NAME)-windows-amd64.zip $(BINARY_NAME).exe
else
	tar xzf dist/$(BINARY_NAME)-$(GOOS)-$(GOARCH).tar.gz --strip-components 1 --directory bin/ $(BINARY_NAME)/$(BINARY_NAME)
endif
