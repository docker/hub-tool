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
	"github.com/docker/cli/cli"
	"github.com/docker/cli/cli/command"
	"github.com/spf13/cobra"

	"github.com/docker/hub-tool/internal/hub"
)

const (
	tagName = "tag"
)

//NewTagCmd configures the tag manage command
func NewTagCmd(streams command.Streams, hubClient *hub.Client) *cobra.Command {
	cmd := &cobra.Command{
		Use:   tagName,
		Short: "Manage tags",
		Args:  cli.NoArgs,
		RunE:  command.ShowHelp(streams.Err()),
	}
	cmd.AddCommand(
		newInspectCmd(streams, hubClient, tagName),
		newListCmd(streams, hubClient, tagName),
		newRmCmd(streams, hubClient, tagName),
	)
	return cmd
}
