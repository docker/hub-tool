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

package commands

import (
	"fmt"
	"os"

	"github.com/docker/cli/cli"
	"github.com/docker/cli/cli/command"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"github.com/docker/hub-tool/internal"
	"github.com/docker/hub-tool/internal/ansi"
	"github.com/docker/hub-tool/internal/commands/account"
	"github.com/docker/hub-tool/internal/commands/org"
	"github.com/docker/hub-tool/internal/commands/repo"
	"github.com/docker/hub-tool/internal/commands/tag"
	"github.com/docker/hub-tool/internal/commands/token"
	"github.com/docker/hub-tool/internal/credentials"
	"github.com/docker/hub-tool/internal/hub"
	"github.com/docker/hub-tool/internal/login"
)

type options struct {
	showVersion bool
	trace       bool
	verbose     bool
}

var (
	anonCmds = []string{"version", "help", "login"}
)

// NewRootCmd returns the main command
func NewRootCmd(streams command.Streams, hubClient *hub.Client, store credentials.Store, name string) *cobra.Command {
	var flags options
	cmd := &cobra.Command{
		Use:                   name,
		Short:                 "Docker Hub Tool",
		Long:                  `A tool to manage your Docker Hub images`,
		Annotations:           map[string]string{},
		SilenceUsage:          true,
		DisableFlagsInUseLine: true,
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			if flags.trace {
				log.SetLevel(log.TraceLevel)
			} else if flags.verbose {
				log.SetLevel(log.DebugLevel)
			}
			if flags.showVersion {
				return nil
			}
			if contains(anonCmds, cmd.Name()) {
				return nil
			}

			if cmd.Annotations["sudo"] == "true" {
				ac, err := store.GetAuth()
				if err != nil {
					return err
				}
				if ac.TokenExpired() {
					return login.RunLogin(streams, hubClient, store, ac.Username)
				}
				return nil
			}

			ac, err := store.GetAuth()
			if err != nil || ac.Username == "" {
				fmt.Println(ansi.Error(`You need to be logged in to Docker Hub to use this tool.
Please login to Docker Hub using the "hub-tool login" command.`))
				os.Exit(1)
			}
			if ac.TokenExpired() {
				t, p, err := hubClient.Login(ac.Username, ac.Password, func() (string, error) {
					return "", nil
				})
				if err != nil {
					return err
				}

				return store.Store(credentials.Auth{
					Username: ac.Username,
					Password: p,
					Token:    t,
				})
			}

			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			if flags.showVersion {
				fmt.Fprintf(streams.Out(), "Docker Hub Tool %s, build %s\n", internal.Version, internal.GitCommit[:7])
				return nil
			}
			return cmd.Help()
		},
	}
	cmd.Flags().BoolVar(&flags.showVersion, "version", false, "Display the version of this tool")
	cmd.PersistentFlags().BoolVar(&flags.verbose, "verbose", false, "Print logs")
	cmd.PersistentFlags().BoolVar(&flags.trace, "trace", false, "Print trace logs")
	_ = cmd.PersistentFlags().MarkHidden("trace")

	cmd.AddCommand(
		newLoginCmd(streams, store, hubClient),
		account.NewAccountCmd(streams, hubClient),
		token.NewTokenCmd(streams, hubClient),
		org.NewOrgCmd(streams, hubClient),
		repo.NewRepoCmd(streams, hubClient),
		tag.NewTagCmd(streams, hubClient),
		newVersionCmd(streams),
	)
	return cmd
}

func contains(haystack []string, needle string) bool {
	for _, v := range haystack {
		if needle == v {
			return true
		}
	}
	return false
}

func newVersionCmd(streams command.Streams) *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Version information about this tool",
		Args:  cli.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			_, err := fmt.Fprintf(streams.Out(), "Version:    %s\nGit commit: %s\n", internal.Version, internal.GitCommit)
			return err
		},
	}
}
