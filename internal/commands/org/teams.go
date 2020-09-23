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

package org

import (
	"context"
	"fmt"
	"os"
	"strings"
	"text/tabwriter"

	"github.com/docker/cli/cli"
	"github.com/docker/cli/cli/command"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/registry"
	"github.com/spf13/cobra"

	"github.com/docker/hub-cli-plugin/internal/hub"
)

var (
	teamsColumns = []teamColumn{
		{"TEAM", func(t hub.Team) string { return t.Name }},
		{"DESCRIPTION", func(t hub.Team) string { return t.Description }},
		{"MEMBERS", func(t hub.Team) string { return fmt.Sprintf("%v", len(t.Members)) }},
	}
)

type teamColumn struct {
	header string
	value  func(t hub.Team) string
}

func newTeamsCmd(ctx context.Context, dockerCli command.Cli) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "teams ORGANIZATION",
		Short: "List all the teams in an organization",
		Args:  cli.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runTeams(ctx, dockerCli, args[0])
		},
	}
	return cmd
}

func runTeams(ctx context.Context, dockerCli command.Cli, organization string) error {
	authResolver := func(hub *registry.IndexInfo) types.AuthConfig {
		return command.ResolveAuthConfig(ctx, dockerCli, hub)
	}
	client, err := hub.NewClient(authResolver)
	if err != nil {
		return err
	}
	teams, err := client.GetTeams(organization)
	if err != nil {
		return err
	}

	w := tabwriter.NewWriter(os.Stdout, 20, 1, 3, ' ', 0)
	var headers []string
	for _, column := range teamsColumns {
		headers = append(headers, column.header)
	}
	fmt.Fprintln(w, strings.Join(headers, "\t"))

	for _, team := range teams {
		var values []string
		for _, column := range teamsColumns {
			values = append(values, column.value(team))
		}
		fmt.Fprintln(w, strings.Join(values, "\t"))
	}
	return w.Flush()
}
