# syntax=docker/dockerfile:experimental


#   Copyright 2020 Docker Inc.

#   Licensed under the Apache License, Version 2.0 (the "License");
#   you may not use this file except in compliance with the License.
#   You may obtain a copy of the License at

#       http://www.apache.org/licenses/LICENSE-2.0

#   Unless required by applicable law or agreed to in writing, software
#   distributed under the License is distributed on an "AS IS" BASIS,
#   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
#   See the License for the specific language governing permissions and
#   limitations under the License.


ARG GO_VERSION=1.15.0
ARG CLI_VERSION=19.03.9
ARG ALPINE_VERSION=3.12.0
ARG GOLANGCI_LINT_VERSION=v1.27.0-alpine

####
# BUILDER
####
FROM --platform=${BUILDPLATFORM} golang:${GO_VERSION} AS builder
WORKDIR /go/src/github.com/docker/hub-cli-plugin

# cache go vendoring
COPY go.* ./
RUN --mount=type=cache,target=/go/pkg \
    go mod download -x
COPY . .

####
# LINT-BASE
####
FROM golangci/golangci-lint:${GOLANGCI_LINT_VERSION} AS lint-base

####
# LINT
####
FROM builder AS lint
ENV CGO_ENABLED=0
COPY --from=lint-base /usr/bin/golangci-lint /usr/bin/golangci-lint
RUN --mount=type=cache,target=/root/.cache/go-build \
    --mount=type=cache,target=/go/pkg \
    --mount=type=cache,target=/root/.cache/golangci-lint \
    make -f builder.Makefile lint

####
# CHECK GO MOD
####
FROM builder AS check-go-mod
RUN scripts/validate/check-go-mod

####
# BUILD
####
FROM builder AS build
ARG TARGETOS
ARG TARGETARCH
RUN --mount=type=cache,target=/root/.cache/go-build \
    --mount=type=cache,target=/go/pkg \
    GOOS=${TARGETOS} \
    GOARCH=${TARGETARCH} \
    make -f builder.Makefile build

####
# HUB
####
FROM scratch AS hub
COPY --from=build /go/src/github.com/docker/hub-cli-plugin/bin/hub-tool_* /

####
# CROSS_BUILD
####
FROM builder AS cross-build
ARG TAG_NAME
ENV TAG_NAME=$TAG_NAME
RUN --mount=type=cache,target=/root/.cache/go-build \
    --mount=type=cache,target=/go/pkg \
    make -f builder.Makefile cross

####
# CROSS
####
FROM scratch AS cross
COPY --from=cross-build /go/src/github.com/docker/hub-cli-plugin/dist /

####
# GOTESTSUM
####
FROM alpine:${ALPINE_VERSION} AS gotestsum
ARG GOTESTSUM_VERSION=0.5.2

RUN apk add -U --no-cache wget tar
# install gotestsum
WORKDIR /root
RUN wget https://github.com/gotestyourself/gotestsum/releases/download/v${GOTESTSUM_VERSION}/gotestsum_${GOTESTSUM_VERSION}_linux_amd64.tar.gz -nv -O - | tar -xz

####
# TEST-UNIT
####
FROM builder AS test-unit
ARG TAG_NAME
ENV TAG_NAME=$TAG_NAME

COPY --from=gotestsum /root/gotestsum /usr/local/bin/gotestsum
CMD ["make", "-f", "builder.Makefile", "test-unit"]

####
# DOWNLOAD
####
FROM golang:${GO_VERSION} AS download
COPY builder.Makefile vars.mk ./
RUN make -f builder.Makefile download

####
# E2E
####
FROM builder AS e2e
ARG TARGETOS
ARG TARGETARCH
ARG TAG_NAME
ARG BINARY_NAME
ARG BINARY
ENV TAG_NAME=$TAG_NAME
ENV BINARY_NAME=$BINARY_NAME
ENV DOCKER_CONFIG="/root/.docker"

# install hub tool
COPY --from=cross-build /go/src/github.com/docker/hub-cli-plugin/dist/${BINARY_NAME}_${TARGETOS}_${TARGETARCH} /go/src/github.com/docker/hub-cli-plugin/bin/${BINARY}
RUN chmod +x ./bin/${BINARY}
CMD ["make", "-f", "builder.Makefile", "e2e"]