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
	"fmt"
	"testing"

	"gotest.tools/v3/assert"
	"gotest.tools/v3/icmd"

	"github.com/docker/hub-cli-plugin/internal"
)

func TestVersion(t *testing.T) {
	cmd, _, cleanup := dockerCli.createTestCmd()
	defer cleanup()

	// docker scan --version should use user's Snyk binary
	cmd.Command = dockerCli.Command("hub", "--version")
	output := icmd.RunCmd(cmd).Assert(t, icmd.Success).Combined()
	expected := fmt.Sprintf(
		`Version:    %s
Git commit: %s
`, internal.Version, internal.GitCommit)

	assert.Equal(t, output, expected)
}
