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

package org

import (
	"fmt"
	"io"

	"github.com/cli/cli/utils"
	"github.com/docker/cli/cli"
	"github.com/docker/cli/cli/command"
	"github.com/docker/hub-tool/internal/ansi"
	"github.com/docker/hub-tool/internal/format"
	"github.com/docker/hub-tool/internal/hub"
	"github.com/docker/hub-tool/internal/metrics"
	"github.com/spf13/cobra"
)

const (
	settingsName = "settings"
)

type settingsOptions struct {
	format.Option
}

func newSettingsCmd(streams command.Streams, hubClient *hub.Client, parent string) *cobra.Command {
	var opts settingsOptions
	cmd := &cobra.Command{
		Use:                   settingsName + " ORGANIZATION",
		Short:                 "Print the organization settings",
		Args:                  cli.ExactArgs(1),
		DisableFlagsInUseLine: true,
		Annotations: map[string]string{
			"sudo": "true",
		},
		PreRun: func(cmd *cobra.Command, args []string) {
			metrics.Send(parent, settingsName)
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return runOrgSettings(streams, hubClient, opts, args[0])
		},
	}
	opts.AddFormatFlag(cmd.Flags())
	return cmd
}

func runOrgSettings(streams command.Streams, hubClient *hub.Client, opts settingsOptions, orgName string) error {
	settings, err := hubClient.GetOrganizationSettings(orgName)
	if hub.IsForbiddenError(err) {
		return fmt.Errorf(ansi.Error("failed to get organization settings, you need to be the organization Owner"))
	}
	if err != nil {
		return err
	}
	return opts.Print(streams.Out(), *settings, printSettings)
}

func printSettings(out io.Writer, value interface{}) error {
	settings := value.(hub.OrgSettings)

	// print user info
	fmt.Fprintf(out, ansi.Key("Restricted Images Access:               ")+"%s\n", enabled(settings.RestrictedImages.Enabled))
	color := ansi.Key
	if !settings.RestrictedImages.Enabled {
		color = utils.Gray
	}
	fmt.Fprintf(out, color("Allow use of Official images:           ")+"%s\n", enabled(settings.RestrictedImages.AllowOfficialImages))
	fmt.Fprintf(out, color("Allow use of Verified Publisher images: ")+"%s\n", enabled(settings.RestrictedImages.AllowVerifiedPublishers))
	return nil
}

func enabled(settings bool) string {
	if settings {
		return ansi.Emphasise("Enabled")
	}
	return "Disabled"
}
