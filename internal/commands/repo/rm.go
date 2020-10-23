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

package repo

import (
	"bufio"
	"fmt"
	"strings"

	"github.com/docker/cli/cli"
	"github.com/docker/cli/cli/command"
	"github.com/docker/distribution/reference"
	"github.com/spf13/cobra"

	"github.com/docker/hub-tool/internal/ansi"
	"github.com/docker/hub-tool/internal/hub"
	"github.com/docker/hub-tool/internal/metrics"
)

const (
	rmName = "rm"
)

type rmOptions struct {
	force bool
}

func newRmCmd(streams command.Streams, hubClient *hub.Client, parent string) *cobra.Command {
	var opts rmOptions
	cmd := &cobra.Command{
		Use:                   rmName + " [OPTIONS] REPOSITORY",
		Short:                 "Delete a repository",
		Args:                  cli.ExactArgs(1),
		DisableFlagsInUseLine: true,
		PreRun: func(cmd *cobra.Command, args []string) {
			metrics.Send(parent, rmName)
		},
		RunE: func(_ *cobra.Command, args []string) error {
			return runRm(streams, hubClient, opts, args[0])
		},
	}
	cmd.Flags().BoolVar(&opts.force, "force", false, "Force deletion of the repository")
	return cmd
}

func runRm(streams command.Streams, hubClient *hub.Client, opts rmOptions, repository string) error {
	ref, err := reference.Parse(repository)
	if err != nil {
		return err
	}
	namedRef, ok := ref.(reference.Named)
	if !ok {
		return fmt.Errorf("invalid reference: repository not specified")
	}

	if !opts.force {
		fmt.Fprintln(streams.Out(), ansi.Warn("Please type the name of your repository to confirm deletion:"), namedRef.Name())
		reader := bufio.NewReader(streams.In())
		input, _ := reader.ReadString('\n')
		input = strings.ToLower(strings.TrimSpace(input))
		if input != namedRef.Name() {
			return fmt.Errorf("%q differs from your repository name, deletion aborted", input)
		}
	}

	if err := hubClient.RemoveRepository(namedRef.Name()); err != nil {
		return err
	}
	fmt.Fprintln(streams.Out(), "Deleted", repository)
	return nil
}
