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
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sort"
	"strings"
	"time"

	"github.com/containerd/containerd/images"
	"github.com/containerd/containerd/platforms"
	"github.com/containerd/containerd/remotes"
	"github.com/containerd/containerd/remotes/docker"
	"github.com/docker/cli/cli"
	"github.com/docker/cli/cli/command"
	"github.com/docker/distribution/reference"
	"github.com/docker/go-units"
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
	"github.com/spf13/cobra"

	"github.com/docker/hub-tool/internal/ansi"
	"github.com/docker/hub-tool/pkg/hub"
	"github.com/docker/hub-tool/internal/metrics"
)

const (
	inspectName = "inspect"
)

type inspectOptions struct {
	format   string
	platform string
}

//Image is the combination of a manifest and its config object
type Image struct {
	Name       string
	Manifest   ocispec.Manifest
	Config     ocispec.Image
	Descriptor ocispec.Descriptor
}

//Index is the combination of an OCI index and its descriptor
type Index struct {
	Name       string
	Index      ocispec.Index
	Descriptor ocispec.Descriptor
}

func newInspectCmd(streams command.Streams, hubClient *hub.Client, parent string) *cobra.Command {
	var opts inspectOptions
	cmd := &cobra.Command{
		Use:                   inspectName + " [OPTIONS] REPOSITORY:TAG",
		Short:                 "Show the details of an image in the registry",
		Args:                  cli.ExactArgs(1),
		DisableFlagsInUseLine: true,
		PreRun: func(cmd *cobra.Command, args []string) {
			metrics.Send(parent, inspectName)
		},
		RunE: func(_ *cobra.Command, args []string) error {
			return runInspect(streams, hubClient, opts, args[0])
		},
	}
	cmd.Flags().StringVar(&opts.format, "format", "", `Print original manifest ("json|raw")`)
	cmd.Flags().StringVar(&opts.platform, "platform", "", `Select a platform if the tag is a multi-architecture image`)
	return cmd
}

func runInspect(streams command.Streams, hubClient *hub.Client, opts inspectOptions, imageRef string) error {
	var (
		platform *ocispec.Platform
	)
	if opts.platform != "" {
		p, err := platforms.Parse(opts.platform)
		if err != nil {
			return fmt.Errorf("invalid platform %q: %s", opts.platform, err)
		}
		platform = &p
	}
	authorizer := docker.NewDockerAuthorizer(docker.WithAuthCreds(func(string) (string, string, error) {
		return hubClient.AuthConfig.Username, hubClient.AuthConfig.Password, nil
	}))
	registryHosts := docker.ConfigureDefaultRegistries(docker.WithClient(http.DefaultClient), docker.WithAuthorizer(authorizer))

	resolver := docker.NewResolver(docker.ResolverOptions{
		Hosts: registryHosts,
	})

	// Parse image reference
	ref, err := reference.ParseNormalizedNamed(imageRef)
	if err != nil {
		return err
	}
	ref = reference.TagNameOnly(ref)

	// Read descriptor
	fullName, descriptor, err := resolver.Resolve(hubClient.Ctx, ref.String())
	if err != nil {
		return err
	}

	raw, err := getBlob(hubClient.Ctx, resolver, fullName, descriptor)
	if err != nil {
		return err
	}

	switch descriptor.MediaType {
	// case images.MediaTypeDockerSchema2Manifest, specs.MediaTypeImageManifest:
	// TODO: handle distribution manifest and schema1
	case images.MediaTypeDockerSchema2ManifestList, ocispec.MediaTypeImageIndex:
		return formatManifestlist(hubClient.Ctx, streams, resolver, opts.format, raw, descriptor, ref.Name(), platform)
	case images.MediaTypeDockerSchema2Manifest, ocispec.MediaTypeImageManifest:
		return formatManifest(hubClient.Ctx, streams, resolver, opts.format, raw, descriptor, ref.Name())
	default:
		fmt.Fprintln(streams.Out(), ansi.Title("Unsupported mediatype"))
		fmt.Fprintln(streams.Out(), raw)
	}

	return nil
}

func getBlob(ctx context.Context, resolver remotes.Resolver, fullName string, descriptor ocispec.Descriptor) ([]byte, error) {
	// Fetch the blob
	fetcher, err := resolver.Fetcher(ctx, fullName)
	if err != nil {
		return nil, err
	}

	rc, err := fetcher.Fetch(ctx, descriptor)
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = rc.Close()
	}()

	// Read the blob
	buf := bytes.NewBuffer(nil)
	if _, err = io.Copy(buf, rc); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

func formatManifestlist(ctx context.Context, streams command.Streams, resolver remotes.Resolver,
	format string, raw []byte, descriptor ocispec.Descriptor, name string, platform *ocispec.Platform) error {
	var index ocispec.Index
	if err := json.Unmarshal(raw, &index); err != nil {
		return err
	}
	if platform != nil {
		return formatSelectManifest(ctx, streams, resolver, format, name, *platform, index)
	}

	image := Index{
		Name:       name,
		Index:      index,
		Descriptor: descriptor,
	}
	switch format {
	case "raw":
		_, err := fmt.Printf("%s", raw) // avoid newline to keep digest
		return err
	case "json":
		buf, err := json.MarshalIndent(index, "", "  ")
		if err != nil {
			return err
		}
		_, err = fmt.Fprint(streams.Out(), string(buf))
		return err
	case "":
		return printManifestList(streams.Out(), image)
	default:
		return fmt.Errorf("unsupported format type: %q", format)
	}
}

func formatManifest(ctx context.Context, streams command.Streams, resolver remotes.Resolver,
	format string, raw []byte, descriptor ocispec.Descriptor, name string) error {
	image, err := readImage(ctx, resolver, raw, descriptor, name)
	if err != nil {
		return err
	}
	switch format {
	case "raw":
		_, err := fmt.Printf("%s", raw) // avoid newline to keep digest
		return err
	case "json":
		buf, err := json.MarshalIndent(image, "", "  ")
		if err != nil {
			return err
		}
		_, err = fmt.Fprint(streams.Out(), string(buf))
		return err
	case "":
		return printImage(streams.Out(), image)
	default:
		return fmt.Errorf("unsupported format type: %q", format)
	}
}

func formatSelectManifest(ctx context.Context, streams command.Streams, resolver remotes.Resolver,
	format string, name string, platform ocispec.Platform, index ocispec.Index) error {
	matcher := platforms.NewMatcher(platform)
	var selectedDescriptor *ocispec.Descriptor
	for _, descriptor := range index.Manifests {
		if descriptor.Platform != nil && matcher.Match(*descriptor.Platform) {
			selectedDescriptor = &descriptor
			break
		}
	}
	if selectedDescriptor == nil {
		return fmt.Errorf("platform %q does not match any available platform for the tag %q", platforms.Format(platform), name)
	}
	raw, err := getBlob(ctx, resolver, name, *selectedDescriptor)
	if err != nil {
		return err
	}
	return formatManifest(ctx, streams, resolver, format, raw, *selectedDescriptor, name)
}

func readImage(ctx context.Context, resolver remotes.Resolver, rawManifest []byte, descriptor ocispec.Descriptor, name string) (*Image, error) {
	var manifest ocispec.Manifest
	if err := json.Unmarshal(rawManifest, &manifest); err != nil {
		return nil, err
	}

	configRef := fmt.Sprintf("%s@%s", name, manifest.Config.Digest)

	configRaw, err := getBlob(ctx, resolver, configRef, manifest.Config)
	if err != nil {
		return nil, err
	}
	var config ocispec.Image
	if err := json.Unmarshal(configRaw, &config); err != nil {
		return nil, err
	}
	return &Image{name, manifest, config, descriptor}, nil
}

func printImage(out io.Writer, image *Image) error {
	if err := printManifest(out, image); err != nil {
		return err
	}
	if err := printConfig(out, image); err != nil {
		return err
	}

	return printLayers(out, image)
}

func printManifestList(out io.Writer, image Index) error {
	fmt.Fprintf(out, ansi.Title("Manifest List:")+"\n")
	fmt.Fprintf(out, ansi.Key("Name:")+"\t\t%s\n", image.Name)
	fmt.Fprintf(out, ansi.Key("MediaType:")+"\t%s\n", image.Descriptor.MediaType)
	fmt.Fprintf(out, ansi.Key("Digest:")+"\t\t%s\n", image.Descriptor.Digest)
	if len(image.Index.Annotations) > 0 {
		printAnnotations(out, image.Index.Annotations)
	} else if len(image.Descriptor.Annotations) > 0 {
		printAnnotations(out, image.Descriptor.Annotations)
	}

	fmt.Fprintf(out, "\n")

	fmt.Fprintf(out, ansi.Title("Manifests:")+"\n")
	for i, m := range image.Index.Manifests {
		if i != 0 {
			fmt.Fprintln(out)
		}
		fmt.Fprintf(out, ansi.Key("Name:")+"\t\t%s\n", fmt.Sprintf("%s@%s", image.Name, m.Digest))
		fmt.Fprintf(out, ansi.Key("Mediatype:")+"\t%s\n", m.MediaType)
		fmt.Fprintf(out, ansi.Key("Platform:")+"\t%s\n", formatPlatform(m.Platform))
	}

	return nil
}

func printManifest(out io.Writer, image *Image) error {
	fmt.Fprintf(out, ansi.Title("Manifest:")+"\n")
	fmt.Fprintf(out, ansi.Key("Name:")+"\t\t%s\n", image.Name)
	fmt.Fprintf(out, ansi.Key("MediaType:")+"\t%s\n", image.Descriptor.MediaType)
	fmt.Fprintf(out, ansi.Key("Digest:")+"\t\t%s\n", image.Descriptor.Digest)
	if image.Descriptor.Platform != nil {
		fmt.Fprintf(out, ansi.Key("Platform:")+"\t%s\n", formatPlatform(image.Descriptor.Platform))
	}
	if len(image.Manifest.Annotations) > 0 {
		printAnnotations(out, image.Manifest.Annotations)
	} else if len(image.Descriptor.Annotations) > 0 {
		printAnnotations(out, image.Descriptor.Annotations)
	}
	if image.Config.Architecture != "" {
		fmt.Fprintf(out, ansi.Key("Os/Arch:")+"\t%s/%s\n", image.Config.OS, image.Config.Architecture)
	}
	if image.Config.Author != "" {
		fmt.Fprintf(out, ansi.Key("Author:")+"\t%s\n", image.Config.Author)
	}
	if image.Config.Created != nil {
		fmt.Fprintf(out, ansi.Key("Created:")+"\t%s ago\n", units.HumanDuration(time.Since(*image.Config.Created)))
	}

	fmt.Fprintf(out, "\n")

	return nil
}

func printConfig(out io.Writer, image *Image) error {
	fmt.Fprintf(out, ansi.Title("Config:")+"\n")
	fmt.Fprintf(out, ansi.Key("MediaType:")+"\t%s\n", image.Manifest.Config.MediaType)
	fmt.Fprintf(out, ansi.Key("Size:")+"\t\t%v\n", units.HumanSize(float64(image.Manifest.Config.Size)))
	fmt.Fprintf(out, ansi.Key("Digest:")+"\t\t%s\n", image.Manifest.Config.Digest)
	if len(image.Config.Config.Cmd) > 0 {
		fmt.Fprintf(out, ansi.Key("Command:")+"\t%q\n", strings.TrimPrefix(strings.Join(image.Config.Config.Cmd, " "), "/bin/sh -c "))
	}
	if len(image.Config.Config.Entrypoint) > 0 {
		fmt.Fprintf(out, ansi.Key("Entrypoint:")+"\t%q\n", strings.Join(image.Config.Config.Entrypoint, " "))
	}
	if image.Config.Config.User != "" {
		fmt.Fprintf(out, ansi.Key("User:")+"\t%s\n", image.Config.Config.User)
	}
	if len(image.Config.Config.ExposedPorts) > 0 {
		fmt.Fprintf(out, ansi.Key("Exposed ports:")+"\t%s\n", getExposedPorts(image.Config.Config.ExposedPorts))
	}
	if len(image.Config.Config.Env) > 0 {
		fmt.Fprintf(out, ansi.Key("Environment:")+"\n")
		for _, env := range image.Config.Config.Env {
			fmt.Fprintf(out, "    %s\n", env)
		}
	}
	if len(image.Config.Config.Volumes) > 0 {
		fmt.Fprintf(out, ansi.Key("Volumes:")+"\n")
		for volume := range image.Config.Config.Volumes {
			fmt.Fprintf(out, "%s\n", volume)
		}
	}
	if image.Config.Config.WorkingDir != "" {
		fmt.Fprintf(out, ansi.Key("Working Directory:")+"\t%q\n", image.Config.Config.WorkingDir)
	}
	if len(image.Config.Config.Labels) > 0 {
		fmt.Fprintf(out, ansi.Key("Labels:")+"\n")
		keys := sortMapKeys(image.Config.Config.Labels)
		for _, k := range keys {
			fmt.Fprintf(out, "    %s=%q\n", k, image.Config.Config.Labels[k])
		}
	}
	if image.Config.Config.StopSignal != "" {
		fmt.Fprintf(out, ansi.Key("Stop signal:")+"\t\t%s\n", image.Config.Config.StopSignal)
	}

	fmt.Fprintf(out, "\n")
	return nil
}

func printLayers(out io.Writer, image *Image) error {
	history := filterEmptyLayers(image.Config.History)
	fmt.Fprintln(out, ansi.Title("Layers:"))
	for i, layer := range image.Manifest.Layers {
		if i != 0 {
			fmt.Fprintln(out)
		}
		fmt.Fprintf(out, ansi.Key("MediaType:")+"\t%s\n", layer.MediaType)
		fmt.Fprintf(out, ansi.Key("Size:")+"\t\t%v\n", units.HumanSize(float64(layer.Size)))
		fmt.Fprintf(out, ansi.Key("Digest:")+"\t\t%s\n", layer.Digest)
		if len(image.Manifest.Layers) == len(history) {
			fmt.Fprintf(out, ansi.Key("Command:")+"\t%s\n", cleanCreatedBy(history[i].CreatedBy))
			if history[i].Created != nil {
				fmt.Fprintf(out, ansi.Key("Created:")+"\t%s ago\n", units.HumanDuration(time.Since(*history[i].Created)))
			}
		}
	}
	return nil
}

func formatPlatform(platform *ocispec.Platform) string {
	if platform == nil {
		return ""
	}
	return platforms.Format(*platform)
}

func cleanCreatedBy(history string) string {
	history = strings.TrimPrefix(history, "/bin/sh -c #(nop) ")
	history = strings.TrimPrefix(history, "/bin/sh -c ")
	return strings.TrimSpace(history)
}

func filterEmptyLayers(history []ocispec.History) []ocispec.History {
	var result []ocispec.History
	for _, h := range history {
		if !h.EmptyLayer {
			result = append(result, h)
		}
	}
	return result
}

func getExposedPorts(configPorts map[string]struct{}) string {
	var ports []string
	for port := range configPorts {
		ports = append(ports, port)
	}
	return strings.Join(ports, " ")
}

func printAnnotations(out io.Writer, annotations map[string]string) {
	fmt.Fprintf(out, ansi.Key("Annotations:")+"\n")
	keys := sortMapKeys(annotations)
	for _, k := range keys {
		fmt.Fprintf(out, "%s:\t%s\n", k, annotations[k])
	}
}

func sortMapKeys(m map[string]string) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}
