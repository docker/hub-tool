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

package account

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
	"github.com/docker/hub-tool/internal/metrics"
	"github.com/docker/hub-tool/pkg/hub"
)

const (
	rateLimitingName = "rate-limiting"
)

type rateLimitingOptions struct {
	format.Option
}

func newRateLimitingCmd(streams command.Streams, hubClient *hub.Client, parent string) *cobra.Command {
	var opts rateLimitingOptions
	cmd := &cobra.Command{
		Use:                   rateLimitingName,
		Short:                 "Print the rate limiting information",
		Args:                  cli.NoArgs,
		DisableFlagsInUseLine: true,
		Annotations: map[string]string{
			"sudo": "true",
		},
		PreRun: func(cmd *cobra.Command, args []string) {
			metrics.Send(parent, rateLimitingName)
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return runRateLimiting(streams, hubClient, opts)
		},
	}
	opts.AddFormatFlag(cmd.Flags())

	return cmd
}

func runRateLimiting(streams command.Streams, hubClient *hub.Client, opts rateLimitingOptions) error {
	rl, err := hubClient.GetRateLimits()
	if err != nil {
		return err
	}
	value := &hub.RateLimits{}
	if rl != nil {
		value = rl
	}
	return opts.Print(streams.Out(), value, printRateLimit(rl))
}

func printRateLimit(rl *hub.RateLimits) func(io.Writer, interface{}) error {
	return func(out io.Writer, _ interface{}) error {
		if rl == nil {
			fmt.Fprintln(out, ansi.Emphasise("Unlimited"))
			return nil
		}
		color := ansi.NoColor
		if *rl.Remaining <= 50 {
			color = ansi.Warn
		}
		if *rl.Remaining <= 10 {
			color = ansi.Error
		}
		fmt.Fprintf(out, color("Limit:     %d, %s window\n"), *rl.Limit, units.HumanDuration(time.Duration(*rl.LimitWindow)*time.Second))
		fmt.Fprintf(out, color("Remaining: %d, %s window\n"), *rl.Remaining, units.HumanDuration(time.Duration(*rl.RemainingWindow)*time.Second))
		return nil
	}
}
