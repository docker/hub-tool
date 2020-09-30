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

package repo

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
	"github.com/docker/hub-cli-plugin/internal/metrics"
)

const (
	listName = "ls"
)

var (
	defaultColumns = []column{
		{"REPOSITORY", func(r hub.Repository) string { return r.Name }},
		{"DESCRIPTION", func(r hub.Repository) string { return r.Description }},
		{"LAST UPDATE", func(r hub.Repository) string {
			if r.LastUpdated.Nanosecond() == 0 {
				return ""
			}
			return fmt.Sprintf("%s ago", units.HumanDuration(time.Since(r.LastUpdated)))
		}},
		{"PULLS", func(r hub.Repository) string { return fmt.Sprintf("%v", r.PullCount) }},
		{"STARS", func(r hub.Repository) string { return fmt.Sprintf("%v", r.StarCount) }},
		{"PRIVATE", func(r hub.Repository) string { return fmt.Sprintf("%v", r.IsPrivate) }},
	}
)

type column struct {
	header string
	value  func(t hub.Repository) string
}

func newListCmd(ctx context.Context, dockerCli command.Cli, parent string) *cobra.Command {
	cmd := &cobra.Command{
		Use:   listName + " [ORGANIZATION]",
		Short: "List all the repositories from your account or an organization",
		Args:  cli.RequiresMaxArgs(1),
		PreRun: func(cmd *cobra.Command, args []string) {
			metrics.Send(parent, listName)
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return runList(ctx, dockerCli, args)
		},
	}
	return cmd
}

func runList(ctx context.Context, dockerCli command.Cli, args []string) error {
	var account string
	authResolver := func(hub *registry.IndexInfo) types.AuthConfig {
		authConfig := command.ResolveAuthConfig(ctx, dockerCli, hub)
		account = authConfig.Username
		return authConfig
	}
	client, err := hub.NewClient(authResolver)
	if err != nil {
		return err
	}
	if len(args) > 0 {
		account = args[0]
	}
	repositories, err := client.GetRepositories(account)
	if err != nil {
		return err
	}

	w := tabwriter.NewWriter(os.Stdout, 20, 1, 3, ' ', 0)
	var headers []string
	for _, column := range defaultColumns {
		headers = append(headers, column.header)
	}
	fmt.Fprintln(w, strings.Join(headers, "\t"))

	for _, repository := range repositories {
		var values []string
		for _, column := range defaultColumns {
			values = append(values, column.value(repository))
		}
		fmt.Fprintln(w, strings.Join(values, "\t"))
	}
	return w.Flush()
}
