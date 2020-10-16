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
	"bufio"
	"fmt"
	"strings"

	"github.com/docker/cli/cli"
	"github.com/docker/cli/cli/command"
	"github.com/google/uuid"
	"github.com/spf13/cobra"

	"github.com/docker/hub-tool/internal/color"
	"github.com/docker/hub-tool/internal/hub"
	"github.com/docker/hub-tool/internal/metrics"
)

const (
	revokeName = "revoke"
)

type revokeOptions struct {
	force bool
}

func newRevokeCmd(streams command.Streams, hubClient *hub.Client, parent string) *cobra.Command {
	var opts revokeOptions
	cmd := &cobra.Command{
		Use:   revokeName + " [OPTIONS] TOKEN_UUID",
		Short: "Revoke a Personal Access Token",
		Args:  cli.ExactArgs(1),
		PreRun: func(cmd *cobra.Command, args []string) {
			metrics.Send(parent, revokeName)
		},
		RunE: func(_ *cobra.Command, args []string) error {
			return runRevoke(streams, hubClient, opts, args[0])
		},
	}
	cmd.Flags().BoolVar(&opts.force, "force", false, "Force deletion of the tag")
	cmd.Flags().SetInterspersed(false)
	return cmd
}

func runRevoke(streams command.Streams, hubClient *hub.Client, opts revokeOptions, tokenUUID string) error {
	u, err := uuid.Parse(tokenUUID)
	if err != nil {
		return err
	}

	if !opts.force {
		fmt.Fprintf(streams.Out(), color.Warn("WARNING: This action is irreversible.")+`
By confirming, you will permanently delete the access token.
Revoking a token will invalidate your credentials on all Docker clients currently authenticated with this token.

Please type your username %q to confirm deletion: `, hubClient.AuthConfig.Username)
		reader := bufio.NewReader(streams.In())
		input, _ := reader.ReadString('\n')
		input = strings.ToLower(strings.TrimSpace(input))
		if input != hubClient.AuthConfig.Username {
			return fmt.Errorf("%q differs from your username, deletion aborted", input)
		}
	}

	if err := hubClient.RevokeToken(u.String()); err != nil {
		return err
	}
	fmt.Fprintln(streams.Out(), color.Emphasise("Revoked"), u)
	return nil
}
