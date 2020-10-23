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

	"github.com/docker/hub-tool/internal"
	"github.com/docker/hub-tool/internal/commands/account"
	"github.com/docker/hub-tool/internal/commands/org"
	"github.com/docker/hub-tool/internal/commands/repo"
	"github.com/docker/hub-tool/internal/commands/tag"
	"github.com/docker/hub-tool/internal/commands/token"
	"github.com/docker/hub-tool/internal/hub"
)

type options struct {
	showVersion bool
}

//NewRootCmd returns the main command
func NewRootCmd(streams command.Streams, hubClient *hub.Client, name string) *cobra.Command {
	var flags options
	cmd := &cobra.Command{
		Use:         name,
		Short:       "Docker Hub Tool",
		Long:        `A tool to manage your Docker Hub images`,
		Annotations: map[string]string{},
		RunE: func(cmd *cobra.Command, args []string) error {
			if flags.showVersion {
				fmt.Fprintf(streams.Out(), "Docker Hub Tool %s, build %s\n", internal.Version, internal.GitCommit[:7])
				return nil
			}
			return cmd.Help()
		},
	}
	cmd.Flags().BoolVar(&flags.showVersion, "version", false, "Display the version of this tool")

	cmd.AddCommand(
		account.NewAccountCmd(streams, hubClient),
		token.NewTokenCmd(streams, hubClient),
		org.NewOrgCmd(streams, hubClient),
		repo.NewRepoCmd(streams, hubClient),
		tag.NewTagCmd(streams, hubClient),
		newVersionCmd(streams),
	)
	return cmd
}

func newVersionCmd(streams command.Streams) *cobra.Command {
	return &cobra.Command{
		Use:  "version",
		Long: "Version information about this tool",
		Args: cli.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			_, err := fmt.Fprintf(streams.Out(), "Version:    %s\nGit commit: %s\n", internal.Version, internal.GitCommit)
			return err
		},
	}
}
