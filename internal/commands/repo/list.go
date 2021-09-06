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
	"time"

	"github.com/docker/cli/cli"
	"github.com/docker/cli/cli/command"
	"github.com/docker/go-units"
	"github.com/spf13/cobra"

	"github.com/docker/hub-tool/internal/ansi"
	"github.com/docker/hub-tool/internal/format"
	"github.com/docker/hub-tool/internal/format/tabwriter"
	"github.com/docker/hub-tool/internal/metrics"
	"github.com/docker/hub-tool/pkg/hub"
)

const (
	listName = "ls"
)

var (
	defaultColumns = []column{
		{"REPOSITORY", func(r hub.Repository) (string, int) {
			return ansi.Link(fmt.Sprintf("https://hub.docker.com/repository/docker/%s", r.Name), r.Name), len(r.Name)
		}},
		{"DESCRIPTION", func(r hub.Repository) (string, int) { return r.Description, len(r.Description) }},
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
	all bool
}

func newListCmd(streams command.Streams, hubClient *hub.Client, parent string) *cobra.Command {
	var opts listOptions
	cmd := &cobra.Command{
		Use:                   listName + " [OPTIONS] [ORGANIZATION]",
		Aliases:               []string{"list"},
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
	cmd.Flags().BoolVar(&opts.all, "all", false, "Fetch all available repositories")
	opts.AddFormatFlag(cmd.Flags())
	return cmd
}

func runList(streams command.Streams, hubClient *hub.Client, opts listOptions, args []string) error {
	account := hubClient.AuthConfig.Username
	if opts.all {
		if err := hubClient.Update(hub.WithAllElements()); err != nil {
			return err
		}
	}
	if len(args) > 0 {
		account = args[0]
	}
	repositories, total, err := hubClient.GetRepositories(account)
	if err != nil {
		return err
	}

	return opts.Print(streams.Out(), repositories, printRepositories(total))
}

func printRepositories(total int) format.PrettyPrinter {
	return func(out io.Writer, values interface{}) error {
		repositories := values.([]hub.Repository)
		tw := tabwriter.New(out, "    ")

		for _, column := range defaultColumns {
			tw.Column(ansi.Header(column.header), len(column.header))
		}

		tw.Line()

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

		if len(repositories) < total {
			fmt.Fprintln(out, ansi.Info(fmt.Sprintf("%v/%v listed, use --all flag to show all", len(repositories), total)))
		}
		return nil
	}
}
