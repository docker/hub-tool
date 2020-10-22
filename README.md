# 🧪 Docker Hub Tool

> :warning: This tool is a Docker experiment to build a Docker Hub CLI tool.
> The intention of this project is to get user feedback and then to add this
> functionality to the Docker CLI.

The Docker Hub Tool is a CLI tool for interacting with the
[Docker Hub](https://hub.docker.com).
It makes it easy to get information about your images from the terminal and to
perform Hub maintenance tasks.

## Get started

### Prerequisites

* [Docker](https://www.docker.com/products/docker-desktop) installed on your
  system
* [A Docker Hub account](https://hub.docker.com)

### Install

* Download the latest release for your platform from
  [here](https://github.com/docker/hub-tool/releases)
* Extract the package and place the `hub-tool` binary somewhere in your `PATH`

### Login to Docker Hub

Login to the [Docker Hub](https://hub.docker.com) using your username and
password:

```console
docker login yourusername
```

> **Note:** When using a
> [personal access token (PAT)](https://docs.docker.com/docker-hub/access-tokens/),
> not all functionality will be available.

### Listing tags

```console
$ hub-tool tag ls docker
TAG                              DIGEST                                                                     STATUS    EXPIRES    LAST UPDATE     LAST PUSHED    LAST PULLED    SIZE
docker:latest                    sha256:279beeb5de99e09af79f13e85e20194ce68db4255e8b2d955e408be69d082b5a                         10 hours ago                                  256.7MB
docker:test-git                  sha256:e89d2f422796bb472a3f6c301076f8f64fb9f6c3078ff96a8cc7918121a9130f                         10 hours ago                                  288.3MB
docker:test-dind-rootless        sha256:7e88eb523dd692072fa8f8467730df9be4dfff616475dc7c64dacf5f7527088f                         10 hours ago                                  96.55MB
docker:test-dind                 sha256:a6b0193cbf4d3c304f3bf6c6c253d88c25a22c6ffe6847fd57a6269e4324745f                         10 hours ago                                  274.6MB
docker:test                      sha256:18d39b6848cecae067cc0d94c554029bfc88d3069c80bb5049d54da659249b94                         10 hours ago                                  256.7MB
docker:stable-git                sha256:e89d2f422796bb472a3f6c301076f8f64fb9f6c3078ff96a8cc7918121a9130f                         10 hours ago                                  288.3MB
...
25/949 listed, use --all flag to show all
```

## Contributing

Docker wants to work with the community to make a tool that is useful and to
ensure that its UX is good. Remember that this is an experiment with the goal of
incorporating the learnings into the Docker CLI so it has some rough edges and
it's not meant to be a final product.

### Feedback

Please leave your feedback in the
[issue tracker](https://github.com/docker/hub-tool/issues)!
We'd love to know how you're using this tool and what features you'd like to see
us add.

### Code

At this stage of the project, we're mostly looking for feedback. We will accept
pull requests but these should be limited to minor improvements and fixes.
Anything larger should first be discussed as an issue.
If you spot a bug or see a typo, please feel free to fix it by putting up a
[pull request](https://github.com/docker/hub-tool/pulls)!


## Building

### Prerequisites

* [Docker](https://www.docker.com/products/docker-desktop)
* `make`

### Compiling

To build for your current platform, simply run `make` and the tool will be
output into the `./bin` directory:

```console
$ make
docker build --build-arg GO_VERSION=1.15.3 --build-arg ALPINE_VERSION=3.12.0 --build-arg GOLANGCI_LINT_VERSION=v1.31.0-alpine --build-arg TAG_NAME= --build-arg GOTESTSUM_VERSION=0.5.2 --build-arg BINARY_NAME=hub-tool --build-arg BINARY=hub-tool . \
                --output type=local,dest=./bin \
                --platform local \
                --target hub
[+] Building 3.7s (6/13)
...
 => => copying files 22.10MB

 $ ls bin/
 hub-tool
```
