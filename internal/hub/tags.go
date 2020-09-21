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

package hub

import (
	"time"

	"github.com/docker/distribution/reference"
)

//Tag can point to a manifest or manifest list
type Tag struct {
	Name                string
	FullSize            int
	LastUpdated         time.Time
	LastUpdaterUserName string
	Images              []Image
}

//Image represents the metadata of a manifest
type Image struct {
	Digest       string
	Architecture string
	Os           string
	Variant      string
	Size         int
}

type hubTagResponse struct {
	Count    int            `json:"count"`
	Next     string         `json:"next,omitempty"`
	Previous string         `json:"previous,omitempty"`
	Results  []hubTagResult `json:"results,omitempty"`
}

type hubTagResult struct {
	Creator             int           `json:"creator"`
	ID                  int           `json:"id"`
	Name                string        `json:"name"`
	ImageID             string        `json:"image_id,omitempty"`
	LastUpdated         time.Time     `json:"last_updated"`
	LastUpdater         int           `json:"last_updater"`
	LastUpdaterUserName string        `json:"last_updater_username"`
	Images              []hubTagImage `json:"images,omitempty"`
	Repository          int           `json:"repository"`
	FullSize            int           `json:"full_size"`
	V2                  bool          `json:"v2"`
}

type hubTagImage struct {
	Architecture string `json:"architecture"`
	Os           string `json:"os"`
	Features     string `json:"features,omitempty"`
	Variant      string `json:"variant,omitempty"`
	Digest       string `json:"digest"`
	OsFeatures   string `json:"os_features,omitempty"`
	OsVersion    string `json:"os_version,omitempty"`
	Size         int    `json:"size"`
}

func getRepoPath(s string) (string, error) {
	ref, err := reference.ParseNormalizedNamed(s)
	if err != nil {
		return "", err
	}
	ref = reference.TagNameOnly(ref)
	ref = reference.TrimNamed(ref)
	return reference.Path(ref), nil
}

func toImages(result []hubTagImage) []Image {
	images := make([]Image, len(result))
	for i := range result {
		images[i] = Image{
			Digest:       result[i].Digest,
			Architecture: result[i].Architecture,
			Os:           result[i].Os,
			Variant:      result[i].Variant,
			Size:         result[i].Size,
		}
	}
	return images
}
