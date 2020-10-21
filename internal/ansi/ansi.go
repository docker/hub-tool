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

package ansi

import (
	"fmt"

	"github.com/cli/cli/utils"
)

var (
	// Title color should be used for any important title
	Title = utils.Green
	// Header color should be used for all the listing column headers
	Header = utils.Blue
	// Key color should be used for all key title content
	Key = utils.Blue
	// Info color should be used when we prompt an info
	Info = utils.Blue
	// Warn color should be used when we warn the user
	Warn = utils.Yellow
	// Error color should be used when something bad happened
	Error = utils.Red
	// Emphasise color should be used with important content
	Emphasise = utils.Green
)

// Link returns an ANSI terminal hyperlink
func Link(url string, text string) string {
	return fmt.Sprintf("\u001B]8;;%s\u0007%s\u001B]8;;\u0007", url, text)
}
