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

package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/docker/cli/cli-plugins/manager"
	"github.com/docker/cli/cli-plugins/plugin"
	"github.com/docker/cli/cli/command"
	"github.com/spf13/cobra"

	"github.com/docker/hub-cli-plugin/internal"
)

func main() {
	ctx, closeFunc := newSigContext()
	defer closeFunc()
	plugin.Run(func(dockerCli command.Cli) *cobra.Command {
		cmd := newHubCmd(ctx, dockerCli)
		originalPreRun := cmd.PersistentPreRunE
		cmd.PersistentPreRunE = func(cmd *cobra.Command, args []string) error {
			if err := plugin.PersistentPreRunE(cmd, args); err != nil {
				return err
			}
			if originalPreRun != nil {
				return originalPreRun(cmd, args)
			}
			return nil
		}
		return cmd
	}, manager.Metadata{
		SchemaVersion: "0.1.0",
		Vendor:        "Docker Inc.",
		Version:       internal.Version,
	})
}

type options struct {
	showVersion bool
}

func newHubCmd(_ context.Context, _ command.Cli) *cobra.Command {
	var flags options
	cmd := &cobra.Command{
		Short:       "Docker Hub",
		Long:        `A tool to manage your Docker Hub images`,
		Use:         "hub",
		Annotations: map[string]string{},
		RunE: func(cmd *cobra.Command, args []string) error {
			if flags.showVersion {
				return runVersion()
			}
			return nil
		},
	}
	cmd.Flags().BoolVar(&flags.showVersion, "version", false, "Display version of the scan plugin")

	return cmd
}

func runVersion() error {
	fmt.Println("Version:   ", internal.Version)
	fmt.Println("Git commit:", internal.GitCommit)
	return nil
}

func newSigContext() (context.Context, func()) {
	ctx, cancel := context.WithCancel(context.Background())
	s := make(chan os.Signal)
	signal.Notify(s, syscall.SIGTERM, syscall.SIGINT)
	go func() {
		<-s
		cancel()
	}()
	return ctx, cancel
}
