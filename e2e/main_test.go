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

package e2e

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	dockerConfigFile "github.com/docker/cli/cli/config/configfile"
	"gotest.tools/v3/icmd"
)

var (
	dockerCli dockerCliCommand
)

type dockerCliCommand struct {
	path         string
	cliPluginDir string
}

func (d dockerCliCommand) createTestCmd() (icmd.Cmd, string, func()) {
	configDir, err := ioutil.TempDir("", "config")
	if err != nil {
		panic(err)
	}
	if err := os.MkdirAll(filepath.Join(configDir, "hub"), 0744); err != nil {
		panic(err)
	}

	configFilePath := filepath.Join(configDir, "config.json")
	dockerConfig := dockerConfigFile.ConfigFile{
		CLIPluginsExtraDirs: []string{
			d.cliPluginDir,
		},
		Filename: configFilePath,
	}
	configFile, err := os.Create(configFilePath)
	if err != nil {
		panic(err)
	}
	//nolint:errcheck
	defer configFile.Close()
	err = json.NewEncoder(configFile).Encode(dockerConfig)
	if err != nil {
		panic(err)
	}
	cleanup := func() {
		_ = os.RemoveAll(configDir)
	}
	env := append(os.Environ(),
		"DOCKER_CONFIG="+configDir,
		"DOCKER_CLI_EXPERIMENTAL=enabled") // TODO: Remove this once docker app plugin is no more experimental
	return icmd.Cmd{Env: env}, configDir, cleanup
}

func (d dockerCliCommand) Command(args ...string) []string {
	return append([]string{d.path}, args...)
}

func TestMain(m *testing.M) {
	// Prepare docker cli to call the docker-scan plugin binary:
	// - Create a symbolic link with the docker-scan binary to the plugin directory
	cliPluginDir, err := ioutil.TempDir("", "configContent")
	if err != nil {
		panic(err)
	}
	//nolint:errcheck
	defer os.RemoveAll(cliPluginDir)
	sourceDir := filepath.Join(os.Getenv("DOCKER_CONFIG"), "cli-plugins")
	copyBinary("docker-hub", sourceDir, cliPluginDir)

	dockerCli = dockerCliCommand{path: "docker", cliPluginDir: cliPluginDir}
	os.Exit(m.Run())
}

func copyBinary(binaryName, sourceDir, configDir string) {
	if runtime.GOOS == "windows" {
		binaryName += ".exe"
	}
	input, err := ioutil.ReadFile(filepath.Join(sourceDir, binaryName))
	if err != nil {
		panic(err)
	}
	err = ioutil.WriteFile(filepath.Join(configDir, binaryName), input, 0744)
	if err != nil {
		panic(err)
	}
}
