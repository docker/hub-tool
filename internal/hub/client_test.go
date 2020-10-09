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
	"net/http"
	"net/http/httptest"
	"testing"

	"gotest.tools/v3/assert"

	"github.com/docker/hub-cli-plugin/internal"
)

func TestDoRequestAddsCustomUserAgent(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, r.Header.Get("Accept"), "application/json")
		assert.Equal(t, r.Header.Get("User-Agent"), "hub-tool/"+internal.Version)
	}))
	defer server.Close()
	req, err := http.NewRequest("GET", server.URL, nil)
	assert.NilError(t, err)
	_, err = doRequest(req)
	assert.NilError(t, err)
}
