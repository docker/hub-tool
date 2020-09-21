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

package commands

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
	listColumns = []struct {
		header string
		value  func(t hub.Tag) string
	}{
		{"TAG", func(t hub.Tag) string { return t.Name }},
		{"DIGEST", func(t hub.Tag) string { return t.Images[0].Digest }},
		{"LAST UPDATE", func(t hub.Tag) string { return fmt.Sprintf("%s ago", units.HumanDuration(time.Since(t.LastUpdated))) }},
		{"SIZE", func(t hub.Tag) string { return units.HumanSize(float64(t.Images[0].Size)) }},
	}
)

func newListCmd(ctx context.Context, dockerCli command.Cli) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "ls NAME",
		Short: "List all the images in a repository",
		Args:  cli.ExactArgs(1),
		RunE: func(_ *cobra.Command, args []string) error {
			return runList(ctx, dockerCli, args)
		},
	}
	return cmd
}

func runList(ctx context.Context, dockerCli command.Cli, args []string) error {
	authResolver := func(hub *registry.IndexInfo) types.AuthConfig {
		return command.ResolveAuthConfig(ctx, dockerCli, hub)
	}
	client, err := hub.NewClient(authResolver)
	if err != nil {
		return err
	}
	tags, err := client.GetTags(args[0])
	if err != nil {
		return err
	}

	w := tabwriter.NewWriter(os.Stdout, 20, 1, 3, ' ', 0)
	var headers []string
	for _, column := range listColumns {
		headers = append(headers, column.header)
	}
	fmt.Fprintln(w, strings.Join(headers, "\t"))

	for _, tag := range tags {
		var values []string
		for _, column := range listColumns {
			values = append(values, column.value(tag))
		}
		fmt.Fprintln(w, strings.Join(values, "\t"))
	}
	return w.Flush()
}
