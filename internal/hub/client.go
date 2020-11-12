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
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/registry"

	"github.com/docker/hub-tool/internal"
)

const (
	// LoginURL path to the Hub login URL
	LoginURL = "/v2/users/login"

	itemsPerPage = 100
)

//Client sends authenticated calls to the Hub API
type Client struct {
	AuthConfig types.AuthConfig
	Ctx        context.Context

	domain           string
	token            string
	fetchAllElements bool
}

//AuthResolver resolves authentication configuration depending the registry
type AuthResolver func(*registry.IndexInfo) types.AuthConfig

//ClientOp represents an option given to NewClient constructor to customize client behavior.
type ClientOp func(*Client) error

//RequestOp represents an option to customize the request sent to the Hub API
type RequestOp func(r *http.Request) error

//NewClient logs the user to the hub and returns a client which can send authenticated requests
// to the Hub API
func NewClient(authResolver AuthResolver, ops ...ClientOp) (*Client, error) {
	hubInstance := getInstance()
	hubAuthConfig := authResolver(hubInstance.RegistryInfo)
	// Check if the user is logged in
	if hubAuthConfig.Username == "" {
		return nil, &authenticationError{}
	}
	client := &Client{
		AuthConfig: hubAuthConfig,
		domain:     hubInstance.APIHubBaseURL,
	}
	for _, op := range ops {
		if err := op(client); err != nil {
			return nil, err
		}
	}
	token, err := client.login(hubInstance.APIHubBaseURL, hubAuthConfig)
	if err != nil {
		return nil, err
	}
	client.token = token
	return client, nil
}

//Apply changes client behavior using ClientOp
func (c *Client) Apply(ops ...ClientOp) error {
	for _, op := range ops {
		if err := op(c); err != nil {
			return err
		}
	}
	return nil
}

//WithAllElements makes the client fetch all the elements it can find, enabling pagination.
func WithAllElements() ClientOp {
	return func(c *Client) error {
		c.fetchAllElements = true
		return nil
	}
}

//WithContext set the client context
func WithContext(ctx context.Context) ClientOp {
	return func(c *Client) error {
		c.Ctx = ctx
		return nil
	}
}

//WithHubToken sets the bearer token to the request
func WithHubToken(token string) RequestOp {
	return func(req *http.Request) error {
		req.Header["Authorization"] = []string{fmt.Sprintf("Bearer %s", token)}
		return nil
	}
}

//WithSortingOrder adds a sorting order query parameter to the request
func WithSortingOrder(order string) RequestOp {
	return func(req *http.Request) error {
		values, err := url.ParseQuery(req.URL.RawQuery)
		if err != nil {
			return err
		}
		values.Add("ordering", order)
		req.URL.RawQuery = values.Encode()
		return nil
	}
}

func (c *Client) login(hubBaseURL string, hubAuthConfig types.AuthConfig) (string, error) {
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
	buf, err := c.doRequest(req)
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

func (c *Client) doRequest(req *http.Request, reqOps ...RequestOp) ([]byte, error) {
	req.Header["Accept"] = []string{"application/json"}
	req.Header["Content-Type"] = []string{"application/json"}
	req.Header["User-Agent"] = []string{fmt.Sprintf("hub-tool/%s", internal.Version)}
	for _, op := range reqOps {
		if err := op(req); err != nil {
			return nil, err
		}
	}
	if c.Ctx != nil {
		req = req.WithContext(c.Ctx)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	if resp.Body != nil {
		defer resp.Body.Close() //nolint:errcheck
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		buf, err := ioutil.ReadAll(resp.Body)
		if err == nil {
			var body map[string]string
			if err := json.Unmarshal(buf, &body); err == nil {
				if msg, ok := body["message"]; ok {
					return nil, fmt.Errorf("bad status code %q: %s", resp.Status, msg)
				}
			}
		}
		return nil, fmt.Errorf("bad status code %q", resp.Status)
	}
	buf, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	return buf, nil
}
