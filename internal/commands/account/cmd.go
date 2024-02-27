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
	"github.com/docker/cli/cli"
	"github.com/docker/cli/cli/command"
	"github.com/spf13/cobra"

	"github.com/docker/hub-tool/pkg/hub"
)

const (
	accountName = "account"
)

// NewAccountCmd configures the org manage command
func NewAccountCmd(streams command.Streams, hubClient *hub.Client) *cobra.Command {
	cmd := &cobra.Command{
		Use:                   accountName,
		Short:                 "Manage your account",
		Args:                  cli.NoArgs,
		DisableFlagsInUseLine: true,
		Annotations: map[string]string{
			"sudo": "true",
		},
		RunE: command.ShowHelp(streams.Err()),
	}
	cmd.AddCommand(
		newInfoCmd(streams, hubClient, accountName),
		newRateLimitingCmd(streams, hubClient, accountName),
	)
	return cmd
}
