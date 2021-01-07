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

- [Docker](https://www.docker.com/products/docker-desktop) installed on your
  system
- [A Docker Hub account](https://hub.docker.com)

### Install

- Download the latest release for your platform from
  [here](https://github.com/docker/hub-tool/releases)
- Extract the package and place the `hub-tool` binary somewhere in your `PATH`

### Login to Docker Hub

Login to the [Docker Hub](https://hub.docker.com) using your username and
password:

```console
hub-tool login yourusername
```

> **Note:** When using a
> [personal access token (PAT)](https://docs.docker.com/docker-hub/access-tokens/),
> not all functionality will be available.

### Listing tags

```console
$ hub-tool tag ls docker
TAG                                   DIGEST                                                                     STATUS    LAST UPDATE    LAST PUSHED    LAST PULLED    SIZE
docker:stable-dind-rootless           sha256:c96432c62569526fc710854c4d8441dae22907119c8987a5e82a2868bd509fd4    stale     3 days ago     3 days                        96.55MB
docker:stable-dind                    sha256:f998921d365053bf7e3f98794f6c23ca44e6809832d78105bc4d2da6bb8521ed    stale     3 days ago     3 days                        274.6MB
docker:rc-git                         sha256:2c4980f5700c775634dd997484834ba0c6f63c5e2384d22c23c067afec8f2596    stale     3 days ago     3 days                        302.6MB
docker:rc-dind-rootless               sha256:ed25cf41ad0d739e26e2416fb97858758f3cfd1c6345a11c2d386bff567e4060    stale     3 days ago     3 days                        103.5MB
docker:rc-dind                        sha256:a1e9f065ea4b31de9aeed07048cf820a64b8637262393b24a4216450da46b7d6    stale     3 days ago     3 days                        288.9MB
docker:rc                             sha256:f8ecea9dc16c9f6471448a78d3e101a3f864be71bfe3b8b27cac6df83f6f0970    stale     3 days ago     3 days                        270.9MB
...
25/957 listed, use --all flag to show all
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

- [Docker](https://www.docker.com/products/docker-desktop)
- `make`

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
