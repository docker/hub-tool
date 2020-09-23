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
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"
)

const (
	//MembersURL path to the Hub API listing the members in an organization
	MembersURL = "/v2/orgs/%s/members/"
)

//Member is a user part of an organization
type Member struct {
	Username string
	FullName string
}

//GetMembers lists all the members in an organization
func (h *Client) GetMembers(organization string) ([]Member, error) {
	u, err := url.Parse(h.domain + fmt.Sprintf(MembersURL, organization))
	if err != nil {
		return nil, err
	}
	q := url.Values{}
	q.Add("page_size", fmt.Sprintf("%v", itemsPerPage))
	q.Add("page", "1")
	u.RawQuery = q.Encode()

	members, next, err := h.getMembersPage(u.String())
	if err != nil {
		return nil, err
	}

	for next != "" {
		pageMembers, n, err := h.getMembersPage(next)
		if err != nil {
			return nil, err
		}
		next = n
		members = append(members, pageMembers...)
	}

	return members, nil
}

func (h *Client) getMembersPage(url string) ([]Member, string, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, "", err
	}
	req.Header["Authorization"] = []string{fmt.Sprintf("Bearer %s", h.token)}
	response, err := doRequest(req)
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
