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
	"io"
	"os"
	"strings"
	"text/tabwriter"

	"github.com/docker/cli/cli"
	"github.com/docker/cli/cli/command"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/registry"
	"github.com/spf13/cobra"

	"github.com/docker/hub-cli-plugin/internal/format"
	"github.com/docker/hub-cli-plugin/internal/hub"
	"github.com/docker/hub-cli-plugin/internal/metrics"
)

const (
	listName = "ls"
)

var (
	defaultColumns = []column{
		{"NAMESPACE", func(o hub.Organization) string { return o.Namespace }},
		{"NAME", func(o hub.Organization) string { return o.FullName }},
		{"MY ROLE", func(o hub.Organization) string { return o.Role }},
		{"TEAMS", func(o hub.Organization) string { return fmt.Sprintf("%v", len(o.Teams)) }},
		{"MEMBERS", func(o hub.Organization) string { return fmt.Sprintf("%v", len(o.Members)) }},
	}
)

type column struct {
	header string
	value  func(o hub.Organization) string
}

type listOptions struct {
	format.Option
}

func newListCmd(ctx context.Context, dockerCli command.Cli, parent string) *cobra.Command {
	var opts listOptions
	cmd := &cobra.Command{
		Use:   listName,
		Short: "List all the organizations",
		Args:  cli.NoArgs,
		PreRun: func(cmd *cobra.Command, args []string) {
			metrics.Send(parent, listName)
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return runList(ctx, dockerCli, opts)
		},
	}
	opts.AddFormatFlag(cmd.Flags())
	cmd.Flags().SetInterspersed(false)
	return cmd
}

func runList(ctx context.Context, dockerCli command.Cli, opts listOptions) error {
	authResolver := func(hub *registry.IndexInfo) types.AuthConfig {
		return command.ResolveAuthConfig(ctx, dockerCli, hub)
	}
	client, err := hub.NewClient(authResolver)
	if err != nil {
		return err
	}
	organizations, err := client.GetOrganizations()
	if err != nil {
		return err
	}
	return opts.Print(os.Stdout, organizations, printOrganizations)
}

func printOrganizations(out io.Writer, values interface{}) error {
	organizations := values.([]hub.Organization)
	w := tabwriter.NewWriter(out, 20, 1, 3, ' ', 0)
	var headers []string
	for _, column := range defaultColumns {
		headers = append(headers, column.header)
	}
	fmt.Fprintln(w, strings.Join(headers, "\t"))

	for _, organization := range organizations {
		var values []string
		for _, column := range defaultColumns {
			values = append(values, column.value(organization))
		}
		fmt.Fprintln(w, strings.Join(values, "\t"))
	}
	return w.Flush()
}
