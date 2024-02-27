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
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/docker/cli/cli/command"
	dockercredentials "github.com/docker/cli/cli/config/credentials"
	cliflags "github.com/docker/cli/cli/flags"

	"github.com/docker/hub-tool/internal/commands"
	"github.com/docker/hub-tool/pkg/credentials"
	"github.com/docker/hub-tool/pkg/hub"
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
		log.Fatal(err)
	}

	store := credentials.NewStore(func(key string) dockercredentials.Store {
		config := dockerCli.ConfigFile()
		return config.GetCredentialsStore(key)
	})
	auth, err := store.GetAuth()
	if err != nil {
		log.Fatal(err)
	}

	hubClient, err := hub.NewClient(
		hub.WithContext(ctx),
		hub.WithInStream(dockerCli.In()),
		hub.WithOutStream(dockerCli.Out()),
		hub.WithHubAccount(auth.Username),
		hub.WithPassword(auth.Password),
		hub.WithRefreshToken(auth.RefreshToken),
		hub.WithHubToken(auth.Token))
	if err != nil {
		log.Fatal(err)
	}

	rootCmd := commands.NewRootCmd(dockerCli, hubClient, store, os.Args[0])
	if err := rootCmd.ExecuteContext(ctx); err != nil {
		os.Exit(1)
	}
	os.Exit(0)
}

func newSigContext() (context.Context, func()) {
	ctx, cancel := context.WithCancel(context.Background())
	s := make(chan os.Signal, 1)
	signal.Notify(s, syscall.SIGTERM, syscall.SIGINT)
	go func() {
		<-s
		cancel()
	}()
	return ctx, cancel
}
