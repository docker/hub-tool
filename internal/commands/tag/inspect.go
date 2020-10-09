/*
   Copyright 2020 Docker Hub Tool authors

   Licensed under the Apache License, Version 2.0 (the "License");
   you may not use this file except in compliance with the License.
   You may obtain a copy of the License at

       http://www.apache.org/licenses/LICENSE-2.0

   Unless required by applicable law or agreed to in writing, software
   distributed under the License is distributed on an "AS IS" BASIS,
   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
   See the License for the specific language governing permissions and
   limitations under the License.
*/

package tag

import (
	"encoding/json"
	"fmt"
	"io"
	"text/tabwriter"

	"github.com/containerd/containerd/images"
	"github.com/docker/buildx/util/imagetools"
	"github.com/docker/cli/cli"
	"github.com/docker/cli/cli/command"
	clitypes "github.com/docker/cli/cli/config/types"
	"github.com/docker/distribution/reference"
	"github.com/docker/docker/api/types"
	"github.com/docker/go-units"
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
	"github.com/spf13/cobra"

	"github.com/docker/hub-cli-plugin/internal/hub"
	"github.com/docker/hub-cli-plugin/internal/metrics"
)

const (
	inspectName = "inspect"
)

type inspectOptions struct {
	format string
}

func newInspectCmd(streams command.Streams, hubClient *hub.Client, parent string) *cobra.Command {
	var opts inspectOptions
	cmd := &cobra.Command{
		Use:   inspectName + " [OPTIONS] REPOSITORY:TAG",
		Short: "Show the details of an image in the registry",
		Args:  cli.ExactArgs(1),
		PreRun: func(cmd *cobra.Command, args []string) {
			metrics.Send(parent, inspectName)
		},
		RunE: func(_ *cobra.Command, args []string) error {
			return runInspect(streams, hubClient, opts, args[0])
		},
	}
	cmd.Flags().StringVar(&opts.format, "format", "", `Print original manifest ("json")`)
	cmd.Flags().SetInterspersed(false)
	return cmd
}

func runInspect(streams command.Streams, hubClient *hub.Client, opts inspectOptions, image string) error {
	resolver := imagetools.New(imagetools.Opt{
		Auth: &authResolver{
			authConfig: convert(hubClient.AuthConfig),
		},
	})

	raw, descriptor, err := resolver.Get(hubClient.Ctx, image)
	if err != nil {
		return err
	}

	if opts.format == "json" {
		fmt.Fprintf(streams.Out(), "%s", raw) // avoid newline to keep digest
		return nil
	} else if opts.format != "" {
		return fmt.Errorf("unsupported format type: %q", opts.format)
	}

	switch descriptor.MediaType {
	// case images.MediaTypeDockerSchema2Manifest, specs.MediaTypeImageManifest:
	// TODO: handle distribution manifest and schema1
	case images.MediaTypeDockerSchema2ManifestList, ocispec.MediaTypeImageIndex:
		if err := imagetools.PrintManifestList(raw, descriptor, image, streams.Out()); err != nil {
			return err
		}
	case images.MediaTypeDockerSchema2Manifest, ocispec.MediaTypeImageManifest:
		if err := printManifest(raw, descriptor, image, streams.Out()); err != nil {
			return err
		}
	default:
		fmt.Fprintf(streams.Out(), "%s\n", raw)
	}

	return nil
}

const (
	pfx = "  "
)

func printManifest(raw []byte, descriptor ocispec.Descriptor, image string, out io.Writer) error {
	ref, err := reference.ParseNormalizedNamed(image)
	if err != nil {
		return err
	}
	ref = reference.TagNameOnly(ref)

	var manifest ocispec.Manifest
	if err := json.Unmarshal(raw, &manifest); err != nil {
		return err
	}

	w := tabwriter.NewWriter(out, 0, 0, 1, ' ', 0)

	fmt.Fprintf(w, "Name:\t%s\n", ref.String())
	fmt.Fprintf(w, "MediaType:\t%s\n", descriptor.MediaType)
	fmt.Fprintf(w, "Digest:\t%s\n", descriptor.Digest)
	fmt.Fprintf(w, "\n")
	if err := w.Flush(); err != nil {
		return err
	}

	w = tabwriter.NewWriter(out, 0, 0, 1, ' ', 0)
	fmt.Fprintf(w, "Config:\n")
	fmt.Fprintf(w, "%sMediaType:\t%s\n", pfx, manifest.Config.MediaType)
	fmt.Fprintf(w, "%sSize:\t%v\n", pfx, units.HumanSize(float64(manifest.Config.Size)))
	fmt.Fprintf(w, "%sDigest:\t%s\n", pfx, manifest.Config.Digest)
	fmt.Fprintf(w, "\n")
	if err := w.Flush(); err != nil {
		return err
	}

	w = tabwriter.NewWriter(out, 0, 0, 1, ' ', 0)
	fmt.Fprintf(w, "Layers:\n")
	for i, layer := range manifest.Layers {
		if i != 0 {
			fmt.Fprintf(w, "\n")
		}
		fmt.Fprintf(w, "%sMediaType:\t%s\n", pfx, layer.MediaType)
		fmt.Fprintf(w, "%sSize:\t%v\n", pfx, units.HumanSize(float64(layer.Size)))
		fmt.Fprintf(w, "%sDigest:\t%s\n", pfx, layer.Digest)
	}

	return w.Flush()
}

type authResolver struct {
	authConfig clitypes.AuthConfig
}

func (a *authResolver) GetAuthConfig(registryHostname string) (clitypes.AuthConfig, error) {
	return a.authConfig, nil
}

func convert(config types.AuthConfig) clitypes.AuthConfig {
	return clitypes.AuthConfig{
		Username:      config.Username,
		Password:      config.Password,
		Auth:          config.Auth,
		Email:         config.Email,
		ServerAddress: config.ServerAddress,
		IdentityToken: config.IdentityToken,
		RegistryToken: config.RegistryToken,
	}
}
