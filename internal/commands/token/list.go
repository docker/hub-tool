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

package token

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
	lsName = "ls"
)

var (
	defaultColumns = []column{
		{"DESCRIPTION", func(t hub.Token) (string, int) { return t.Description, len(t.Description) }},
		{"UUID", func(t hub.Token) (string, int) { return t.UUID.String(), len(t.UUID.String()) }},
		{"LAST USED", func(t hub.Token) (string, int) {
			s := "Never"
			if !t.LastUsed.IsZero() {
				s = fmt.Sprintf("%s ago", units.HumanDuration(time.Since(t.LastUsed)))
			}
			return s, len(s)
		}},
		{"CREATED", func(t hub.Token) (string, int) {
			s := units.HumanDuration(time.Since(t.CreatedAt))
			return s, len(s)
		}},
		{"ACTIVE", func(t hub.Token) (string, int) {
			s := fmt.Sprintf("%v", t.IsActive)
			return s, len(s)
		}},
		{"SCOPE", func(t hub.Token) (string, int) {
			if len(t.Scopes) == 0 {
				return "", 0
			}
			return t.Scopes[0][5:], len(t.Scopes[0]) - 5
		}},
	}
)

type column struct {
	header string
	value  func(t hub.Token) (string, int)
}

type listOptions struct {
	format.Option
	all bool
}

func newListCmd(streams command.Streams, hubClient *hub.Client, parent string) *cobra.Command {
	var opts listOptions
	cmd := &cobra.Command{
		Use:                   lsName + " [OPTION]",
		Aliases:               []string{"list"},
		Short:                 "List all the Personal Access Tokens",
		Args:                  cli.NoArgs,
		DisableFlagsInUseLine: true,
		Annotations: map[string]string{
			"sudo": "true",
		},
		PreRun: func(cmd *cobra.Command, args []string) {
			metrics.Send(parent, lsName)
		},
		RunE: func(_ *cobra.Command, args []string) error {
			return runList(streams, hubClient, opts)
		},
	}
	cmd.Flags().BoolVar(&opts.all, "all", false, "Fetch all available tokens")
	opts.AddFormatFlag(cmd.Flags())
	return cmd
}

func runList(streams command.Streams, hubClient *hub.Client, opts listOptions) error {
	if opts.all {
		if err := hubClient.Update(hub.WithAllElements()); err != nil {
			return err
		}
	}
	tokens, total, err := hubClient.GetTokens()
	if err != nil {
		return err
	}
	return opts.Print(streams.Out(), tokens, printTokens(total))
}

func printTokens(total int) format.PrettyPrinter {
	return func(out io.Writer, values interface{}) error {
		tokens := values.([]hub.Token)
		tw := tabwriter.New(out, "    ")
		for _, column := range defaultColumns {
			tw.Column(ansi.Header(column.header), len(column.header))
		}

		tw.Line()
		for _, token := range tokens {
			for _, column := range defaultColumns {
				value, width := column.value(token)
				tw.Column(value, width)
			}
			tw.Line()
		}
		if err := tw.Flush(); err != nil {
			return err
		}

		if len(tokens) < total {
			fmt.Fprintln(out, ansi.Info(fmt.Sprintf("%v/%v listed, use --all flag to show all", len(tokens), total)))
		}
		return nil
	}
}
