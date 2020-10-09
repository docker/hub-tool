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
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/docker/distribution/reference"
)

const (
	// TagsURL path to the Hub API listing the tags
	TagsURL = "/v2/repositories/%s/tags/"
	// DeleteTagURL path to the Hub API to remove a tag
	DeleteTagURL = "/v2/repositories/%s/tags/%s/"
)

//Tag can point to a manifest or manifest list
type Tag struct {
	Name                string
	FullSize            int
	LastUpdated         time.Time
	LastUpdaterUserName string
	Images              []Image
	Expires             time.Time
	LastPulled          time.Time
	LastPushed          time.Time
	Status              string
}

//Image represents the metadata of a manifest
type Image struct {
	Digest       string
	Architecture string
	Os           string
	Variant      string
	Size         int
	Expires      time.Time
	LastPulled   time.Time
	LastPushed   time.Time
	Status       string
}

//GetTags calls the hub repo API and returns all the information on all tags
func (c *Client) GetTags(repository string, reqOps ...RequestOp) ([]Tag, error) {
	repoPath, err := getRepoPath(repository)
	if err != nil {
		return nil, err
	}
	u, err := url.Parse(c.domain + fmt.Sprintf(TagsURL, repoPath))
	if err != nil {
		return nil, err
	}
	q := url.Values{}
	q.Add("page_size", fmt.Sprintf("%v", itemsPerPage))
	q.Add("page", "1")
	u.RawQuery = q.Encode()

	tags, next, err := c.getTagsPage(u.String(), repository, reqOps...)
	if err != nil {
		return nil, err
	}
	if c.fetchAllElements {
		for next != "" {
			pageTags, n, err := c.getTagsPage(next, repository)
			if err != nil {
				return nil, err
			}
			next = n
			tags = append(tags, pageTags...)
		}
	}

	return tags, nil
}

//RemoveTag removes a tag in a repository on Hub
func (c *Client) RemoveTag(repository, tag string) error {
	req, err := http.NewRequest("DELETE", c.domain+fmt.Sprintf(DeleteTagURL, repository, tag), nil)
	if err != nil {
		return err
	}
	_, err = doRequest(req, WithHubToken(c.token))
	return err
}

func (c *Client) getTagsPage(url, repository string, reqOps ...RequestOp) ([]Tag, string, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, "", err
	}
	response, err := doRequest(req, append(reqOps, WithHubToken(c.token))...)
	if err != nil {
		return nil, "", err
	}
	var hubResponse hubTagResponse
	if err := json.Unmarshal(response, &hubResponse); err != nil {
		return nil, "", err
	}
	var tags []Tag
	for _, result := range hubResponse.Results {
		tag := Tag{
			Name:                fmt.Sprintf("%s:%s", repository, result.Name),
			FullSize:            result.FullSize,
			LastUpdated:         result.LastUpdated,
			LastUpdaterUserName: result.LastUpdaterUserName,
			Images:              toImages(result.Images),
			Status:              result.Status,
			LastPulled:          result.LastPulled,
			LastPushed:          result.LastPushed,
			Expires:             result.Expires,
		}
		tags = append(tags, tag)
	}
	return tags, hubResponse.Next, nil
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
	// New API
	Expires    time.Time `json:"tag_expires,omitempty"`
	LastPulled time.Time `json:"tag_last_pulled,omitempty"`
	LastPushed time.Time `json:"tag_last_pushed,omitempty"`
	Status     string    `json:"tag_status,omitempty"`
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
	// New API
	Expires    time.Time `json:"expires,omitempty"`
	LastPulled time.Time `json:"last_pulled,omitempty"`
	LastPushed time.Time `json:"last_pushed,omitempty"`
	Status     string    `json:"status,omitempty"`
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
			Status:       result[i].Status,
			Expires:      result[i].Expires,
			LastPulled:   result[i].LastPulled,
			LastPushed:   result[i].LastPushed,
		}
	}
	return images
}
