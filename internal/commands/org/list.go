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

package org

import (
	"context"
	"fmt"
	"io"

	"github.com/docker/cli/cli"
	"github.com/docker/cli/cli/command"
	"github.com/spf13/cobra"

	"github.com/docker/hub-tool/internal/ansi"
	"github.com/docker/hub-tool/internal/format"
	"github.com/docker/hub-tool/internal/format/tabwriter"
	"github.com/docker/hub-tool/pkg/hub"
	"github.com/docker/hub-tool/internal/metrics"
)

const (
	listName = "ls"
)

var (
	defaultColumns = []column{
		{"NAMESPACE", func(o hub.Organization) (string, int) {
			return ansi.Link(fmt.Sprintf("https://hub.docker.com/orgs/%s", o.Namespace), o.Namespace), len(o.Namespace)
		}},
		{"NAME", func(o hub.Organization) (string, int) { return o.FullName, len(o.FullName) }},
		{"MY ROLE", func(o hub.Organization) (string, int) { return o.Role, len(o.Role) }},
		{"TEAMS", func(o hub.Organization) (string, int) {
			s := fmt.Sprintf("%v", len(o.Teams))
			return s, len(s)
		}},
		{"MEMBERS", func(o hub.Organization) (string, int) {
			s := fmt.Sprintf("%v", len(o.Members))
			return s, len(s)
		}},
	}
)

type column struct {
	header string
	value  func(o hub.Organization) (string, int)
}

type listOptions struct {
	format.Option
}

func newListCmd(streams command.Streams, hubClient *hub.Client, parent string) *cobra.Command {
	var opts listOptions
	cmd := &cobra.Command{
		Use:                   listName,
		Aliases:               []string{"list"},
		Short:                 "List all the organizations",
		Args:                  cli.NoArgs,
		DisableFlagsInUseLine: true,
		PreRun: func(cmd *cobra.Command, args []string) {
			metrics.Send(parent, listName)
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return runList(cmd.Context(), streams, hubClient, opts)
		},
	}
	opts.AddFormatFlag(cmd.Flags())
	return cmd
}

func runList(ctx context.Context, streams command.Streams, hubClient *hub.Client, opts listOptions) error {
	organizations, err := hubClient.GetOrganizations(ctx)
	if err != nil {
		return err
	}
	return opts.Print(streams.Out(), organizations, printOrganizations)
}

func printOrganizations(out io.Writer, values interface{}) error {
	organizations := values.([]hub.Organization)

	tw := tabwriter.New(out, "    ")

	for _, c := range defaultColumns {
		tw.Column(ansi.Header(c.header), len(c.header))
	}

	tw.Line()

	for _, org := range organizations {
		for _, column := range defaultColumns {
			value, width := column.value(org)
			tw.Column(value, width)
		}
		tw.Line()
	}

	return tw.Flush()
}
