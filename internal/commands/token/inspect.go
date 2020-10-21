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
	"github.com/spf13/cobra"

	"github.com/docker/hub-tool/internal/ansi"
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

func printInspectToken(out io.Writer, value interface{}) error {
	token := value.(*hub.Token)

	fmt.Fprintf(out, ansi.Title("Token:")+"\n")
	fmt.Fprintf(out, ansi.Key("UUID:")+"\t%s\n", token.UUID)
	if token.Description != "" {
		fmt.Fprintf(out, ansi.Key("Description:")+"\t%s\n", token.Description)
	}
	fmt.Fprintf(out, ansi.Key("Is Active:")+"\t%v\n", token.IsActive)
	fmt.Fprintf(out, ansi.Key("Created:")+"\t%s\n", fmt.Sprintf("%s ago", units.HumanDuration(time.Since(token.CreatedAt))))
	fmt.Fprintf(out, ansi.Key("Last Used:")+"\t%s\n", getLastUsed(token.LastUsed))
	fmt.Fprintf(out, ansi.Key("Creator User Agent:")+"\t%s\n", token.CreatorUA)
	fmt.Fprintf(out, ansi.Key("Creator IP:")+"\t%s\n", token.CreatorIP)
	fmt.Fprintf(out, ansi.Key("Generated:")+"\t%s\n", getGeneratedBy(token))
	return nil
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
