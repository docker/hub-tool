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

package login

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os"
	"runtime"
	"strings"

	"github.com/docker/cli/cli/command"
	dockerstreams "github.com/docker/cli/cli/streams"
	"github.com/moby/term"
	"github.com/pkg/errors"

	"github.com/docker/hub-tool/internal/ansi"
	"github.com/docker/hub-tool/internal/credentials"
	"github.com/docker/hub-tool/internal/errdef"
	"github.com/docker/hub-tool/pkg/hub"
)

// RunLogin logs the user and asks for the 2FA code if needed
func RunLogin(ctx context.Context, streams command.Streams, hubClient *hub.Client, store credentials.Store, candidateUsername string) error {
	username := candidateUsername
	if username == "" {
		username = os.Getenv("DOCKER_USERNAME")
	}
	if username == "" {
		var err error
		if username, err = readClearText(ctx, streams, "Username: "); err != nil {
			return err
		}
	}
	password := os.Getenv("DOCKER_PASSWORD")
	if password == "" {
		var err error
		if password, err = readPassword(streams); err != nil {
			return err
		}
	}

	token, refreshToken, err := Login(ctx, streams, hubClient, username, password)
	if err != nil {
		return err
	}

	if err := hubClient.Update(hub.WithHubToken(token)); err != nil {
		return err
	}

	return store.Store(credentials.Auth{
		Username:     username,
		Password:     password,
		Token:        token,
		RefreshToken: refreshToken,
	})
}

// Login runs login and optionnaly the 2FA
func Login(ctx context.Context, streams command.Streams, hubClient *hub.Client, username string, password string) (string, string, error) {
	return hubClient.Login(username, password, func() (string, error) {
		return readClearText(ctx, streams, "2FA required, please provide the 6 digit code: ")
	})
}

func readClearText(ctx context.Context, streams command.Streams, prompt string) (string, error) {
	userIn := make(chan string, 1)
	go func() {
		fmt.Fprint(streams.Out(), ansi.Info(prompt))
		reader := bufio.NewReader(streams.In())
		input, _ := reader.ReadString('\n')
		userIn <- strings.TrimSpace(input)
	}()
	input := ""
	select {
	case <-ctx.Done():
		return "", errdef.ErrCanceled
	case input = <-userIn:
	}
	return input, nil
}

func readPassword(streams command.Streams) (string, error) {
	in := streams.In()
	// On Windows, force the use of the regular OS stdin stream. Fixes #14336/#14210
	if runtime.GOOS == "windows" {
		in = dockerstreams.NewIn(os.Stdin)
	}

	// Some links documenting this:
	// - https://code.google.com/archive/p/mintty/issues/56
	// - https://github.com/docker/docker/issues/15272
	// - https://mintty.github.io/ (compatibility)
	// Linux will hit this if you attempt `cat | docker login`, and Windows
	// will hit this if you attempt docker login from mintty where stdin
	// is a pipe, not a character based console.
	if !streams.In().IsTerminal() {
		return "", errors.Errorf("cannot perform an interactive login from a non TTY device")
	}

	oldState, err := term.SaveState(in.FD())
	if err != nil {
		return "", err
	}
	fmt.Fprint(streams.Out(), ansi.Info("Password: "))
	if err := term.DisableEcho(in.FD(), oldState); err != nil {
		return "", err
	}

	password := readInput(in, streams.Out())
	fmt.Fprint(streams.Out(), "\n")

	if err := term.RestoreTerminal(in.FD(), oldState); err != nil {
		return "", err
	}
	if password == "" {
		return "", errors.Errorf("password required")
	}

	return password, nil
}

func readInput(in io.Reader, out io.Writer) string {
	reader := bufio.NewReader(in)
	line, _, err := reader.ReadLine()
	if err != nil {
		fmt.Fprintln(out, err.Error())
		os.Exit(1)
	}
	return string(line)
}
