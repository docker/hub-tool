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
	"fmt"
	"os"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/docker/cli/cli"
	"github.com/docker/cli/cli/command"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/registry"
	"github.com/docker/go-units"
	"github.com/spf13/cobra"

	"github.com/docker/hub-cli-plugin/internal/hub"
)

var (
	defaultColumns = []column{
		{"TAG", func(t hub.Tag) string { return t.Name }},
		{"DIGEST", func(t hub.Tag) string {
			if len(t.Images) > 0 {
				return t.Images[0].Digest
			}
			return ""
		}},
		{"STATUS", func(t hub.Tag) string {
			if len(t.Images) > 0 {
				return t.Images[0].Status // TODO: status should be on the tag/manifest list level too, not only on the image level
			}
			return ""
		}},
		{"EXPIRES", func(t hub.Tag) string {
			if len(t.Images) > 0 && t.Images[0].Expires.Nanosecond() != 0 {
				return units.HumanDuration(time.Until(t.Images[0].Expires))
			}
			return ""
		}},
		{"LAST UPDATE", func(t hub.Tag) string {
			if t.LastUpdated.Nanosecond() == 0 {
				return ""
			}
			return fmt.Sprintf("%s ago", units.HumanDuration(time.Since(t.LastUpdated)))
		}},
		{"LAST PUSHED", func(t hub.Tag) string {
			if len(t.Images) > 0 && t.Images[0].LastPushed.Nanosecond() != 0 {
				return units.HumanDuration(time.Since(t.Images[0].LastPushed))
			}
			return ""
		}},
		{"LAST PULLED", func(t hub.Tag) string {
			if len(t.Images) > 0 && t.Images[0].LastPulled.Nanosecond() != 0 {
				return units.HumanDuration(time.Since(t.Images[0].LastPulled))
			}
			return ""
		}},
		{"SIZE", func(t hub.Tag) string {
			size := t.FullSize
			if len(t.Images) > 0 {
				size = 0
				for _, image := range t.Images {
					size += image.Size
				}
			}
			return units.HumanSize(float64(size))
		}},
	}
	platformColumn = column{
		"OS/ARCH",
		func(t hub.Tag) string {
			var platforms []string
			for _, image := range t.Images {
				platform := fmt.Sprintf("%s/%s", image.Os, image.Architecture)
				if image.Variant != "" {
					platform += "/" + image.Variant
				}
				platforms = append(platforms, platform)
			}
			return strings.Join(platforms, ",")
		},
	}
)

type column struct {
	header string
	value  func(t hub.Tag) string
}

type listOptions struct {
	platforms bool
}

func newListCmd(ctx context.Context, dockerCli command.Cli) *cobra.Command {
	var opts listOptions
	cmd := &cobra.Command{
		Use:   "ls [OPTION] REPOSITORY",
		Short: "List all the images in a repository",
		Args:  cli.ExactArgs(1),
		RunE: func(_ *cobra.Command, args []string) error {
			return runList(ctx, dockerCli, opts, args[0])
		},
	}
	cmd.Flags().BoolVar(&opts.platforms, "platforms", false, "List all available platforms per tag")
	return cmd
}

func runList(ctx context.Context, dockerCli command.Cli, opts listOptions, repository string) error {
	authResolver := func(hub *registry.IndexInfo) types.AuthConfig {
		return command.ResolveAuthConfig(ctx, dockerCli, hub)
	}
	client, err := hub.NewClient(authResolver)
	if err != nil {
		return err
	}
	tags, err := client.GetTags(repository)
	if err != nil {
		return err
	}

	if opts.platforms {
		defaultColumns = append(defaultColumns, platformColumn)
	}

	w := tabwriter.NewWriter(os.Stdout, 20, 1, 3, ' ', 0)
	var headers []string
	for _, column := range defaultColumns {
		headers = append(headers, column.header)
	}
	fmt.Fprintln(w, strings.Join(headers, "\t"))

	for _, tag := range tags {
		var values []string
		for _, column := range defaultColumns {
			values = append(values, column.value(tag))
		}
		fmt.Fprintln(w, strings.Join(values, "\t"))
	}
	return w.Flush()
}
