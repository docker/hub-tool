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
	"context"
	"fmt"
	"strings"

	"github.com/pkg/errors"

	"github.com/docker/hub-tool/internal/errdef"

	"github.com/distribution/reference"
	"github.com/docker/cli/cli"
	"github.com/docker/cli/cli/command"
	"github.com/spf13/cobra"

	"github.com/docker/hub-tool/internal/ansi"
	"github.com/docker/hub-tool/internal/metrics"
	"github.com/docker/hub-tool/pkg/hub"
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
		Use:                   rmName + " [OPTIONS] NAMESPACE/REPOSITORY",
		Short:                 "Delete a repository",
		Args:                  cli.ExactArgs(1),
		DisableFlagsInUseLine: true,
		PreRun: func(cmd *cobra.Command, args []string) {
			metrics.Send(parent, rmName)
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			err := runRm(cmd.Context(), streams, hubClient, opts, args[0])
			if err == nil || errors.Is(err, errdef.ErrCanceled) {
				return nil
			}
			return err
		},
	}
	cmd.Flags().BoolVarP(&opts.force, "force", "f", false, "Force deletion of the repository")
	return cmd
}

func runRm(ctx context.Context, streams command.Streams, hubClient *hub.Client, opts rmOptions, repository string) error {
	ref, err := reference.Parse(repository)
	if err != nil {
		return err
	}
	namedRef, ok := ref.(reference.Named)
	if !ok {
		return errors.New("invalid reference: repository not specified")
	}

	if !strings.Contains(repository, "/") {
		return fmt.Errorf("repository name must include username or organization name, example: hub-tool repo rm username/repository")
	}

	if !opts.force {
		_, count, err := hubClient.GetTags(namedRef.Name())
		if err != nil {
			return err
		}
		fmt.Fprintln(streams.Out(), ansi.Warn(fmt.Sprintf("WARNING: You are about to permanently delete repository %q including %d tag(s)", namedRef.Name(), count)))
		fmt.Fprintln(streams.Out(), ansi.Warn("         This action is irreversible"))
		fmt.Fprintln(streams.Out(), ansi.Info("Enter the name of the repository to confirm deletion:"), namedRef.Name())
		userIn := make(chan string, 1)
		go func() {
			reader := bufio.NewReader(streams.In())
			input, _ := reader.ReadString('\n')
			userIn <- strings.ToLower(strings.TrimSpace(input))
		}()
		input := ""
		select {
		case <-ctx.Done():
			return errdef.ErrCanceled
		case input = <-userIn:
		}
		if input != namedRef.Name() {
			return fmt.Errorf("%q differs from your repository name, deletion aborted", input)
		}
	}

	if err := hubClient.RemoveRepository(namedRef.Name()); err != nil {
		return err
	}
	_, err = fmt.Fprintf(streams.Out(), "Repository %q was successfully deleted\n", repository)
	if err != nil {
		return err
	}
	return nil
}
