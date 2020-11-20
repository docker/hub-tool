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

package e2e

import (
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/docker/cli/cli/config/configfile"
	clitypes "github.com/docker/cli/cli/config/types"
	"gotest.tools/v3/assert"
	"gotest.tools/v3/env"
	"gotest.tools/v3/fs"
	"gotest.tools/v3/icmd"
)

func hubToolCmd(t *testing.T, args ...string) (icmd.Cmd, func()) {
	user := os.Getenv("E2E_HUB_USERNAME")
	token := os.Getenv("E2E_HUB_TOKEN")

	config := configfile.ConfigFile{
		AuthConfigs: map[string]clitypes.AuthConfig{"https://index.docker.io/v1/": {
			Username: user,
			Password: token,
		}},
	}
	data, err := json.Marshal(&config)
	assert.NilError(t, err)

	pwd, err := os.Getwd()
	assert.NilError(t, err)
	hubTool := os.Getenv("BINARY")
	configDir := fs.NewDir(t, t.Name(), fs.WithFile("config.json", string(data)))
	cleanup := env.Patch(t, "PATH", os.Getenv("PATH")+getPathSeparator()+filepath.Join(pwd, "..", "bin"))
	env := append(os.Environ(), "DOCKER_CONFIG="+configDir.Path())

	return icmd.Cmd{Command: append([]string{hubTool}, args...), Env: env}, func() { cleanup(); configDir.Remove() }
}

func getPathSeparator() string {
	if runtime.GOOS == "windows" {
		return ";"
	}
	return ":"
}
