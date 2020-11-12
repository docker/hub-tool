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

package token

import (
	"fmt"

	"github.com/docker/cli/cli"
	"github.com/docker/cli/cli/command"
	"github.com/google/uuid"
	"github.com/spf13/cobra"

	"github.com/docker/hub-tool/internal/ansi"
	"github.com/docker/hub-tool/internal/hub"
	"github.com/docker/hub-tool/internal/metrics"
)

const (
	updateName = "update"
)

type updateOptions struct {
	description string
	setActive   bool
}

func newUpdateCmd(streams command.Streams, hubClient *hub.Client, parent string) *cobra.Command {
	var opts updateOptions
	cmd := &cobra.Command{
		Use:                   updateName + " [OPTIONS] TOKEN_UUID",
		Short:                 "Update a Personal Access Token",
		Args:                  cli.ExactArgs(1),
		DisableFlagsInUseLine: true,
		Annotations: map[string]string{
			"sudo": "true",
		},
		PreRun: func(cmd *cobra.Command, args []string) {
			metrics.Send(parent, updateName)
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return runUpdate(streams, hubClient, cmd, opts, args[0])
		},
	}
	cmd.Flags().StringVar(&opts.description, "description", "", "Set token's description")
	cmd.Flags().BoolVar(&opts.setActive, "set-active", true, "Activate or deactivate the token")
	return cmd
}

func runUpdate(streams command.Streams, hubClient *hub.Client, cmd *cobra.Command, opts updateOptions, tokenUUID string) error {
	u, err := uuid.Parse(tokenUUID)
	if err != nil {
		return err
	}
	token, err := hubClient.GetToken(u.String())
	if err != nil {
		return err
	}
	description := token.Description
	if cmd.Flag("description").Changed {
		description = opts.description
	}
	isActive := token.IsActive
	if cmd.Flag("set-active").Changed {
		isActive = opts.setActive
	}
	_, err = hubClient.UpdateToken(u.String(), description, isActive)
	if err != nil {
		return err
	}
	fmt.Fprintln(streams.Out(), ansi.Emphasise("Updated"), u.String())
	return nil
}
