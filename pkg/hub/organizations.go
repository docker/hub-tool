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
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"sort"
	"time"

	"golang.org/x/sync/errgroup"
)

const (
	// OrganizationsURL path to the Hub API listing the organizations
	OrganizationsURL = "/v2/user/orgs/"
	// OrganizationInfoURL path to the Hub API returning organization info
	OrganizationInfoURL = "/v2/orgs/%s"
)

// Organization represents a Docker Hub organization
type Organization struct {
	Namespace string
	FullName  string
	Role      string
	Teams     []Team
	Members   []Member
}

// GetOrganizations lists all the organizations a user has joined
func (c *Client) GetOrganizations(ctx context.Context) ([]Organization, error) {
	u, err := url.Parse(c.domain + OrganizationsURL)
	if err != nil {
		return nil, err
	}
	q := url.Values{}
	q.Add("page_size", fmt.Sprintf("%v", itemsPerPage))
	q.Add("page", "1")
	u.RawQuery = q.Encode()

	organizations, next, err := c.getOrganizationsPage(ctx, u.String())
	if err != nil {
		return nil, err
	}

	for next != "" {
		pageOrganizations, n, err := c.getOrganizationsPage(ctx, next)
		if err != nil {
			return nil, err
		}
		next = n
		organizations = append(organizations, pageOrganizations...)
	}

	return organizations, nil
}

// GetOrganizationInfo returns organization info
func (c *Client) GetOrganizationInfo(orgname string) (*Account, error) {
	u, err := url.Parse(c.domain + fmt.Sprintf(OrganizationInfoURL, orgname))
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequest("GET", u.String(), nil)
	if err != nil {
		return nil, err
	}
	response, err := c.doRequest(req, withHubToken(c.token))
	if err != nil {
		return nil, err
	}
	var hubResponse hubOrgInfoResponse
	if err := json.Unmarshal(response, &hubResponse); err != nil {
		return nil, err
	}

	return &Account{
		ID:       hubResponse.ID,
		Name:     hubResponse.OrgName,
		FullName: hubResponse.FullName,
		Location: hubResponse.Location,
		Company:  hubResponse.Company,
		Joined:   hubResponse.DateJoined,
	}, nil
}

func (c *Client) getOrganizationsPage(ctx context.Context, url string) ([]Organization, string, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, "", err
	}
	req = req.WithContext(ctx)
	response, err := c.doRequest(req, withHubToken(c.token))
	if err != nil {
		return nil, "", err
	}
	var hubResponse hubOrganizationResponse
	if err := json.Unmarshal(response, &hubResponse); err != nil {
		return nil, "", err
	}

	var organizations []Organization
	eg, _ := errgroup.WithContext(ctx)

	for _, result := range hubResponse.Results {
		result := result
		eg.Go(func() error {
			var (
				teams   []Team
				members []Member
			)
			subeg, _ := errgroup.WithContext(ctx)

			subeg.Go(func() error {
				teams, err = c.GetTeams(result.OrgName)
				return err
			})
			subeg.Go(func() error {
				members, err = c.GetMembers(result.OrgName)
				return err
			})

			if err := subeg.Wait(); err != nil {
				return err
			}
			organization := Organization{
				Namespace: result.OrgName,
				FullName:  result.FullName,
				Role:      getRole(teams),
				Teams:     teams,
				Members:   members,
			}
			organizations = append(organizations, organization)

			return nil
		})
	}

	if err := eg.Wait(); err != nil {
		return []Organization{}, "", err
	}

	sort.Slice(organizations, func(i, j int) bool {
		return organizations[i].Namespace < organizations[j].Namespace
	})
	return organizations, hubResponse.Next, nil
}

func getRole(teams []Team) string {
	for _, t := range teams {
		if t.Name == "owners" {
			return "Owner"
		}
	}
	return "Member"
}

type hubOrganizationResponse struct {
	Count    int                     `json:"count"`
	Next     string                  `json:"next,omitempty"`
	Previous string                  `json:"previous,omitempty"`
	Results  []hubOrganizationResult `json:"results,omitempty"`
}

type hubOrganizationResult struct {
	OrgName       string    `json:"orgname"`
	FullName      string    `json:"full_name"`
	Company       string    `json:"company"`
	Location      string    `json:"location"`
	Type          string    `json:"type"`
	DateJoined    time.Time `json:"date_joined"`
	GravatarEmail string    `json:"gravatar_email"`
	GravatarURL   string    `json:"gravatar_url"`
	ProfileURL    string    `json:"profile_url"`
	ID            string    `json:"id"`
}

type hubOrgInfoResponse struct {
	ID            string    `json:"id"`
	OrgName       string    `json:"orgname"`
	FullName      string    `json:"full_name"`
	Location      string    `json:"location"`
	Company       string    `json:"company"`
	GravatarEmail string    `json:"gravatar_email"`
	GravatarURL   string    `json:"gravatar_url"`
	ProfileURL    string    `json:"profile_url"`
	DateJoined    time.Time `json:"date_joined"`
	Type          string    `json:"type"`
}
