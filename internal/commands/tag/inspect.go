/*
   Copyright 2020 Docker Inc.

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
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"text/tabwriter"

	"github.com/containerd/containerd/images"
	"github.com/docker/buildx/util/imagetools"
	"github.com/docker/cli/cli"
	"github.com/docker/cli/cli/command"
	"github.com/docker/distribution/reference"
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
	"github.com/spf13/cobra"

	"github.com/docker/hub-cli-plugin/internal/metrics"
)

const (
	inspectName = "inspect"
)

type inspectOptions struct {
	raw bool
}

func newInspectCmd(ctx context.Context, dockerCli command.Cli, parent string) *cobra.Command {
	var opts inspectOptions
	cmd := &cobra.Command{
		Use:   inspectName + " [OPTIONS] REPOSITORY:TAG",
		Short: "Show the details of an image in the registry",
		Args:  cli.ExactArgs(1),
		PreRun: func(cmd *cobra.Command, args []string) {
			metrics.Send(parent, inspectName)
		},
		RunE: func(_ *cobra.Command, args []string) error {
			return runInspect(ctx, dockerCli, opts, args[0])
		},
	}
	cmd.Flags().BoolVar(&opts.raw, "raw", false, "Show original JSON manifest")
	return cmd
}

func runInspect(ctx context.Context, dockerCli command.Cli, opts inspectOptions, image string) error {
	resolver := imagetools.New(imagetools.Opt{
		Auth: dockerCli.ConfigFile(),
	})

	raw, descriptor, err := resolver.Get(ctx, image)
	if err != nil {
		return err
	}

	if opts.raw {
		fmt.Printf("%s", raw) // avoid newline to keep digest
		return nil
	}

	switch descriptor.MediaType {
	// case images.MediaTypeDockerSchema2Manifest, specs.MediaTypeImageManifest:
	// TODO: handle distribution manifest and schema1
	case images.MediaTypeDockerSchema2ManifestList, ocispec.MediaTypeImageIndex:
		if err := imagetools.PrintManifestList(raw, descriptor, image, os.Stdout); err != nil {
			return err
		}
	case images.MediaTypeDockerSchema2Manifest, ocispec.MediaTypeImageManifest:
		if err := printManifest(raw, descriptor, image, os.Stdout); err != nil {
			return err
		}
	default:
		fmt.Printf("%s\n", raw)
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
	fmt.Fprintf(w, "\t\n")
	if err := w.Flush(); err != nil {
		return err
	}

	w = tabwriter.NewWriter(out, 0, 0, 1, ' ', 0)
	fmt.Fprintf(w, "Config:\t\n")
	fmt.Fprintf(w, "%sMediaType:\t%s\n", pfx, manifest.Config.MediaType)
	fmt.Fprintf(w, "%sSize:\t%v\n", pfx, manifest.Config.Size)
	fmt.Fprintf(w, "%sDigest:\t%s\n", pfx, manifest.Config.Digest)
	fmt.Fprintf(w, "\t\n")
	if err := w.Flush(); err != nil {
		return err
	}

	w = tabwriter.NewWriter(os.Stdout, 0, 0, 1, ' ', 0)
	fmt.Fprintf(w, "Layers:\t\n")
	for i, layer := range manifest.Layers {
		if i != 0 {
			fmt.Fprintf(w, "\t\n")
		}
		fmt.Fprintf(w, "%sMediaType:\t%s\n", pfx, layer.MediaType)
		fmt.Fprintf(w, "%sSize:\t%v\n", pfx, manifest.Config.Size)
		fmt.Fprintf(w, "%sDigest:\t%s\n", pfx, manifest.Config.Digest)
	}

	return w.Flush()
}
