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
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/registry"
)

const (
	// LoginURL path to the Hub login URL
	LoginURL = "/v2/users/login"
	// TagsURL path to the Hub provider token generation URL
	TagsURL = "/v2/repositories/%s/tags/?page_size=25&page=1"
	// DeleteTagURL path to the Hub API to remove a tag
	DeleteTagURL = "/v2/repositories/%s/tags/%s/"

	tagsPerPage = 25
)

//Client sends authenticated calls to the Hub API
type Client struct {
	domain string
	token  string
}

//AuthResolver resolves authentication configuration depending the registry
type AuthResolver func(*registry.IndexInfo) types.AuthConfig

//NewClient logs the user to the hub and returns a client which can send authenticated requests
// to the Hub API
func NewClient(authResolver AuthResolver) (*Client, error) {
	hubInstance := getInstance()
	hubAuthConfig := authResolver(hubInstance.RegistryInfo)
	token, err := login(hubInstance.APIHubBaseURL, hubAuthConfig)
	if err != nil {
		return nil, err
	}
	return &Client{
		domain: hubInstance.APIHubBaseURL,
		token:  token,
	}, nil
}

//GetTags calls the hub repo API and returns all the information on all tags
func (h *Client) GetTags(repository string) ([]Tag, error) {
	repoPath, err := getRepoPath(repository)
	if err != nil {
		return nil, err
	}
	u, err := url.Parse(h.domain + fmt.Sprintf(TagsURL, repoPath))
	if err != nil {
		return nil, err
	}
	q := url.Values{}
	q.Add("page_size", fmt.Sprintf("%v", tagsPerPage))
	q.Add("page", "1")
	u.RawQuery = q.Encode()

	tags, next, err := h.getTagsPage(u.String())
	if err != nil {
		return nil, err
	}

	for next != "" {
		pageTags, n, err := h.getTagsPage(next)
		if err != nil {
			return nil, err
		}
		next = n
		tags = append(tags, pageTags...)
	}

	return tags, nil
}

//RemoveTag removes a tag in a repository on Hub
func (h *Client) RemoveTag(repository, tag string) error {
	req, err := http.NewRequest("DELETE", h.domain+fmt.Sprintf(DeleteTagURL, repository, tag), nil)
	fmt.Println(req.URL)
	if err != nil {
		return err
	}
	req.Header["Authorization"] = []string{fmt.Sprintf("Bearer %s", h.token)}
	_, err = doRequest(req)
	return err
}

func (h *Client) getTagsPage(url string) ([]Tag, string, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, "", err
	}
	req.Header["Authorization"] = []string{fmt.Sprintf("Bearer %s", h.token)}
	response, err := doRequest(req)
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
			Name:                result.Name,
			FullSize:            result.FullSize,
			LastUpdated:         result.LastUpdated,
			LastUpdaterUserName: result.LastUpdaterUserName,
			Images:              toImages(result.Images),
		}
		tags = append(tags, tag)
	}
	return tags, hubResponse.Next, nil
}

func login(hubBaseURL string, hubAuthConfig types.AuthConfig) (string, error) {
	data, err := json.Marshal(hubAuthConfig)
	if err != nil {
		return "", err
	}
	body := bytes.NewBuffer(data)

	// Login on the Docker Hub
	req, err := http.NewRequest("POST", hubBaseURL+LoginURL, ioutil.NopCloser(body))
	if err != nil {
		return "", err
	}
	req.Header["Content-Type"] = []string{"application/json"}
	buf, err := doRequest(req)
	if err != nil {
		return "", err
	}

	creds := struct {
		Token string `json:"token"`
	}{}
	if err := json.Unmarshal(buf, &creds); err != nil {
		return "", err
	}
	return creds.Token, nil
}

func doRequest(req *http.Request) ([]byte, error) {
	req.Header["Accept"] = []string{"application/json"}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	if resp.Body != nil {
		defer resp.Body.Close() //nolint:errcheck
	}
	if resp.StatusCode < 200 && resp.StatusCode >= 300 {
		return nil, fmt.Errorf("bad status code %q", resp.Status)
	}
	buf, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	return buf, nil
}
