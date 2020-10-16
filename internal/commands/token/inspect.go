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
	"github.com/google/uuid"
	"github.com/juju/ansiterm"
	"github.com/spf13/cobra"

	"github.com/docker/hub-tool/internal/color"
	"github.com/docker/hub-tool/internal/format"
	"github.com/docker/hub-tool/internal/hub"
	"github.com/docker/hub-tool/internal/metrics"
)

const (
	inspectName = "inspect"
)

type inspectOptions struct {
	format.Option
}

func newInspectCmd(streams command.Streams, hubClient *hub.Client, parent string) *cobra.Command {
	var opts inspectOptions
	cmd := &cobra.Command{
		Use:   inspectName + " [OPTIONS] TOKEN_UUID",
		Short: "Inspect a Personal Access Token",
		Args:  cli.ExactArgs(1),
		PreRun: func(cmd *cobra.Command, args []string) {
			metrics.Send(parent, inspectName)
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return runInspect(streams, hubClient, opts, args[0])
		},
	}
	opts.AddFormatFlag(cmd.Flags())
	cmd.Flags().SetInterspersed(false)
	return cmd
}

func runInspect(streams command.Streams, hubClient *hub.Client, opts inspectOptions, tokenUUID string) error {
	u, err := uuid.Parse(tokenUUID)
	if err != nil {
		return err
	}
	token, err := hubClient.GetToken(u.String())
	if err != nil {
		return err
	}
	return opts.Print(streams.Out(), token, printInspectToken)
}

const (
	prefix = "  "
)

func printInspectToken(out io.Writer, value interface{}) error {
	token := value.(*hub.Token)

	w := ansiterm.NewTabWriter(out, 0, 0, 1, ' ', 0)
	fmt.Fprintf(w, color.Title("Token:")+"\n")
	fmt.Fprintf(w, color.Key("%sUUID:")+"\t%s\n", prefix, token.UUID)
	if token.Description != "" {
		fmt.Fprintf(w, color.Key("%sDescription:")+"\t%s\n", prefix, token.Description)
	}
	fmt.Fprintf(w, color.Key("%sIs Active:")+"\t%v\n", prefix, token.IsActive)
	if len(token.Scopes) > 0 {
		fmt.Fprintf(w, color.Key("%sScopes:")+"\t%s\n", prefix, strings.Join(token.Scopes, ", "))
	}
	fmt.Fprintf(w, color.Key("%sCreated:")+"\t%s\n", prefix, fmt.Sprintf("%s ago", units.HumanDuration(time.Since(token.CreatedAt))))
	fmt.Fprintf(w, color.Key("%sLast Used:")+"\t%s\n", prefix, getLastUsed(token.LastUsed))
	fmt.Fprintf(w, color.Key("%sCreator User Agent:")+"\t%s\n", prefix, token.CreatorUA)
	fmt.Fprintf(w, color.Key("%sCreator IP:")+"\t%s\n", prefix, token.CreatorIP)
	fmt.Fprintf(w, color.Key("%sGenerated:")+"\t%s\n", prefix, getGeneratedBy(token))
	return w.Flush()
}

func getLastUsed(t time.Time) string {
	if t.IsZero() {
		return "Never"
	}
	return fmt.Sprintf("%s ago", units.HumanDuration(time.Since(t)))
}

func getGeneratedBy(token *hub.Token) string {
	if strings.Contains(token.CreatorUA, "hub-tool") {
		return "By hub-tool"
	}
	return "By user via Web UI"
}
