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
	"fmt"
	"io"

	"github.com/docker/cli/cli"
	"github.com/docker/cli/cli/command"
	"github.com/spf13/cobra"

	"github.com/docker/hub-tool/internal/ansi"
	"github.com/docker/hub-tool/internal/format"
	"github.com/docker/hub-tool/internal/format/tabwriter"
	"github.com/docker/hub-tool/internal/hub"
	"github.com/docker/hub-tool/internal/metrics"
)

const (
	teamsName = "teams"
)

var (
	teamsColumns = []teamColumn{
		{"TEAM", func(t hub.Team) (string, int) { return t.Name, len(t.Name) }},
		{"DESCRIPTION", func(t hub.Team) (string, int) { return t.Description, len(t.Description) }},
		{"MEMBERS", func(t hub.Team) (string, int) {
			s := fmt.Sprintf("%v", len(t.Members))
			return s, len(s)
		}},
	}
)

type teamColumn struct {
	header string
	value  func(t hub.Team) (string, int)
}

type teamsOptions struct {
	format.Option
}

func newTeamsCmd(streams command.Streams, hubClient *hub.Client, parent string) *cobra.Command {
	var opts teamsOptions
	cmd := &cobra.Command{
		Use:   teamsName + " ORGANIZATION",
		Short: "List all the teams in an organization",
		Args:  cli.ExactArgs(1),
		PreRun: func(cmd *cobra.Command, args []string) {
			metrics.Send(parent, teamsName)
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return runTeams(streams, hubClient, opts, args[0])
		},
	}
	opts.AddFormatFlag(cmd.Flags())
	cmd.Flags().SetInterspersed(false)
	return cmd
}

func runTeams(streams command.Streams, hubClient *hub.Client, opts teamsOptions, organization string) error {
	teams, err := hubClient.GetTeams(organization)
	if err != nil {
		return err
	}
	return opts.Print(streams.Out(), teams, printTeams)
}

func printTeams(out io.Writer, values interface{}) error {
	teams := values.([]hub.Team)
	tw := tabwriter.New(out, "    ")

	for _, column := range teamsColumns {
		tw.Column(ansi.Header(column.header), len(column.header))
	}

	tw.Line()
	for _, team := range teams {
		for _, column := range teamsColumns {
			value, width := column.value(team)
			tw.Column(value, width)
		}
		tw.Line()
	}

	return nil
}
