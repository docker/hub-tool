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
	"testing"
	"time"

	"github.com/docker/distribution/reference"
	"github.com/opencontainers/image-spec/specs-go"
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
	"gotest.tools/v3/assert"
	"gotest.tools/v3/golden"
)

func TestPrintImage(t *testing.T) {
	manifest := ocispec.Manifest{
		Versioned: specs.Versioned{},
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
		Annotations: map[string]string{
			"annotation1": "value1",
			"annotation2": "value2",
		},
	}
	now := time.Now()
	config := ocispec.Image{
		Created:      &now,
		Author:       "author",
		Architecture: "arch",
		OS:           "os",
		Config: ocispec.ImageConfig{
			User:         "user",
			ExposedPorts: map[string]struct{}{"80/tcp": {}},
			Env:          []string{"KEY=VALUE"},
			Entrypoint:   []string{"./entrypoint", "parameter"},
			Cmd:          []string{"./cmd", "parameter"},
			Volumes:      map[string]struct{}{"/volume": {}},
			WorkingDir:   "/workingdir",
			Labels:       map[string]string{"label": "value"},
			StopSignal:   "SIGTERM",
		},
		History: []ocispec.History{
			// empty layer is ignored
			{
				Created:    &now,
				CreatedBy:  "created-by-ignored",
				Author:     "author-ignored",
				Comment:    "comment-ignored",
				EmptyLayer: true,
			},
			{
				Created:    &now,
				CreatedBy:  "/bin/sh -c #(nop) apt-get install",
				Author:     "author-history",
				Comment:    "comment-history",
				EmptyLayer: false,
			},
		},
	}
	manifestDescriptor := ocispec.Descriptor{
		MediaType: "mediatype/manifest",
		Digest:    "sha256:abcdef",
		Size:      789,
		Platform: &ocispec.Platform{
			Architecture: "arch",
			OS:           "os",
			OSVersion:    "osversion",
			OSFeatures:   []string{"feature1", "feature2"},
			Variant:      "variant",
		},
	}
	ref, err := reference.ParseDockerRef("image:latest")
	assert.NilError(t, err)
	image := Image{ref.Name(), manifest, config, manifestDescriptor}

	out := bytes.NewBuffer(nil)
	err = printImage(out, &image)
	assert.NilError(t, err)
	golden.Assert(t, out.String(), "printimage.golden")
}
