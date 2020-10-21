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
	"strings"
	"time"

	"github.com/docker/cli/cli"
	"github.com/docker/cli/cli/command"
	"github.com/docker/go-units"
	"github.com/juju/ansiterm"
	"github.com/spf13/cobra"

	"github.com/docker/hub-tool/internal/ansi"
	"github.com/docker/hub-tool/internal/format"
	"github.com/docker/hub-tool/internal/hub"
	"github.com/docker/hub-tool/internal/metrics"
)

const (
	lsName = "ls"
)

var (
	defaultColumns = []column{
		{"DESCRIPTION", func(t hub.Token) string { return t.Description }},
		{"UUID", func(t hub.Token) string { return t.UUID.String() }},
		{"LAST USED", func(t hub.Token) string {
			if t.LastUsed.IsZero() {
				return "Never"
			}
			return fmt.Sprintf("%s ago", units.HumanDuration(time.Since(t.LastUsed)))
		}},
		{"CREATED", func(t hub.Token) string {
			return units.HumanDuration(time.Since(t.CreatedAt))
		}},
		{"ACTIVE", func(t hub.Token) string {
			return fmt.Sprintf("%v", t.IsActive)
		}},
	}
)

type column struct {
	header string
	value  func(t hub.Token) string
}

type listOptions struct {
	format.Option
	all bool
}

func newListCmd(streams command.Streams, hubClient *hub.Client, parent string) *cobra.Command {
	var opts listOptions
	cmd := &cobra.Command{
		Use:   lsName + " [OPTION]",
		Short: "List all the Personal Access Tokens",
		Args:  cli.NoArgs,
		PreRun: func(cmd *cobra.Command, args []string) {
			metrics.Send(parent, lsName)
		},
		RunE: func(_ *cobra.Command, args []string) error {
			return runList(streams, hubClient, opts)
		},
	}
	cmd.Flags().BoolVar(&opts.all, "all", false, "Fetch all available tokens")
	opts.AddFormatFlag(cmd.Flags())
	cmd.Flags().SetInterspersed(false)
	return cmd
}

func runList(streams command.Streams, hubClient *hub.Client, opts listOptions) error {
	if opts.all {
		if err := hubClient.Apply(hub.WithAllElements()); err != nil {
			return err
		}
	}
	tokens, total, err := hubClient.GetTokens()
	if err != nil {
		return err
	}
	return opts.Print(streams.Out(), &helper{tokens, total}, printTokens)
}

func printTokens(out io.Writer, values interface{}) error {
	h := values.(*helper)
	w := ansiterm.NewTabWriter(out, 20, 1, 3, ' ', 0)
	var headers []string
	for _, column := range defaultColumns {
		headers = append(headers, column.header)
	}
	fmt.Fprintln(w, ansi.Header(strings.Join(headers, "\t")))
	for _, token := range h.tokens {
		var values []string
		for _, column := range defaultColumns {
			values = append(values, column.value(token))
		}
		fmt.Fprintln(w, strings.Join(values, "\t"))
	}
	if err := w.Flush(); err != nil {
		return err
	}

	if len(h.tokens) < h.total {
		fmt.Fprintln(out, ansi.Info(fmt.Sprintf("%v/%v listed, use --all flag to show all", len(h.tokens), h.total)))
	}
	return nil
}

type helper struct {
	tokens []hub.Token
	total  int
}
