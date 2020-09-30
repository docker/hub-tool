/*
   Copyright 2020 Docker Inc.

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

package org

import (
	"context"

	"github.com/docker/cli/cli"
	"github.com/docker/cli/cli/command"
	"github.com/spf13/cobra"
)

const (
	orgName = "org"
)

//NewOrgCmd configures the org manage command
func NewOrgCmd(ctx context.Context, dockerCli command.Cli) *cobra.Command {
	cmd := &cobra.Command{
		Use:  orgName,
		Long: "Manage organizations",
		Args: cli.NoArgs,
		RunE: command.ShowHelp(dockerCli.Err()),
	}
	cmd.AddCommand(
		newListCmd(ctx, dockerCli, orgName),
		newMembersCmd(ctx, dockerCli, orgName),
		newTeamsCmd(ctx, dockerCli, orgName),
	)
	return cmd
}
