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
	"github.com/docker/hub-tool/pkg/hub"
	"github.com/docker/hub-tool/internal/metrics"
)

const (
	activateName = "activate"
)

func newActivateCmd(streams command.Streams, hubClient *hub.Client, parent string) *cobra.Command {
	cmd := &cobra.Command{
		Use:                   activateName + " TOKEN_UUID",
		Short:                 "Activate a Personal Access Token",
		Args:                  cli.ExactArgs(1),
		DisableFlagsInUseLine: true,
		Annotations: map[string]string{
			"sudo": "true",
		},
		PreRun: func(cmd *cobra.Command, args []string) {
			metrics.Send(parent, activateName)
		},
		RunE: func(_ *cobra.Command, args []string) error {
			return runActivate(streams, hubClient, args[0])
		},
	}
	return cmd
}

func runActivate(streams command.Streams, hubClient *hub.Client, tokenUUID string) error {
	u, err := uuid.Parse(tokenUUID)
	if err != nil {
		return err
	}
	if _, err := hubClient.UpdateToken(u.String(), "", true); err != nil {
		return err
	}
	fmt.Fprintf(streams.Out(), ansi.Emphasise("%s is active\n"), u.String())
	return nil
}
