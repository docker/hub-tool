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

# Pinned Versions
GO_VERSION=1.18-alpine
CLI_VERSION=20.10.12
ALPINE_VERSION=3.15.4
GOLANGCI_LINT_VERSION=v1.45.2-alpine
GOTESTSUM_VERSION=1.8.0

GOOS?=$(shell go env GOOS)
GOARCH?=$(shell go env GOARCH)
BINARY_EXT=
ifeq ($(GOOS),windows)
    BINARY_EXT=.exe
endif
BINARY_NAME=hub-tool
PLATFORM_BINARY:=$(BINARY_NAME)_$(GOOS)_$(GOARCH)$(BINARY_EXT)
BINARY=$(BINARY_NAME)$(BINARY_EXT)
