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

package repo

import (
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/docker/cli/cli"
	"github.com/docker/cli/cli/command"
	"github.com/docker/go-units"
	"github.com/spf13/cobra"

	"github.com/docker/hub-tool/internal/ansi"
	"github.com/docker/hub-tool/internal/format"
	"github.com/docker/hub-tool/internal/format/tabwriter"
	"github.com/docker/hub-tool/internal/hub"
	"github.com/docker/hub-tool/internal/metrics"
)

const (
	listName = "ls"
)

var (
	defaultColumns = []column{
		{"REPOSITORY", func(r hub.Repository) (string, int) {
			return ansi.Link(fmt.Sprintf("https://hub.docker.com/repository/docker/%s", r.Name), r.Name), len(r.Name)
		}},
		{"DESCRIPTION", func(r hub.Repository) (string, int) {
			return strings.TrimSuffix(r.Description, "\n"), len(r.Description)
		}},
		{"LAST UPDATE", func(r hub.Repository) (string, int) {
			if r.LastUpdated.Nanosecond() == 0 {
				return "", 0
			}
			s := fmt.Sprintf("%s ago", units.HumanDuration(time.Since(r.LastUpdated)))
			return s, len(s)
		}},
		{"PULLS", func(r hub.Repository) (string, int) {
			s := fmt.Sprintf("%d", r.PullCount)
			return s, len(s)
		}},
		{"STARS", func(r hub.Repository) (string, int) {
			s := fmt.Sprintf("%d", r.StarCount)
			return s, len(s)
		}},
		{"PRIVATE", func(r hub.Repository) (string, int) {
			s := fmt.Sprintf("%v", r.IsPrivate)
			return s, len(s)
		}},
	}
)

type column struct {
	header string
	value  func(t hub.Repository) (string, int)
}

type listOptions struct {
	format.Option
}

func newListCmd(streams command.Streams, hubClient *hub.Client, parent string) *cobra.Command {
	var opts listOptions
	cmd := &cobra.Command{
		Use:                   listName + " [ORGANIZATION]",
		Short:                 "List all the repositories from your account or an organization",
		Args:                  cli.RequiresMaxArgs(1),
		DisableFlagsInUseLine: true,
		PreRun: func(cmd *cobra.Command, args []string) {
			metrics.Send(parent, listName)
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return runList(streams, hubClient, opts, args)
		},
	}
	opts.AddFormatFlag(cmd.Flags())
	return cmd
}

func runList(streams command.Streams, hubClient *hub.Client, opts listOptions, args []string) error {
	account := hubClient.AuthConfig.Username
	if len(args) > 0 {
		account = args[0]
	}

	repositories, err := hubClient.GetRepositories(account)
	if err != nil {
		return err
	}

	return opts.Print(streams.Out(), repositories, printRepositories)
}

func printRepositories(out io.Writer, values interface{}) error {
	tw := tabwriter.New(out, "    ")

	for _, column := range defaultColumns {
		tw.Column(ansi.Header(column.header), len(column.header))
	}

	tw.Line()

	repositories := values.([]hub.Repository)

	for _, repository := range repositories {
		for _, column := range defaultColumns {
			value, width := column.value(repository)
			tw.Column(value, width)
		}
		tw.Line()
	}
	if err := tw.Flush(); err != nil {
		return err
	}

	return nil
}
