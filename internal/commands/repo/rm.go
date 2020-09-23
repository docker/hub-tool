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

package repo

import (
	"context"

	"github.com/docker/cli/cli"
	"github.com/docker/cli/cli/command"
	"github.com/spf13/cobra"
)

type rmiOptions struct {
	force bool
}

func newRmCmd(ctx context.Context, dockerCli command.Cli) *cobra.Command {
	var opts rmiOptions
	cmd := &cobra.Command{
		Use:   "rm REPOSITORY:TAG",
		Short: "Delete a tag in a repository",
		Args:  cli.ExactArgs(1),
		RunE: func(_ *cobra.Command, args []string) error {
			return runRm(ctx, dockerCli, opts, args[0])
		},
	}
	cmd.Flags().BoolVar(&opts.force, "platforms", false, "List all available platforms per tag")
	return cmd
}

func runRm(ctx context.Context, dockerCli command.Cli, opts rmiOptions, image string) error {
	return nil
}
