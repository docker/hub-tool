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
	"text/tabwriter"
	"time"

	"github.com/docker/cli/cli"
	"github.com/docker/cli/cli/command"
	"github.com/docker/go-units"
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

type listOptions struct {
	format.Option
	all bool
}

func newListCmd(streams command.Streams, hubClient *hub.Client, parent string) *cobra.Command {
	var opts listOptions
	cmd := &cobra.Command{
		Use:   listName + " [ORGANIZATION]",
		Short: "List all the repositories from your account or an organization",
		Args:  cli.RequiresMaxArgs(1),
		PreRun: func(cmd *cobra.Command, args []string) {
			metrics.Send(parent, listName)
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return runList(streams, hubClient, opts, args)
		},
	}
	cmd.Flags().BoolVar(&opts.all, "all", false, "Fetch all available repositories")
	opts.AddFormatFlag(cmd.Flags())
	cmd.Flags().SetInterspersed(false)
	return cmd
}

func runList(streams command.Streams, hubClient *hub.Client, opts listOptions, args []string) error {
	account := hubClient.AuthConfig.Username
	if opts.all {
		if err := hubClient.Apply(hub.WithAllElements()); err != nil {
			return err
		}
	}
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
	repositories := values.([]hub.Repository)
	w := tabwriter.NewWriter(out, 20, 1, 3, ' ', 0)
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
