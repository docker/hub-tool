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
	"context"
	"encoding/json"
	"fmt"
	"io"
	"sort"
	"strings"
	"time"

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

	"github.com/docker/hub-tool/internal/ansi"
	"github.com/docker/hub-tool/internal/hub"
	"github.com/docker/hub-tool/internal/metrics"
)

const (
	inspectName = "inspect"
)

type inspectOptions struct {
	format string
}

//Image is the combination of a manifest and its config object
type Image struct {
	Name       string
	Manifest   ocispec.Manifest
	Config     ocispec.Image
	Descriptor ocispec.Descriptor
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
	cmd.Flags().StringVar(&opts.format, "format", "", `Print original manifest ("json|raw")`)
	cmd.Flags().SetInterspersed(false)
	return cmd
}

func runInspect(streams command.Streams, hubClient *hub.Client, opts inspectOptions, imageRef string) error {
	resolver := imagetools.New(imagetools.Opt{
		Auth: &authResolver{
			authConfig: convert(hubClient.AuthConfig),
		},
	})

	raw, descriptor, err := resolver.Get(hubClient.Ctx, imageRef)
	if err != nil {
		return err
	}

	ref, err := reference.ParseNormalizedNamed(imageRef)
	if err != nil {
		return err
	}
	ref = reference.TagNameOnly(ref)

	switch descriptor.MediaType {
	// case images.MediaTypeDockerSchema2Manifest, specs.MediaTypeImageManifest:
	// TODO: handle distribution manifest and schema1
	case images.MediaTypeDockerSchema2ManifestList, ocispec.MediaTypeImageIndex:
		return formatManifestlist(streams, opts.format, raw, descriptor, imageRef)
	case images.MediaTypeDockerSchema2Manifest, ocispec.MediaTypeImageManifest:
		return formatManifest(hubClient.Ctx, streams, resolver, opts.format, raw, descriptor, ref.Name())
	default:
		fmt.Println("Other mediatype")
		fmt.Printf("%s\n", raw)
	}

	return nil
}

func formatManifestlist(streams command.Streams, format string, raw []byte, descriptor ocispec.Descriptor, imageRef string) error {
	switch format {
	case "json", "raw":
		_, err := fmt.Printf("%s", raw) // avoid newline to keep digest
		return err
	case "":
		return imagetools.PrintManifestList(raw, descriptor, imageRef, streams.Out())
	default:
		return fmt.Errorf("unsupported format type: %q", format)
	}
}

func formatManifest(ctx context.Context, streams command.Streams, resolver *imagetools.Resolver, format string, raw []byte, descriptor ocispec.Descriptor, name string) error {
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

func readImage(ctx context.Context, resolver *imagetools.Resolver, rawManifest []byte, descriptor ocispec.Descriptor, name string) (*Image, error) {
	var manifest ocispec.Manifest
	if err := json.Unmarshal(rawManifest, &manifest); err != nil {
		return nil, err
	}

	configRef := fmt.Sprintf("%s@%s", name, manifest.Config.Digest)
	configRaw, err := resolver.GetDescriptor(ctx, configRef, manifest.Config)
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

func printManifest(out io.Writer, image *Image) error {
	fmt.Fprintf(out, ansi.Title("Manifest:")+"\n")
	fmt.Fprintf(out, ansi.Key("Name:")+"\t\t%s\n", image.Name)
	fmt.Fprintf(out, ansi.Key("MediaType:")+"\t%s\n", image.Descriptor.MediaType)
	fmt.Fprintf(out, ansi.Key("Digest:")+"\t\t%s\n", image.Descriptor.Digest)
	if image.Descriptor.Platform != nil {
		fmt.Fprintf(out, ansi.Key("Platform:")+"\t%s\n", getPlatform(image.Descriptor.Platform))
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
	fmt.Fprintf(out, ansi.Key("Digest:")+"\t%s\n", image.Manifest.Config.Digest)
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
		fmt.Fprintf(out, ansi.Key("Digest:")+"\t%s\n", layer.Digest)
		if len(image.Manifest.Layers) == len(history) {
			fmt.Fprintf(out, ansi.Key("Command:")+"\t%s\n", cleanCreatedBy(history[i].CreatedBy))
			if history[i].Created != nil {
				fmt.Fprintf(out, ansi.Key("Created:")+"\t%s ago\n", units.HumanDuration(time.Since(*history[i].Created)))
			}
		}
	}
	return nil
}

func getPlatform(platform *ocispec.Platform) string {
	result := fmt.Sprintf("%s/%s", platform.OS, platform.Architecture)
	if platform.Variant != "" {
		result += "/" + platform.Variant
	}
	if platform.OSVersion != "" {
		result += "/" + platform.OSVersion
	}
	if len(platform.OSFeatures) > 0 {
		result += "/" + strings.Join(platform.OSFeatures, "/")
	}
	return result
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

func sortMapKeys(m map[string]string) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}
