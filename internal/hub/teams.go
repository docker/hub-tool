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
)

const (
	//GroupsURL path to the Hub API listing the groups in an organization
	GroupsURL = "/v2/orgs/%s/groups/"
)

//Team represents a hub group in an organization
type Team struct {
	Name        string
	Description string
	Members     []Member
}

//GetTeams lists all the teams in an organization
func (c *Client) GetTeams(organization string) ([]Team, error) {
	u, err := url.Parse(c.domain + fmt.Sprintf(GroupsURL, organization))
	if err != nil {
		return nil, err
	}
	q := url.Values{}
	q.Add("page_size", fmt.Sprintf("%v", itemsPerPage))
	q.Add("page", "1")
	u.RawQuery = q.Encode()

	teams, next, err := c.getTeamsPage(u.String(), organization)
	if err != nil {
		return nil, err
	}

	for next != "" {
		pageTeams, n, err := c.getTeamsPage(next, organization)
		if err != nil {
			return nil, err
		}
		next = n
		teams = append(teams, pageTeams...)
	}

	return teams, nil
}

func (c *Client) getTeamsPage(url, organization string) ([]Team, string, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, "", err
	}
	response, err := c.doRequest(req, WithHubToken(c.token))
	if err != nil {
		return nil, "", err
	}
	var hubResponse hubGroupResponse
	if err := json.Unmarshal(response, &hubResponse); err != nil {
		return nil, "", err
	}
	var teams []Team
	for _, result := range hubResponse.Results {
		members, err := c.GetMembersPerTeam(organization, result.Name)
		if err != nil {
			return nil, "", err
		}
		team := Team{
			Name:        result.Name,
			Description: result.Description,
			Members:     members,
		}
		teams = append(teams, team)
	}
	return teams, hubResponse.Next, nil
}

type hubGroupResponse struct {
	Count    int              `json:"count"`
	Next     string           `json:"next,omitempty"`
	Previous string           `json:"previous,omitempty"`
	Results  []hubGroupResult `json:"results,omitempty"`
}

type hubGroupResult struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	ID          int    `json:"id"`
}
