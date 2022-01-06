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
	"github.com/docker/hub-tool/internal/ansi"
	"github.com/docker/hub-tool/internal/errdef"
	"github.com/docker/hub-tool/internal/metrics"
	"github.com/docker/hub-tool/pkg/hub"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
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
			err := runRm(cmd.Context(), streams, hubClient, opts, args[0])
			if err == nil || errors.Is(err, errdef.ErrCanceled) {
				return nil
			}
			return err
		},
	}
	cmd.Flags().BoolVarP(&opts.force, "force", "f", false, "Force deletion of the tag")
	return cmd
}

func runRm(ctx context.Context, streams command.Streams, hubClient *hub.Client, opts rmOptions, image string) error {
	normRef, err := reference.ParseNormalizedNamed(image)
	if err != nil {
		return err
	}
	normRef = reference.TagNameOnly(normRef)
	ref, ok := normRef.(reference.NamedTagged)
	if !ok {
		return fmt.Errorf("invalid reference: tag must be specified")
	}

	if !opts.force {
		fmt.Fprintln(streams.Out(), ansi.Warn(fmt.Sprintf(`WARNING: You are about to permanently delete image "%s:%s"`, reference.FamiliarName(ref), ref.Tag())))
		fmt.Fprintln(streams.Out(), ansi.Warn("         This action is irreversible"))
		fmt.Fprintf(streams.Out(), ansi.Info("Are you sure you want to delete the image tagged %q from repository %q? [y/N] "), ref.Tag(), reference.FamiliarName(ref))
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
		if strings.ToLower(input) != "y" {
			return errors.New("deletion aborted")
		}
	}

	if err := hubClient.RemoveTag(reference.FamiliarName(ref), ref.Tag()); err != nil {
		if strings.Contains(err.Error(), "404 NOT FOUND") {
			fmt.Fprintln(streams.Out(), "Not Found", image)
			return nil
		} else {
			return err
		}
	}
	fmt.Fprintln(streams.Out(), "Deleted", image)
	return nil
}
