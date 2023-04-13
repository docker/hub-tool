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
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"path"
	"strings"
	"time"
)

const (
	// RepositoriesURL is the Hub API base URL
	RepositoriesURL = "/v2/repositories/"
)

// Repository represents a Docker Hub repository
type Repository struct {
	Name        string
	Description string
	LastUpdated time.Time
	PullCount   int
	StarCount   int
	IsPrivate   bool
}

// GetRepositories lists all the repositories a user can access
func (c *Client) GetRepositories(account string) ([]Repository, int, error) {
	if account == "" {
		account = c.account
	}
	repositoriesURL := fmt.Sprintf("%s%s%s", c.domain, RepositoriesURL, account)
	u, err := url.Parse(repositoriesURL)
	if err != nil {
		return nil, 0, err
	}
	q := url.Values{}
	q.Add("page_size", fmt.Sprintf("%v", itemsPerPage))
	q.Add("page", "1")
	q.Add("ordering", "last_updated")
	u.RawQuery = q.Encode()

	repos, total, next, err := c.getRepositoriesPage(u.String(), account)
	if err != nil {
		return nil, 0, err
	}

	if c.fetchAllElements {
		for next != "" {
			pageRepos, _, n, err := c.getRepositoriesPage(next, account)
			if err != nil {
				return nil, 0, err
			}
			next = n
			repos = append(repos, pageRepos...)
		}
	}

	return repos, total, nil
}

// RemoveRepository removes a repository on Hub
func (c *Client) RemoveRepository(repository string) error {
	repositoryURL := fmt.Sprintf("%s%s%s/", c.domain, RepositoriesURL, repository)
	req, err := http.NewRequest(http.MethodDelete, repositoryURL, nil)
	if err != nil {
		return err
	}
	_, err = c.doRequest(req, withHubToken(c.token))
	if err != nil {
		return err
	}

	return nil
}

// CreateRepository creates a repository on Hub
func (c *Client) CreateRepository(
	repository string,
	description string,
	isPrivate bool,
) (*Repository, error) {
	namespace := path.Dir(repository)
	name := path.Base(repository)
	data, err := json.Marshal(hubCreateRepositoryRequest{
		Registry:    "docker",
		Namespace:   namespace,
		Name:        name,
		Description: description,
		IsPrivate:   isPrivate,
	})
	if err != nil {
		return nil, err
	}
	body := bytes.NewBuffer(data)
	createRepositoryURL := fmt.Sprintf("%s%s", c.domain, RepositoriesURL)
	req, err := http.NewRequest(http.MethodPost, createRepositoryURL, body)
	if err != nil {
		return nil, err
	}
	response, err := c.doRequest(req, withHubToken(c.token))
	if err != nil {
		return nil, err
	}
	var repositoryResult hubRepositoryResult
	if err := json.Unmarshal(response, &repositoryResult); err != nil {
		return nil, err
	}
	ret := convertRepository(repositoryResult)
	return &ret, nil
}

// ToggleScanning turns image scanning on/off
func (c *Client) ToggleScanning(repository string, scanEnabled bool) error {
	enableScanningURL := fmt.Sprintf("%s%s%s", c.domain, "/api/scan/v1/accounts/", repository)
	data, err := json.Marshal(map[string]bool{"scan_enabled": scanEnabled})
	if err != nil {
		return err
	}
	body := bytes.NewBuffer(data)
	req, err := http.NewRequest(http.MethodPost, enableScanningURL, body)
	if err != nil {
		return err
	}
	res, err := c.doRequest(req, withHubToken(c.token))
	if err != nil {
		return fmt.Errorf("%s, body: %s", err, res)
	}
	return nil
}

// ConfigurePermission configures a team permission on a repository
func (c *Client) ConfigurePermission(repository string, team string, permission PermissionType) error {
	org := strings.Split(repository, "/")[0]
	teams, err := c.GetTeams(org)
	if err != nil {
		return err
	}
	var teamID int
	for i := range teams {
		if teams[i].Name == team {
			teamID = teams[i].ID
		}
	}
	fmt.Println(teamID)

	configurePermissionsURL := fmt.Sprintf("%s%s%s%s", c.domain, RepositoriesURL, repository, "/groups")
	data, err := json.Marshal(map[string]interface{}{"group_id": teamID, "permission": permission})
	if err != nil {
		return err
	}
	body := bytes.NewBuffer(data)
	req, err := http.NewRequest(http.MethodPost, configurePermissionsURL, body)
	if err != nil {
		return err
	}
	res, err := c.doRequest(req, withHubToken(c.token))
	if err != nil {
		return fmt.Errorf("%s, body: %s", err, res)
	}
	return nil
}

func (c *Client) getRepositoriesPage(url, account string) ([]Repository, int, string, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, 0, "", err
	}
	response, err := c.doRequest(req, withHubToken(c.token))
	if err != nil {
		return nil, 0, "", err
	}
	var hubResponse hubRepositoryResponse
	if err := json.Unmarshal(response, &hubResponse); err != nil {
		return nil, 0, "", err
	}
	var repos []Repository
	for _, result := range hubResponse.Results {
		repo := convertRepository(result)
		repos = append(repos, repo)
	}
	return repos, hubResponse.Count, hubResponse.Next, nil
}

type hubRepositoryResponse struct {
	Count    int                   `json:"count"`
	Next     string                `json:"next,omitempty"`
	Previous string                `json:"previous,omitempty"`
	Results  []hubRepositoryResult `json:"results,omitempty"`
}

type hubRepositoryResult struct {
	Name           string         `json:"name"`
	Namespace      string         `json:"namespace"`
	PullCount      int            `json:"pull_count"`
	StarCount      int            `json:"star_count"`
	RepositoryType RepositoryType `json:"repository_type"`
	CanEdit        bool           `json:"can_edit"`
	Description    string         `json:"description,omitempty"`
	IsAutomated    bool           `json:"is_automated"`
	IsMigrated     bool           `json:"is_migrated"`
	IsPrivate      bool           `json:"is_private"`
	LastUpdated    time.Time      `json:"last_updated"`
	Status         int            `json:"status"`
	User           string         `json:"user"`
}

func convertRepository(res hubRepositoryResult) Repository {
	return Repository{
		Name:        fmt.Sprintf("%s/%s", res.Namespace, res.Name),
		Description: res.Description,
		LastUpdated: res.LastUpdated,
		PullCount:   res.PullCount,
		StarCount:   res.StarCount,
		IsPrivate:   res.IsPrivate,
	}
}

type hubCreateRepositoryRequest struct {
	Registry    string `json:"registry"`
	Namespace   string `json:"namespace"`
	Description string `json:"description"`
	Name        string `json:"name"`
	IsPrivate   bool   `json:"is_private"`
}

// RepositoryType lists all the different repository types handled by the Docker Hub
type RepositoryType string

const (
	//ImageType is the classic image type
	ImageType = RepositoryType("image")
)

type PermissionType string

const (
	ReadOnly  PermissionType = "read"
	ReadWrite PermissionType = "write"
)
