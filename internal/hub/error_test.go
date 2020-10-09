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

package hub

import (
	"errors"
	"testing"

	"gotest.tools/v3/assert"
)

func TestIsAuthenticationError(t *testing.T) {
	assert.Assert(t, IsAuthenticationError(&authenticationError{}))
	assert.Assert(t, !IsAuthenticationError(errors.New("")))
}

func TestIsInvalidTokenError(t *testing.T) {
	assert.Assert(t, IsInvalidTokenError(&invalidTokenError{}))
	assert.Assert(t, !IsInvalidTokenError(errors.New("")))
}
