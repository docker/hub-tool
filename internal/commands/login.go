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

package commands

import (
	"fmt"

	"github.com/docker/cli/cli"
	"github.com/docker/cli/cli/command"
	"github.com/spf13/cobra"

	"github.com/docker/hub-tool/internal/ansi"
	"github.com/docker/hub-tool/internal/credentials"
	"github.com/docker/hub-tool/internal/hub"
	"github.com/docker/hub-tool/internal/login"
	"github.com/docker/hub-tool/internal/metrics"
)

const (
	loginName = "login"
)

func newLoginCmd(streams command.Streams, store credentials.Store, hubClient *hub.Client) *cobra.Command {
	cmd := &cobra.Command{
		Use:                   loginName + " USERNAME",
		Short:                 "Login to the Hub",
		Args:                  cli.ExactArgs(1),
		DisableFlagsInUseLine: true,
		PreRun: func(cmd *cobra.Command, args []string) {
			metrics.Send("root", loginName)
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := login.RunLogin(cmd.Context(), streams, hubClient, store, args[0]); err != nil {
				return err
			}
			fmt.Fprintln(streams.Out(), ansi.Info("Login Succeeded"))
			return nil
		},
	}
	return cmd
}
