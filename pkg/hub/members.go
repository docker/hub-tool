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
)

const (
	//MembersURL path to the Hub API listing the members in an organization
	MembersURL = "/v2/orgs/%s/members/"
	//MembersPerTeamURL path to the Hub API listing the members in a team
	MembersPerTeamURL = "/v2/orgs/%s/groups/%s/members/"
)

//Member is a user part of an organization
type Member struct {
	Username string `json:"username"`
	FullName string `json:"full_name"`
}

//GetMembers lists all the members in an organization
func (c *Client) GetMembers(organization string) ([]Member, error) {
	u, err := url.Parse(c.domain + fmt.Sprintf(MembersURL, organization))
	if err != nil {
		return nil, err
	}
	q := url.Values{}
	q.Add("page_size", fmt.Sprintf("%v", itemsPerPage))
	q.Add("page", "1")
	u.RawQuery = q.Encode()

	members, next, err := c.getMembersPage(u.String())
	if err != nil {
		return nil, err
	}

	for next != "" {
		pageMembers, n, err := c.getMembersPage(next)
		if err != nil {
			return nil, err
		}
		next = n
		members = append(members, pageMembers...)
	}

	return members, nil
}

// GetMembersCount return the number of members in an organization
func (c *Client) GetMembersCount(organization string) (int, error) {
	u, err := url.Parse(c.domain + fmt.Sprintf(MembersURL, organization))
	if err != nil {
		return 0, err
	}
	q := url.Values{}
	q.Add("page_size", "1")
	q.Add("page", "1")
	u.RawQuery = q.Encode()

	req, err := http.NewRequest("GET", u.String(), nil)
	if err != nil {
		return 0, err
	}
	response, err := c.doRequest(req, withHubToken(c.token))
	if err != nil {
		return 0, err
	}
	var hubResponse hubMemberResponse
	if err := json.Unmarshal(response, &hubResponse); err != nil {
		return 0, err
	}
	return hubResponse.Count, nil
}

// GetMembersPerTeam returns the members of a team in an organization
func (c *Client) GetMembersPerTeam(organization, team string) ([]Member, error) {
	u := c.domain + fmt.Sprintf(MembersPerTeamURL, organization, team)
	req, err := http.NewRequest("GET", u, nil)
	if err != nil {
		return nil, err
	}
	response, err := c.doRequest(req, withHubToken(c.token))
	if err != nil {
		return nil, err
	}
	var members []Member
	if err := json.Unmarshal(response, &members); err != nil {
		return nil, err
	}
	return members, nil
}

func (c *Client) getMembersPage(url string) ([]Member, string, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, "", err
	}
	response, err := c.doRequest(req, withHubToken(c.token))
	if err != nil {
		return nil, "", err
	}
	var hubResponse hubMemberResponse
	if err := json.Unmarshal(response, &hubResponse); err != nil {
		return nil, "", err
	}
	var members []Member
	for _, result := range hubResponse.Results {
		member := Member{
			Username: result.UserName,
			FullName: result.FullName,
		}
		members = append(members, member)
	}
	return members, hubResponse.Next, nil
}

type hubMemberResponse struct {
	Count    int               `json:"count"`
	Next     string            `json:"next,omitempty"`
	Previous string            `json:"previous,omitempty"`
	Results  []hubMemberResult `json:"results,omitempty"`
}

type hubMemberResult struct {
	UserName    string    `json:"username"`
	GravatarURL string    `json:"gravatar_url"`
	FullName    string    `json:"full_name"`
	Company     string    `json:"company"`
	Type        string    `json:"type"`
	Location    string    `json:"location"`
	DateJoined  time.Time `json:"date_joined"`
	ID          string    `json:"id"`
	ProfileURL  string    `json:"profile_url"`
}
