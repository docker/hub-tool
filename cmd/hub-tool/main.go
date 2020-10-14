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

package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/cli/cli/utils"
	"github.com/docker/cli/cli/command"
	cliflags "github.com/docker/cli/cli/flags"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/registry"

	"github.com/docker/hub-tool/internal/commands"
	"github.com/docker/hub-tool/internal/hub"
)

func main() {
	ctx, closeFunc := newSigContext()
	defer closeFunc()

	dockerCli, err := command.NewDockerCli()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	opts := cliflags.NewClientOptions()
	if err := dockerCli.Initialize(opts); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	authResolver := func(hub *registry.IndexInfo) types.AuthConfig {
		return command.ResolveAuthConfig(ctx, dockerCli, hub)
	}

	hubClient, err := hub.NewClient(authResolver, hub.WithContext(ctx))
	if err != nil {
		if hub.IsAuthenticationError(err) {
			fmt.Println(utils.Red(`You need to be logged in to Docker Hub to use this tool.
Please login to Docker Hub using the "docker login" command.`))
			os.Exit(1)
		}
		fmt.Println(err)
		os.Exit(1)
	}

	rootCmd := commands.NewRootCmd(dockerCli, hubClient, os.Args[0])
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
	os.Exit(0)
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
