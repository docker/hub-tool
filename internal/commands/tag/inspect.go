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
	"github.com/juju/ansiterm"
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
	"github.com/spf13/cobra"

	"github.com/docker/hub-tool/internal/color"
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
	Name       reference.Named
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
		return formatManifest(hubClient.Ctx, streams, resolver, opts.format, raw, descriptor, ref)
	default:
		fmt.Println("Other mediatype")
		fmt.Fprintf(streams.Out(), "%s\n", raw)
	}

	return nil
}

const (
	prefix = "  "
)

func formatManifestlist(streams command.Streams, format string, raw []byte, descriptor ocispec.Descriptor, imageRef string) error {
	switch format {
	case "json", "raw":
		_, err := fmt.Fprintf(streams.Out(), "%s", raw) // avoid newline to keep digest
		return err
	case "":
		return imagetools.PrintManifestList(raw, descriptor, imageRef, streams.Out())
	default:
		return fmt.Errorf("unsupported format type: %q", format)
	}
}

func formatManifest(ctx context.Context, streams command.Streams, resolver *imagetools.Resolver, format string, raw []byte, descriptor ocispec.Descriptor, ref reference.Named) error {
	image, err := readImage(ctx, resolver, raw, descriptor, ref)
	if err != nil {
		return err
	}
	switch format {
	case "raw":
		_, err := fmt.Fprintf(streams.Out(), "%s", raw) // avoid newline to keep digest
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

func readImage(ctx context.Context, resolver *imagetools.Resolver, rawManifest []byte, descriptor ocispec.Descriptor, ref reference.Named) (*Image, error) {
	var manifest ocispec.Manifest
	if err := json.Unmarshal(rawManifest, &manifest); err != nil {
		return nil, err
	}

	configRef := fmt.Sprintf("%s@%s", ref.Name(), manifest.Config.Digest)
	configRaw, err := resolver.GetDescriptor(ctx, configRef, manifest.Config)
	if err != nil {
		return nil, err
	}
	var config ocispec.Image
	if err := json.Unmarshal(configRaw, &config); err != nil {
		return nil, err
	}
	return &Image{ref, manifest, config, descriptor}, nil
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
	w := ansiterm.NewTabWriter(out, 0, 0, 1, ' ', 0)

	fmt.Fprintf(w, color.Title("Manifest:")+"\n")
	fmt.Fprintf(w, color.Key("%sName:")+"\t%s\n", prefix, image.Name)
	fmt.Fprintf(w, color.Key("%sMediaType:")+"\t%s\n", prefix, image.Descriptor.MediaType)
	fmt.Fprintf(w, color.Key("%sDigest:")+"\t%s\n", prefix, image.Descriptor.Digest)
	if image.Descriptor.Platform != nil {
		fmt.Fprintf(w, color.Key("%sPlatform:")+"\t%s\n", prefix, getPlatform(image.Descriptor.Platform))
	}
	if len(image.Manifest.Annotations) > 0 {
		printAnnotations(w, image.Manifest.Annotations)
	} else if len(image.Descriptor.Annotations) > 0 {
		printAnnotations(w, image.Descriptor.Annotations)
	}
	if image.Config.Architecture != "" {
		fmt.Fprintf(w, color.Key("%sOs/Arch:")+"\t%s/%s\n", prefix, image.Config.OS, image.Config.Architecture)
	}
	if image.Config.Author != "" {
		fmt.Fprintf(w, color.Key("%sAuthor:")+"\t%s\n", prefix, image.Config.Author)
	}
	if image.Config.Created != nil {
		fmt.Fprintf(w, color.Key("%sCreated:")+"\t%s ago\n", prefix, units.HumanDuration(time.Since(*image.Config.Created)))
	}
	fmt.Fprintf(w, "\n")
	return w.Flush()
}

func printConfig(out io.Writer, image *Image) error {
	w := ansiterm.NewTabWriter(out, 0, 0, 1, ' ', 0)
	fmt.Fprintf(w, color.Title("Config:")+"\n")
	fmt.Fprintf(w, color.Key("%sMediaType:")+"\t%s\n", prefix, image.Manifest.Config.MediaType)
	fmt.Fprintf(w, color.Key("%sSize:")+"\t%v\n", prefix, units.HumanSize(float64(image.Manifest.Config.Size)))
	fmt.Fprintf(w, color.Key("%sDigest:")+"\t%s\n", prefix, image.Manifest.Config.Digest)
	if len(image.Config.Config.Cmd) > 0 {
		fmt.Fprintf(w, color.Key("%sCommand:")+"\t%q\n", prefix, strings.TrimPrefix(strings.Join(image.Config.Config.Cmd, " "), "/bin/sh -c "))
	}
	if len(image.Config.Config.Entrypoint) > 0 {
		fmt.Fprintf(w, color.Key("%sEntrypoint:")+"\t%q\n", prefix, strings.Join(image.Config.Config.Entrypoint, " "))
	}
	if image.Config.Config.User != "" {
		fmt.Fprintf(w, color.Key("%sUser:")+"\t%s\n", prefix, image.Config.Config.User)
	}
	if len(image.Config.Config.ExposedPorts) > 0 {
		fmt.Fprintf(w, color.Key("%sExposed ports:")+"\t%s\n", prefix, getExposedPorts(image.Config.Config.ExposedPorts))
	}
	if len(image.Config.Config.Env) > 0 {
		fmt.Fprintf(w, color.Key("%sEnvironment:")+"\n", prefix)
		for _, env := range image.Config.Config.Env {
			fmt.Fprintf(w, "%s%s%s\n", prefix, prefix, env)
		}
	}
	if len(image.Config.Config.Volumes) > 0 {
		fmt.Fprintf(w, color.Key("%sVolumes:")+"\n", prefix)
		for volume := range image.Config.Config.Volumes {
			fmt.Fprintf(w, "%s%s%s\n", prefix, prefix, volume)
		}
	}
	if image.Config.Config.WorkingDir != "" {
		fmt.Fprintf(w, color.Key("%sWorking Directory:")+"\t%q\n", prefix, image.Config.Config.WorkingDir)
	}
	if len(image.Config.Config.Labels) > 0 {
		fmt.Fprintf(w, color.Key("%sLabels:")+"\n", prefix)
		keys := sortMapKeys(image.Config.Config.Labels)
		for _, k := range keys {
			fmt.Fprintf(w, "%s%s%s=%q\n", prefix, prefix, k, image.Config.Config.Labels[k])
		}
	}
	if image.Config.Config.StopSignal != "" {
		fmt.Fprintf(w, color.Key("%sStop signal:")+"\t%s\n", prefix, image.Config.Config.StopSignal)
	}

	fmt.Fprintf(w, "\n")
	return w.Flush()
}

func printLayers(out io.Writer, image *Image) error {
	history := filterEmptyLayers(image.Config.History)
	w := ansiterm.NewTabWriter(out, 0, 0, 1, ' ', 0)
	fmt.Fprintln(w, color.Title("Layers:"))
	for i, layer := range image.Manifest.Layers {
		if i != 0 {
			fmt.Fprintln(w)
		}
		fmt.Fprintf(w, color.Key("%sMediaType:")+"\t%s\n", prefix, layer.MediaType)
		fmt.Fprintf(w, color.Key("%sSize:")+"\t%v\n", prefix, units.HumanSize(float64(layer.Size)))
		fmt.Fprintf(w, color.Key("%sDigest:")+"\t%s\n", prefix, layer.Digest)
		if len(image.Manifest.Layers) == len(history) {
			fmt.Fprintf(w, color.Key("%sCommand:")+"\t%s\n", prefix, cleanCreatedBy(history[i].CreatedBy))
			if history[i].Created != nil {
				fmt.Fprintf(w, color.Key("%sCreated:")+"\t%s ago\n", prefix, units.HumanDuration(time.Since(*history[i].Created)))
			}
		}
	}
	return w.Flush()
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

func printAnnotations(w io.Writer, annotations map[string]string) {
	fmt.Fprintf(w, color.Key("%sAnnotations:")+"\n", prefix)
	keys := sortMapKeys(annotations)
	for _, k := range keys {
		fmt.Fprintf(w, "%s%s%s:\t%s\n", prefix, prefix, k, annotations[k])
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
