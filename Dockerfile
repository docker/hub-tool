# syntax=docker/dockerfile:experimental


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


ARG GO_VERSION=1.16.3-alpine
ARG CLI_VERSION=20.10.2
ARG ALPINE_VERSION=3.12.2
ARG GOLANGCI_LINT_VERSION=v1.33.0-alpine

####
# BUILDER
####
FROM --platform=${BUILDPLATFORM} golang:${GO_VERSION} AS builder
WORKDIR /go/src/github.com/docker/hub-tool
RUN apk add --no-cache \
    bash \
    git \
    make

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
COPY --from=lint-base /usr/bin/golangci-lint /usr/bin/golangci-lint
RUN --mount=type=cache,target=/root/.cache/go-build \
    --mount=type=cache,target=/go/pkg \
    --mount=type=cache,target=/root/.cache/golangci-lint \
    make -f builder.Makefile lint

####
# VALIDATE HEADERS
####
FROM builder AS validate-headers
RUN --mount=type=cache,target=/root/.cache/go-build \
    --mount=type=cache,target=/go/pkg \
    go get -u github.com/kunalkushwaha/ltag && ./scripts/validate/fileheader

####
# CHECK GO MOD
####
FROM builder AS check-go-mod
RUN scripts/validate/check-go-mod

####
# GO MOD TIDY
####
FROM builder as go-mod-tidy-run
RUN --mount=type=cache,target=/root/.cache/go-build \
    --mount=type=cache,target=/go/pkg \
    go mod tidy

FROM scratch AS go-mod-tidy
COPY --from=go-mod-tidy-run /go/src/github.com/docker/hub-tool/go.mod /
COPY --from=go-mod-tidy-run /go/src/github.com/docker/hub-tool/go.sum /

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
ARG BINARY_NAME
COPY --from=build /go/src/github.com/docker/hub-tool/bin/${BINARY_NAME} /${BINARY_NAME}

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
COPY --from=cross-build /go/src/github.com/docker/hub-tool/bin/* /

####
# PACKAGE
####
FROM scratch AS package
ARG BINARY_NAME
ARG TARGETOS
ARG TARGETARCH
COPY --from=builder /go/src/github.com/docker/hub-tool/packaging/LICENSE /${BINARY_NAME}/LICENSE
COPY --from=cross /${BINARY_NAME}_${TARGETOS}_${TARGETARCH} /${BINARY_NAME}/${BINARY_NAME}

####
# GOTESTSUM
####
FROM alpine:${ALPINE_VERSION} AS gotestsum
RUN apk add --no-cache \
    tar \
    wget
# install gotestsum
WORKDIR /root
ARG GOTESTSUM_VERSION=0.6.0
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
ARG TAG_NAME
ARG BINARY
ENV TAG_NAME=$TAG_NAME
ENV BINARY=$BINARY
ENV DOCKER_CONFIG="/root/.docker"
# install hub tool
COPY --from=build /go/src/github.com/docker/hub-tool/bin/${BINARY} ./bin/${BINARY}
RUN chmod +x ./bin/${BINARY}
CMD ["make", "-f", "builder.Makefile", "e2e"]
