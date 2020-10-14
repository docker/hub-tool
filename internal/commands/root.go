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

	"github.com/docker/cli/cli/command"
	"github.com/spf13/cobra"

	"github.com/docker/hub-tool/internal"
	"github.com/docker/hub-tool/internal/commands/account"
	"github.com/docker/hub-tool/internal/commands/org"
	"github.com/docker/hub-tool/internal/commands/repo"
	"github.com/docker/hub-tool/internal/commands/tag"
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
				return runVersion()
			}
			return cmd.Help()
		},
	}
	cmd.Flags().BoolVar(&flags.showVersion, "version", false, "Display version of the scan plugin")

	cmd.AddCommand(
		account.NewAccountCmd(streams, hubClient),
		org.NewOrgCmd(streams, hubClient),
		repo.NewRepoCmd(streams, hubClient),
		tag.NewTagCmd(streams, hubClient),
	)
	return cmd
}

func runVersion() error {
	fmt.Println("Version:   ", internal.Version)
	fmt.Println("Git commit:", internal.GitCommit)
	return nil
}
