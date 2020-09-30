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

package metrics

import (
	"strings"

	"github.com/docker/compose-cli/metrics"
)

//Send emit a metric message to the local Desktop metric listener
func Send(parent, command string) {
	client := metrics.NewClient()
	client.Send(metrics.Command{
		Command: strings.Join([]string{parent, command}, " "),
		Context: "hub",
		Source:  "hub-cli",
		Status:  metrics.SuccessStatus,
	})
}
