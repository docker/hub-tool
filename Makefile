#   Copyright 2020 Docker Hub Tool authors

#   Licensed under the Apache License, Version 2.0 (the "License");
#   you may not use this file except in compliance with the License.
#   You may obtain a copy of the License at

#       http://www.apache.org/licenses/LICENSE-2.0

#   Unless required by applicable law or agreed to in writing, software
#   distributed under the License is distributed on an "AS IS" BASIS,
#   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
#   See the License for the specific language governing permissions and
#   limitations under the License.
include vars.mk
export DOCKER_BUILDKIT=1

BUILD_ARGS:=--build-arg GO_VERSION=$(GO_VERSION) \
    --build-arg ALPINE_VERSION=$(ALPINE_VERSION) \
    --build-arg GOLANGCI_LINT_VERSION=$(GOLANGCI_LINT_VERSION) \
    --build-arg TAG_NAME=$(TAG_NAME) \
    --build-arg GOTESTSUM_VERSION=$(GOTESTSUM_VERSION) \
    --build-arg BINARY_NAME=$(BINARY_NAME) \
    --build-arg BINARY=$(BINARY)

E2E_ENV:=--env E2E_HUB_USERNAME \
    --env E2E_HUB_TOKEN \
    --env E2E_TEST_NAME

TMPDIR_WIN_PKG:=$(shell mktemp -d)

.PHONY: all
all: build

.PHONY: build
build: ## Build the tool in a container
	docker build $(BUILD_ARGS) . \
		--output type=local,dest=./bin \
		--platform local \
		--target hub

.PHONY: cross
cross: ## Cross compile the tool binaries in a container
	docker build $(BUILD_ARGS) . \
		--output type=local,dest=./bin \
		--target cross

.PHONY: package-cross
package-cross: cross ## Package the cross compiled binaries in tarballs for *nix and a zip for Windows
	mkdir -p dist
	docker build $(BUILD_ARGS) . \
		--platform linux/amd64 \
		--output type=tar,dest=- \
		--target package | gzip -9 > dist/$(BINARY_NAME)-linux-amd64.tar.gz
	docker build $(BUILD_ARGS) . \
		--platform linux/arm64 \
		--output type=tar,dest=- \
		--target package | gzip -9 > dist/$(BINARY_NAME)-linux-arm64.tar.gz
	docker build $(BUILD_ARGS) . \
		--platform linux/arm \
		--output type=tar,dest=- \
		--target package | gzip -9 > dist/$(BINARY_NAME)-linux-arm.tar.gz
	docker build $(BUILD_ARGS) . \
		--platform darwin/amd64 \
		--output type=tar,dest=- \
		--target package | gzip -9 > dist/$(BINARY_NAME)-darwin-amd64.tar.gz
	cp bin/$(BINARY_NAME)_windows_amd64.exe $(TMPDIR_WIN_PKG)/$(BINARY_NAME).exe
	rm -f dist/$(BINARY_NAME)-windows-amd64.zip && zip dist/$(BINARY_NAME)-windows-amd64.zip -j packaging/LICENSE $(TMPDIR_WIN_PKG)/$(BINARY_NAME).exe
	rm -r $(TMPDIR_WIN_PKG)

.PHONY: install
install: build ## Install the tool to your /usr/local/bin/
	cp bin/$(BINARY_NAME) /usr/local/bin/$(BINARY)

.PHONY: test ## Run unit tests then end-to-end tests
test: test-unit e2e

.PHONY: e2e-build
e2e-build:
	docker build $(BUILD_ARGS) . --target e2e -t $(BINARY_NAME):e2e

.PHONY: e2e
e2e: e2e-build ## Run the end-to-end tests
	docker run $(E2E_ENV) --rm \
		-v /var/run/docker.sock:/var/run/docker.sock \
		-v $(shell go env GOCACHE):/root/.cache/go-build \
		-v $(shell go env GOMODCACHE):/go/pkg/mod \
		$(BINARY_NAME):e2e

test-unit-build:
	docker build $(BUILD_ARGS) . --target test-unit -t $(BINARY_NAME):test-unit

.PHONY: test-unit
test-unit: test-unit-build ## Run unit tests
	docker run --rm \
		-v $(shell go env GOCACHE):/root/.cache/go-build \
		-v $(shell go env GOMODCACHE):/go/pkg/mod \
		$(BINARY_NAME):test-unit

.PHONY: lint
lint: ## Run the go linter
	@docker build . --target lint

.PHONY: validate-headers
validate-headers: ## Validate files license header
	docker run --rm -v $(CURDIR):/work -w /work \
		golang:${GO_VERSION} \
		bash -c 'go get -u github.com/kunalkushwaha/ltag && ./scripts/validate/fileheader'

.PHONY: validate-go-mod
validate-go-mod: ## Validate go.mod and go.sum are up-to-date
	@docker build . --target check-go-mod

.PHONY: validate
validate: validate-go-mod validate-headers ## Validate sources

.PHONY: help
help: ## Show help
	@echo Please specify a build target. The choices are:
	@grep -E '^[0-9a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":| ## "}; {printf "\033[36m%-30s\033[0m %s\n", $$2, $$NF}'
