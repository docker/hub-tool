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
	"fmt"
	"testing"

	"gotest.tools/v3/assert"
	"gotest.tools/v3/icmd"

	"github.com/docker/hub-tool/internal"
)

func TestVersionCmd(t *testing.T) {
	cmd, cleanup := hubToolCmd(t, "version")
	defer cleanup()

	output := icmd.RunCmd(cmd).Assert(t, icmd.Success).Combined()
	expected := fmt.Sprintf("Version:    %s\nGit commit: %s\n", internal.Version, internal.GitCommit)

	assert.Equal(t, output, expected)
}

func TestVersionFlag(t *testing.T) {
	cmd, cleanup := hubToolCmd(t, "--version")
	defer cleanup()

	output := icmd.RunCmd(cmd).Assert(t, icmd.Success).Combined()
	expected := fmt.Sprintf("Docker Hub Tool %s, build %s\n", internal.Version, internal.GitCommit[:7])

	assert.Equal(t, output, expected)
}
