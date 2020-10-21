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

package tabwriter

import (
	"bytes"
	"testing"

	"gotest.tools/v3/assert"
	"gotest.tools/v3/golden"
)

func TestSimple(t *testing.T) {
	b := bytes.NewBuffer(nil)
	tw := New(b, "")
	s := "test"
	tw.Column(s, len(s))
	err := tw.Flush()
	assert.NilError(t, err)
	assert.Equal(t, s+"\n", b.String())
}

func TestTwoLines(t *testing.T) {
	b := bytes.NewBuffer(nil)
	tw := New(b, " ")
	first := "test"
	second := "docker"
	tw.Column(first, len(first))
	tw.Column(first, len(first))
	tw.Line()
	tw.Column(second, len(second))
	tw.Column(second, len(second))
	err := tw.Flush()
	assert.NilError(t, err)
	golden.Assert(t, b.String(), "twolines.golden")
}
