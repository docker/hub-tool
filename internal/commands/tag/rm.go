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

package tag

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/docker/cli/cli"
	"github.com/docker/cli/cli/command"
	"github.com/docker/distribution/reference"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/registry"
	"github.com/spf13/cobra"

	"github.com/docker/hub-cli-plugin/internal/hub"
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
	ref, err := reference.Parse(image)
	if err != nil {
		return err
	}
	namedTaggedRef, ok := ref.(reference.NamedTagged)
	if !ok {
		return fmt.Errorf("invalid reference: tag must be specified")
	}

	if !opts.force {
		fmt.Println("Please type the name of your repository to confirm deletion:", namedTaggedRef.Name())
		reader := bufio.NewReader(os.Stdin)
		input, _ := reader.ReadString('\n')
		input = strings.ToLower(strings.TrimSpace(input))
		if input != namedTaggedRef.Name() {
			return fmt.Errorf("%q differs from your repository name, deletion aborted", input)
		}
	}

	authResolver := func(hub *registry.IndexInfo) types.AuthConfig {
		return command.ResolveAuthConfig(ctx, dockerCli, hub)
	}
	client, err := hub.NewClient(authResolver)
	if err != nil {
		return err
	}
	if err := client.RemoveTag(namedTaggedRef.Name(), namedTaggedRef.Tag()); err != nil {
		return err
	}
	fmt.Println("Deleted", image)
	return nil
}
