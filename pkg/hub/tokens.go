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
	"time"

	"github.com/google/uuid"
)

const (
	// TokensURL path to the Hub API listing the Personal Access Tokens
	TokensURL = "/v2/api_tokens"
	// TokenURL path to the Hub API Personal Access Token
	TokenURL = "/v2/api_tokens/%s"
)

// Token is a personal access token. The token field will only be filled at creation and can never been accessed again.
type Token struct {
	UUID        uuid.UUID
	ClientID    string
	CreatorIP   string
	CreatorUA   string
	CreatedAt   time.Time
	LastUsed    time.Time
	GeneratedBy string
	IsActive    bool
	Token       string
	Description string
}

// CreateToken creates a Personal Access Token and returns the token field only once
func (c *Client) CreateToken(description string) (*Token, error) {
	data, err := json.Marshal(hubTokenRequest{Description: description})
	if err != nil {
		return nil, err
	}
	body := bytes.NewBuffer(data)
	req, err := http.NewRequest("POST", c.domain+TokensURL, body)
	if err != nil {
		return nil, err
	}
	response, err := c.doRequest(req, withHubToken(c.token))
	if err != nil {
		return nil, err
	}
	var tokenResponse hubTokenResult
	if err := json.Unmarshal(response, &tokenResponse); err != nil {
		return nil, err
	}
	token, err := convertToken(tokenResponse)
	if err != nil {
		return nil, err
	}
	return &token, nil
}

// GetTokens calls the hub repo API and returns all the information on all tokens
func (c *Client) GetTokens() ([]Token, int, error) {
	u, err := url.Parse(c.domain + TokensURL)
	if err != nil {
		return nil, 0, err
	}
	q := url.Values{}
	q.Add("page_size", fmt.Sprintf("%v", itemsPerPage))
	q.Add("page", "1")
	u.RawQuery = q.Encode()

	tokens, total, next, err := c.getTokensPage(u.String())
	if err != nil {
		return nil, 0, err
	}
	if c.fetchAllElements {
		for next != "" {
			pageTokens, _, n, err := c.getTokensPage(next)
			if err != nil {
				return nil, 0, err
			}
			next = n
			tokens = append(tokens, pageTokens...)
		}
	}

	return tokens, total, nil
}

// GetToken calls the hub repo API and returns the information on one token
func (c *Client) GetToken(tokenUUID string) (*Token, error) {
	req, err := http.NewRequest("GET", c.domain+fmt.Sprintf(TokenURL, tokenUUID), nil)
	if err != nil {
		return nil, err
	}
	response, err := c.doRequest(req, withHubToken(c.token))
	if err != nil {
		return nil, err
	}
	var tokenResponse hubTokenResult
	if err := json.Unmarshal(response, &tokenResponse); err != nil {
		return nil, err
	}
	token, err := convertToken(tokenResponse)
	if err != nil {
		return nil, err
	}
	return &token, nil
}

// UpdateToken updates a token's description and activeness
func (c *Client) UpdateToken(tokenUUID, description string, isActive bool) (*Token, error) {
	tokenRequest := hubTokenRequest{IsActive: isActive}
	if description != "" {
		tokenRequest.Description = description
	}
	data, err := json.Marshal(tokenRequest)
	if err != nil {
		return nil, err
	}
	body := bytes.NewBuffer(data)
	req, err := http.NewRequest("PATCH", c.domain+fmt.Sprintf(TokenURL, tokenUUID), body)
	if err != nil {
		return nil, err
	}
	response, err := c.doRequest(req, withHubToken(c.token))
	if err != nil {
		return nil, err
	}
	var tokenResponse hubTokenResult
	if err := json.Unmarshal(response, &tokenResponse); err != nil {
		return nil, err
	}
	token, err := convertToken(tokenResponse)
	if err != nil {
		return nil, err
	}
	return &token, nil
}

// RemoveToken deletes a token from personal access token
func (c *Client) RemoveToken(tokenUUID string) error {
	//DELETE https://hub.docker.com/v2/api_tokens/8208674e-d08a-426f-b6f4-e3aba7058459 => 202
	req, err := http.NewRequest("DELETE", c.domain+fmt.Sprintf(TokenURL, tokenUUID), nil)
	if err != nil {
		return err
	}
	_, err = c.doRequest(req, withHubToken(c.token))
	return err
}

func (c *Client) getTokensPage(url string) ([]Token, int, string, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, 0, "", err
	}
	response, err := c.doRequest(req, withHubToken(c.token))
	if err != nil {
		return nil, 0, "", err
	}
	var hubResponse hubTokenResponse
	if err := json.Unmarshal(response, &hubResponse); err != nil {
		return nil, 0, "", err
	}
	var tokens []Token
	for _, result := range hubResponse.Results {
		token, err := convertToken(result)
		if err != nil {
			return nil, 0, "", err
		}
		tokens = append(tokens, token)
	}
	return tokens, hubResponse.Count, hubResponse.Next, nil
}

type hubTokenRequest struct {
	Description string `json:"token_label,omitempty"`
	IsActive    bool   `json:"is_active"`
}

type hubTokenResponse struct {
	Count    int              `json:"count"`
	Next     string           `json:"next,omitempty"`
	Previous string           `json:"previous,omitempty"`
	Results  []hubTokenResult `json:"results,omitempty"`
}

type hubTokenResult struct {
	UUID        string    `json:"uuid"`
	ClientID    string    `json:"client_id"`
	CreatorIP   string    `json:"creator_ip"`
	CreatorUA   string    `json:"creator_ua"`
	CreatedAt   time.Time `json:"created_at"`
	LastUsed    time.Time `json:"last_used,omitempty"`
	GeneratedBy string    `json:"generated_by"`
	IsActive    bool      `json:"is_active"`
	Token       string    `json:"token"`
	TokenLabel  string    `json:"token_label"`
}

func convertToken(response hubTokenResult) (Token, error) {
	u, err := uuid.Parse(response.UUID)
	if err != nil {
		return Token{}, err
	}
	return Token{
		UUID:        u,
		ClientID:    response.ClientID,
		CreatorIP:   response.CreatorIP,
		CreatorUA:   response.CreatorUA,
		CreatedAt:   response.CreatedAt,
		LastUsed:    response.LastUsed,
		GeneratedBy: response.GeneratedBy,
		IsActive:    response.IsActive,
		Token:       response.Token,
		Description: response.TokenLabel,
	}, nil
}
