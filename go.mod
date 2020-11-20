module github.com/docker/hub-tool

go 1.15

require (
	github.com/bitly/go-hostpool v0.1.0 // indirect
	github.com/bugsnag/bugsnag-go v1.5.3 // indirect
	github.com/cenkalti/backoff v2.2.1+incompatible // indirect
	github.com/cli/cli v1.1.0
	github.com/cloudflare/cfssl v1.4.1 // indirect
	github.com/containerd/containerd v1.4.1
	github.com/docker/buildx v0.4.2
	github.com/docker/cli v20.10.0-rc1+incompatible
	github.com/docker/compose-cli v1.0.2
	github.com/docker/distribution v2.7.1+incompatible
	github.com/docker/docker v20.10.0-rc1+incompatible
	github.com/docker/go-units v0.4.0
	github.com/docker/libtrust v0.0.0-20160708172513-aabc10ec26b7 // indirect
	github.com/fvbommel/sortorder v1.0.2 // indirect
	github.com/gofrs/uuid v3.3.0+incompatible // indirect
	github.com/google/uuid v1.1.2
	github.com/jinzhu/gorm v1.9.16 // indirect
	github.com/lib/pq v1.8.0 // indirect
	github.com/mattn/go-sqlite3 v1.14.3 // indirect
	github.com/miekg/pkcs11 v1.0.3 // indirect
	github.com/moby/sys/mount v0.2.0 // indirect
	github.com/moby/term v0.0.0-20200915141129-7f0af18e79f2
	github.com/opencontainers/image-spec v1.0.1
	github.com/pkg/errors v0.9.1
	github.com/sirupsen/logrus v1.7.0
	github.com/spf13/cobra v1.1.1
	github.com/spf13/pflag v1.0.5
	golang.org/x/sync v0.0.0-20201020160332-67f06af15bc9
	gopkg.in/square/go-jose.v2 v2.5.1
	gotest.tools/v3 v3.0.3
)

replace (
	github.com/containerd/containerd => github.com/containerd/containerd v1.4.1
	github.com/docker/distribution => github.com/docker/distribution v0.0.0-20201114020330-551158e6008e
	github.com/docker/docker => github.com/docker/docker v20.10.0-rc1+incompatible
)
