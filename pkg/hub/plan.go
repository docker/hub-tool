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
	//HubPlanURL path to the billing API returning the account hub plan
	HubPlanURL = "/api/billing/v4/accounts/%s/hub-plan"
	//TeamPlan refers to a hub team paid account
	TeamPlan = "team"
	//ProPlan refers to a hub individual paid account
	ProPlan = "pro"
	//FreePlan refers to a hub non-paid account
	FreePlan = "free"
)

// Plan represents the current account Hub plan
type Plan struct {
	Name   string
	Limits Limits
}

// Limits represents the current account limits
type Limits struct {
	Seats          int
	PrivateRepos   int
	Teams          int
	Collaborators  int
	ParallelBuilds int
}

// GetHubPlan returns an account current Hub plan
func (c *Client) GetHubPlan(accountID string) (*Plan, error) {
	u, err := url.Parse(c.domain + fmt.Sprintf(HubPlanURL, accountID))
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
	var hubResponse hubPlanResponse
	if err := json.Unmarshal(response, &hubResponse); err != nil {
		return nil, err
	}
	return &Plan{
		Name: hubResponse.Name,
		Limits: Limits{
			Seats:          hubResponse.Seats,
			PrivateRepos:   hubResponse.PrivateRepos,
			Teams:          hubResponse.Teams,
			Collaborators:  hubResponse.Collaborators,
			ParallelBuilds: hubResponse.ParallelBuilds,
		},
	}, nil
}

type hubPlanResponse struct {
	Name           string `json:"name"`
	Legacy         bool   `json:"legacy"`
	Seats          int    `json:"seats"`
	PrivateRepos   int    `json:"private_repos"`
	Teams          int    `json:"teams"`
	Collaborators  int    `json:"collaborators"`
	ParallelBuilds int    `json:"parallel_builds"`
	Duration       string `json:"duration"`
}
