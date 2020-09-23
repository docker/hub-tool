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

package commands

import (
	"context"
	"fmt"

	"github.com/docker/cli/cli/command"
	"github.com/spf13/cobra"

	"github.com/docker/hub-cli-plugin/internal"
	"github.com/docker/hub-cli-plugin/internal/commands/tag"
)

type options struct {
	showVersion bool
}

//NewHubCmd returns the main command
func NewHubCmd(ctx context.Context, dockerCli command.Cli) *cobra.Command {
	var flags options
	cmd := &cobra.Command{
		Use:         "hub",
		Short:       "Docker Hub",
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

	cmd.AddCommand(tag.NewTagCmd(ctx, dockerCli))
	return cmd
}

func runVersion() error {
	fmt.Println("Version:   ", internal.Version)
	fmt.Println("Git commit:", internal.GitCommit)
	return nil
}
