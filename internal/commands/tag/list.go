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

package tag

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
	"github.com/docker/hub-tool/pkg/hub"
	"github.com/docker/hub-tool/internal/metrics"
)

const (
	lsName = "ls"
)

var (
	defaultColumns = []column{
		{"TAG", func(t hub.Tag) (string, int) { return t.Name, len(t.Name) }},
		{"DIGEST", func(t hub.Tag) (string, int) {
			if len(t.Images) > 0 {
				return t.Images[0].Digest, len(t.Images[0].Digest)
			}
			return "", 0
		}},
		{"STATUS", func(t hub.Tag) (string, int) {
			return t.Status, len(t.Status)
		}},
		{"LAST UPDATE", func(t hub.Tag) (string, int) {
			if t.LastUpdated.Nanosecond() == 0 {
				return "", 0
			}
			s := fmt.Sprintf("%s ago", units.HumanDuration(time.Since(t.LastUpdated)))
			return s, len(s)
		}},
		{"LAST PUSHED", func(t hub.Tag) (string, int) {
			if t.LastPushed.Nanosecond() == 0 {
				return "", 0
			}
			s := units.HumanDuration(time.Since(t.LastPushed))
			return s, len(s)
		}},
		{"LAST PULLED", func(t hub.Tag) (string, int) {
			if t.LastPulled.Nanosecond() == 0 {
				return "", 0
			}
			s := units.HumanDuration(time.Since(t.LastPulled))
			return s, len(s)
		}},
		{"SIZE", func(t hub.Tag) (string, int) {
			size := t.FullSize
			if len(t.Images) > 0 {
				size = 0
				for _, image := range t.Images {
					size += image.Size
				}
			}
			s := units.HumanSize(float64(size))
			return s, len(s)
		}},
	}
	platformColumn = column{
		"OS/ARCH",
		func(t hub.Tag) (string, int) {
			var platforms []string
			for _, image := range t.Images {
				platform := fmt.Sprintf("%s/%s", image.Os, image.Architecture)
				if image.Variant != "" {
					platform += "/" + image.Variant
				}
				platforms = append(platforms, platform)
			}
			s := strings.Join(platforms, ",")
			return s, len(s)
		},
	}
)

type column struct {
	header string
	value  func(t hub.Tag) (string, int)
}

type listOptions struct {
	format.Option
	platforms bool
	all       bool
	sort      string
}

func newListCmd(streams command.Streams, hubClient *hub.Client, parent string) *cobra.Command {
	var opts listOptions
	cmd := &cobra.Command{
		Use:                   lsName + " [OPTION] REPOSITORY",
		Aliases:               []string{"list"},
		Short:                 "List all the images in a repository",
		Args:                  cli.ExactArgs(1),
		DisableFlagsInUseLine: true,
		PreRun: func(cmd *cobra.Command, args []string) {
			metrics.Send(parent, lsName)
		},
		RunE: func(_ *cobra.Command, args []string) error {
			return runList(streams, hubClient, opts, args[0])
		},
	}
	cmd.Flags().BoolVar(&opts.platforms, "platforms", false, "List all available platforms per tag")
	cmd.Flags().BoolVar(&opts.all, "all", false, "Fetch all available tags")
	cmd.Flags().StringVar(&opts.sort, "sort", "", "Sort tags by (updated|name)[=(asc|desc)] (e.g.: --sort updated or --sort name=desc)")
	opts.AddFormatFlag(cmd.Flags())
	return cmd
}

func runList(streams command.Streams, hubClient *hub.Client, opts listOptions, repository string) error {
	ordering, err := mapOrdering(opts.sort)
	if err != nil {
		return err
	}
	if opts.all {
		if err := hubClient.Update(hub.WithAllElements()); err != nil {
			return err
		}
	}

	var reqOps []hub.RequestOp
	if ordering != "" {
		reqOps = append(reqOps, hub.WithSortingOrder(ordering))
	}
	tags, total, err := hubClient.GetTags(repository, reqOps...)
	if err != nil {
		return err
	}

	if opts.platforms {
		defaultColumns = append(defaultColumns, platformColumn)
	}

	return opts.Print(streams.Out(), tags, printTags(total))
}

func printTags(total int) format.PrettyPrinter {
	return func(out io.Writer, values interface{}) error {
		tags := values.([]hub.Tag)
		tw := tabwriter.New(out, "    ")
		for _, column := range defaultColumns {
			tw.Column(ansi.Header(column.header), len(column.header))
		}

		tw.Line()

		for _, tag := range tags {
			for _, column := range defaultColumns {
				value, width := column.value(tag)
				tw.Column(value, width)
			}
			tw.Line()
		}
		if err := tw.Flush(); err != nil {
			return err
		}

		if len(tags) < total {
			fmt.Fprintln(out, ansi.Info(fmt.Sprintf("%v/%v listed, use --all flag to show all", len(tags), total)))
		}

		return nil
	}
}

const (
	sortAsc  = "asc"
	sortDesc = "desc"
)

func mapOrdering(order string) (string, error) {
	if order == "" {
		return "", nil
	}
	name := "-name"
	update := "last_updated"
	fields := strings.SplitN(order, "=", 2)
	if len(fields) == 2 {
		switch fields[1] {
		case sortDesc:
			name = "name"
			update = "-last_updated"
		case sortAsc:
		default:
			return "", fmt.Errorf(`invalid sorting direction %q: should be either "asc" or "desc"`, fields[1])
		}
	}
	switch fields[0] {
	case "updated":
		return update, nil
	case "name":
		return name, nil
	default:
		return "", fmt.Errorf(`unknown sorting column %q: should be either "name" or "updated"`, fields[0])
	}
}
