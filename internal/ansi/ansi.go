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
	"os"

	"github.com/cli/cli/pkg/iostreams"
	"github.com/mattn/go-isatty"
	"github.com/mgutz/ansi"
)

var (
	// Outputs ANSI color if stdout is a tty
	Red    = makeColorFunc("red")
	Yellow = makeColorFunc("yellow")
	Blue   = makeColorFunc("blue")
	Green  = makeColorFunc("green")
)

func makeColorFunc(color string) func(string) string {
	cf := ansi.ColorFunc(color)
	return func(arg string) string {
		if isColorEnabled() {
			if color == "black+h" && iostreams.Is256ColorSupported() {
				return fmt.Sprintf("\x1b[%d;5;%dm%s\x1b[m", 38, 242, arg)
			}
			return cf(arg)
		}
		return arg
	}
}

func isColorEnabled() bool {
	if iostreams.EnvColorForced() {
		return true
	}

	if iostreams.EnvColorDisabled() {
		return false
	}

	// TODO ignores cmd.OutOrStdout
	return IsTerminal(os.Stdout)
}

var IsTerminal = func(f *os.File) bool {
	return isatty.IsTerminal(f.Fd()) || IsCygwinTerminal(f)
}

func IsCygwinTerminal(f *os.File) bool {
	return isatty.IsCygwinTerminal(f.Fd())
}

var (
	// Title color should be used for any important title
	Title = Green
	// Header color should be used for all the listing column headers
	Header = Blue
	// Key color should be used for all key title content
	Key = Blue
	// Info color should be used when we prompt an info
	Info = Blue
	// Warn color should be used when we warn the user
	Warn = Yellow
	// Error color should be used when something bad happened
	Error = Red
	// Emphasise color should be used with important content
	Emphasise = Green
	// NoColor doesn't add any colors to the output
	NoColor = noop
)

func noop(in string) string {
	return in
}

// Link returns an ANSI terminal hyperlink
func Link(url string, text string) string {
	return fmt.Sprintf("\u001B]8;;%s\u0007%s\u001B]8;;\u0007", url, text)
}
