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

package tag

import (
	"bytes"
	"encoding/json"
	"testing"

	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
	"gotest.tools/golden"
	"gotest.tools/v3/assert"
)

func TestPrintManifest(t *testing.T) {
	manifest := ocispec.Manifest{
		Config: ocispec.Descriptor{
			MediaType: "mediatype/config",
			Digest:    "sha256:beef",
			Size:      123,
		},
		Layers: []ocispec.Descriptor{
			{
				MediaType: "mediatype/layer",
				Digest:    "sha256:c0ffee",
				Size:      456,
			},
		},
	}
	raw, err := json.Marshal(&manifest)
	assert.NilError(t, err)
	manifestDescriptor := ocispec.Descriptor{
		MediaType: "mediatype/manifest",
		Digest:    "sha256:abcdef",
		Size:      789,
	}
	out := bytes.NewBuffer(nil)

	err = printManifest(raw, manifestDescriptor, "image:latest", out)
	assert.NilError(t, err)
	golden.Assert(t, out.String(), "printmanifest.golden")
}
