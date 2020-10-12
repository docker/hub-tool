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
	"testing"

	"gotest.tools/v3/icmd"
)

func TestUserNeedsToBeLoggedIn(t *testing.T) {
	cmd, cleanup := hubToolCmd(t, "--version")
	// Remove the config file
	cleanup()

	output := icmd.RunCmd(cmd)
	output.Equal(icmd.Expected{
		ExitCode: 1,
		Out: `You need to be logged in to Docker Hub to use this tool.
Please login to Docker Hub using the "docker login" command
`,
	})
}
