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

package tag

import (
	"bufio"
	"context"
	"fmt"
	"strings"

	"github.com/docker/cli/cli"
	"github.com/docker/cli/cli/command"
	"github.com/docker/distribution/reference"
	"github.com/pkg/errors"
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
		Use:                   rmName + " [OPTIONS] REPOSITORY:TAG",
		Short:                 "Delete a tag in a repository",
		Args:                  cli.ExactArgs(1),
		DisableFlagsInUseLine: true,
		PreRun: func(cmd *cobra.Command, args []string) {
			metrics.Send(parent, rmName)
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return runRm(cmd.Context(), streams, hubClient, opts, args[0])
		},
	}
	cmd.Flags().BoolVarP(&opts.force, "force", "f", false, "Force deletion of the tag")
	return cmd
}

func runRm(ctx context.Context, streams command.Streams, hubClient *hub.Client, opts rmOptions, image string) error {
	ref, err := reference.ParseNormalizedNamed(image)
	if err != nil {
		return err
	}
	ref = reference.TagNameOnly(ref)
	namedTaggedRef, ok := ref.(reference.NamedTagged)
	if !ok {
		return fmt.Errorf("invalid reference: tag must be specified")
	}

	if !opts.force {
		fmt.Fprintln(streams.Out(), ansi.Warn("Please type the name of your tag to confirm deletion:"), reference.FamiliarString(namedTaggedRef))
		userIn := make(chan string, 1)
		go func() {
			reader := bufio.NewReader(streams.In())
			input, _ := reader.ReadString('\n')
			userIn <- strings.ToLower(strings.TrimSpace(input))
		}()
		input := ""
		select {
		case <-ctx.Done():
			return errors.New("canceled")
		case input = <-userIn:
		}
		if input != reference.FamiliarString(namedTaggedRef) {
			return fmt.Errorf("%q differs from your tag name, deletion aborted", input)
		}
	}

	if err := hubClient.RemoveTag(reference.FamiliarName(namedTaggedRef), namedTaggedRef.Tag()); err != nil {
		return err
	}
	fmt.Fprintln(streams.Out(), "Deleted", image)
	return nil
}
