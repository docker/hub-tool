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
	"errors"
	"fmt"
	"io"

	"github.com/docker/cli/cli"
	"github.com/docker/cli/cli/command"
	"github.com/spf13/cobra"

	"github.com/docker/hub-tool/internal/ansi"
	"github.com/docker/hub-tool/internal/format"
	"github.com/docker/hub-tool/internal/metrics"
	"github.com/docker/hub-tool/pkg/hub"
)

const (
	createName = "create"
)

type createOptions struct {
	format.Option
	description string
	quiet       bool
	scope       string
}

func validScope(scope string) bool {
	switch scope {
	case
		"admin",
		"write",
		"read",
		"public_read":
		return true
	}
	return false
}

func newCreateCmd(streams command.Streams, hubClient *hub.Client, parent string) *cobra.Command {
	var opts createOptions
	cmd := &cobra.Command{
		Use:                   createName + " [OPTIONS]",
		Short:                 "Create a Personal Access Token",
		Args:                  cli.NoArgs,
		DisableFlagsInUseLine: true,
		Annotations: map[string]string{
			"sudo": "true",
		},
		PreRun: func(cmd *cobra.Command, args []string) {
			metrics.Send(parent, createName)
		},
		RunE: func(_ *cobra.Command, args []string) error {

			return runCreate(streams, hubClient, opts)
		},
	}
	opts.AddFormatFlag(cmd.Flags())
	cmd.Flags().StringVar(&opts.description, "description", "", "Set token's description")
	cmd.Flags().BoolVar(&opts.quiet, "quiet", false, "Display only created token")
	cmd.Flags().StringVar(&opts.scope, "scope", "", "Set token's repo scope (admin,write,read,public_read)")

	return cmd
}

func runCreate(streams command.Streams, hubClient *hub.Client, opts createOptions) error {
	if len(opts.scope) > 0 && !validScope(opts.scope) {
		return errors.New("scope must be one of admin,write,read,public_read")
	}
	token, err := hubClient.CreateToken(opts.description, opts.scope)
	if err != nil {
		return err
	}
	if opts.quiet {
		fmt.Fprintln(streams.Out(), token.Token)
		return nil
	}
	return opts.Print(streams.Out(), token, printCreatedToken(hubClient))
}

func printCreatedToken(hubClient *hub.Client) format.PrettyPrinter {
	return func(out io.Writer, value interface{}) error {
		helper := value.(*hub.Token)
		fmt.Fprintf(out, ansi.Emphasise("Personal Access Token successfully created!")+`

When logging in from your Docker CLI client, use this token as a password.
`+ansi.Header("Description:")+` %s

To use the access token from your Docker CLI client:
1. Run: docker login --username %s
2. At the password prompt, enter the personal access token.

    %s

`+ansi.Warn(`WARNING: This access token cannot be displayed again.
It will not be stored and cannot be retrieved. Please be sure to save it now.
`),
			helper.Description,
			hubClient.AuthConfig.Username,
			ansi.Emphasise(helper.Token))
		return nil
	}
}
