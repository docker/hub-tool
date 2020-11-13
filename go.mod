module github.com/docker/hub-tool

go 1.15

require (
	github.com/bitly/go-hostpool v0.1.0 // indirect
	github.com/bugsnag/bugsnag-go v1.5.3 // indirect
	github.com/cenkalti/backoff v2.2.1+incompatible // indirect
	github.com/cli/cli v1.1.0
	github.com/cloudflare/cfssl v1.4.1 // indirect
	github.com/containerd/containerd v1.4.0-0
	github.com/containerd/continuity v0.0.0-20200710164510-efbc4488d8fe // indirect
	github.com/docker/buildx v0.4.2
	github.com/docker/cli v0.0.0-20200528204125-dd360c7c0de8
	github.com/docker/compose-cli v0.1.18
	github.com/docker/distribution v2.7.1+incompatible
	github.com/docker/docker v17.12.0-ce-rc1.0.20200309214505-aa6a9891b09c+incompatible
	github.com/docker/go-units v0.4.0
	github.com/docker/libtrust v0.0.0-20160708172513-aabc10ec26b7 // indirect
	github.com/gofrs/uuid v3.3.0+incompatible // indirect
	github.com/google/uuid v1.1.2
	github.com/gorilla/mux v1.8.0 // indirect
	github.com/jinzhu/gorm v1.9.16 // indirect
	github.com/lib/pq v1.8.0 // indirect
	github.com/mattn/go-sqlite3 v1.14.3 // indirect
	github.com/miekg/pkcs11 v1.0.3 // indirect
	github.com/opencontainers/image-spec v1.0.1
	github.com/pkg/errors v0.9.1
	github.com/sirupsen/logrus v1.6.0
	github.com/spf13/cobra v1.0.0
	github.com/spf13/pflag v1.0.5
	golang.org/x/sync v0.0.0-20200625203802-6e8e738ad208
	gotest.tools v2.2.0+incompatible
	gotest.tools/v3 v3.0.2
)

replace github.com/docker/docker => github.com/docker/docker v1.4.2-0.20200227233006-38f52c9fec82

replace github.com/containerd/containerd => github.com/containerd/containerd v1.3.1-0.20200227195959-4d242818bf55

replace github.com/docker/distribution => github.com/docker/distribution v0.0.0-20200708230824-53e18a9d9bfe
