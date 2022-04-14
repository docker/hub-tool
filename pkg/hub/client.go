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
	"io"
	"io/ioutil"
	"net/http"
	"net/url"

	"github.com/docker/docker/api/types"
	log "github.com/sirupsen/logrus"

	"github.com/docker/hub-tool/internal"
)

const (
	// LoginURL path to the Hub login URL
	LoginURL = "/v2/users/login?refresh_token=true"
	// TwoFactorLoginURL path to the 2FA
	TwoFactorLoginURL = "/v2/users/2fa-login?refresh_token=true"
	// SecondFactorDetailMessage returned by login if 2FA is enabled
	SecondFactorDetailMessage = "Require secondary authentication on MFA enabled account"

	itemsPerPage = 100
)

//Client sends authenticated calls to the Hub API
type Client struct {
	AuthConfig types.AuthConfig
	Ctx        context.Context

	client           *http.Client
	domain           string
	token            string
	refreshToken     string
	password         string
	account          string
	fetchAllElements bool
	in               io.Reader
	out              io.Writer
}

type twoFactorResponse struct {
	Detail        string `json:"detail"`
	Login2FAToken string `json:"login_2fa_token"`
}

type twoFactorRequest struct {
	Code          string `json:"code"`
	Login2FAToken string `json:"login_2fa_token"`
}

type tokenResponse struct {
	Detail       string `json:"detail"`
	Token        string `json:"token"`
	RefreshToken string `json:"refresh_token"`
}

//ClientOp represents an option given to NewClient constructor to customize client behavior.
type ClientOp func(*Client) error

//RequestOp represents an option to customize the request sent to the Hub API
type RequestOp func(r *http.Request) error

//NewClient logs the user to the hub and returns a client which can send authenticated requests
// to the Hub API
func NewClient(ops ...ClientOp) (*Client, error) {
	hubInstance := getInstance()

	client := &Client{
		client: http.DefaultClient,
		domain: hubInstance.APIHubBaseURL,
	}
	for _, op := range ops {
		if err := op(client); err != nil {
			return nil, err
		}
	}

	return client, nil
}

//Update changes client behavior using ClientOp
func (c *Client) Update(ops ...ClientOp) error {
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

//WithInStream sets the input stream
func WithInStream(in io.Reader) ClientOp {
	return func(c *Client) error {
		c.in = in
		return nil
	}
}

//WithOutStream sets the output stream
func WithOutStream(out io.Writer) ClientOp {
	return func(c *Client) error {
		c.out = out
		return nil
	}
}

// WithHubAccount sets the current account name
func WithHubAccount(account string) ClientOp {
	return func(c *Client) error {
		c.AuthConfig.Username = account
		c.account = account
		return nil
	}
}

// WithHubToken sets the bearer token to the client
func WithHubToken(token string) ClientOp {
	return func(c *Client) error {
		c.token = token
		return nil
	}
}

// WithRefreshToken sets the refresh token to the client
func WithRefreshToken(refreshToken string) ClientOp {
	return func(c *Client) error {
		c.refreshToken = refreshToken
		return nil
	}
}

// WithPassword sets the password to the client
func WithPassword(password string) ClientOp {
	return func(c *Client) error {
		c.password = password
		return nil
	}
}

// WithHTTPClient sets the *http.Client for the client
func WithHTTPClient(client *http.Client) ClientOp {
	return func(c *Client) error {
		c.client = client
		return nil
	}
}

func withHubToken(token string) RequestOp {
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

// Login tries to authenticate, it will call the twoFactorCodeProvider if the
// user has 2FA activated
func (c *Client) Login(username string, password string, twoFactorCodeProvider func() (string, error)) (string, string, error) {
	data, err := json.Marshal(types.AuthConfig{
		Username: username,
		Password: password,
	})
	if err != nil {
		return "", "", err
	}
	body := bytes.NewBuffer(data)

	// Login on the Docker Hub
	req, err := http.NewRequest("POST", c.domain+LoginURL, body)
	if err != nil {
		return "", "", err
	}
	resp, err := c.doRawRequest(req)
	if err != nil {
		return "", "", err
	}
	defer func() { _ = resp.Body.Close() }()
	buf, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", "", err
	}

	// Login is OK, return the token
	if resp.StatusCode == http.StatusOK {
		creds := tokenResponse{}
		if err := json.Unmarshal(buf, &creds); err != nil {
			return "", "", err
		}
		return creds.Token, "", nil
	} else if resp.StatusCode == http.StatusUnauthorized {
		response2FA := twoFactorResponse{}
		if err := json.Unmarshal(buf, &response2FA); err != nil {
			return "", "", err
		}
		// Check if 2FA is enabled and needs a second authentication
		if response2FA.Detail != SecondFactorDetailMessage {
			return "", "", fmt.Errorf(response2FA.Detail)
		}
		return c.getTwoFactorToken(response2FA.Login2FAToken, twoFactorCodeProvider)
	}
	if ok, err := extractError(buf, resp); ok {
		return "", "", err
	}
	return "", "", fmt.Errorf("failed to authenticate: bad status code %q: %s", resp.Status, string(buf))
}

func (c *Client) getTwoFactorToken(token string, twoFactorCodeProvider func() (string, error)) (string, string, error) {
	code, err := twoFactorCodeProvider()
	if err != nil {
		return "", "", err
	}

	body2FA := twoFactorRequest{
		Code:          code,
		Login2FAToken: token,
	}
	data, err := json.Marshal(body2FA)
	if err != nil {
		return "", "", err
	}

	body := bytes.NewBuffer(data)

	// Request 2FA on the Docker Hub
	req, err := http.NewRequest("POST", c.domain+TwoFactorLoginURL, body)
	if err != nil {
		return "", "", err
	}
	resp, err := c.doRawRequest(req)
	if err != nil {
		return "", "", err
	}
	defer func() { _ = resp.Body.Close() }()

	buf, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", "", err
	}

	// Login is OK, return the token
	if resp.StatusCode == http.StatusOK {
		creds := tokenResponse{}
		if err := json.Unmarshal(buf, &creds); err != nil {
			return "", "", err
		}

		return creds.Token, creds.RefreshToken, nil
	}

	return "", "", fmt.Errorf("failed to authenticate: bad status code %q: %s", resp.Status, string(buf))
}

func (c *Client) doRequest(req *http.Request, reqOps ...RequestOp) ([]byte, error) {
	log.Debugf("HTTP %s on: %s", req.Method, req.URL)
	log.Tracef("HTTP request: %+v", req)
	resp, err := c.doRawRequest(req, reqOps...)
	if err != nil {
		return nil, err
	}
	if resp.Body != nil {
		defer resp.Body.Close() //nolint:errcheck
	}
	log.Tracef("HTTP response: %+v", resp)
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		if resp.StatusCode == http.StatusForbidden {
			return nil, &forbiddenError{}
		}
		buf, err := ioutil.ReadAll(resp.Body)
		log.Debugf("bad status code %q: %s", resp.Status, buf)
		if err == nil {
			if ok, err := extractError(buf, resp); ok {
				return nil, err
			}
		}
		return nil, fmt.Errorf("bad status code %q", resp.Status)
	}
	buf, err := ioutil.ReadAll(resp.Body)
	log.Tracef("HTTP response body: %s", buf)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("bad status code %q: %s", resp.Status, string(buf))
	}

	return buf, nil
}

func (c *Client) doRawRequest(req *http.Request, reqOps ...RequestOp) (*http.Response, error) {
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
	return c.client.Do(req)
}

func extractError(buf []byte, resp *http.Response) (bool, error) {
	var responseBody map[string]string
	if err := json.Unmarshal(buf, &responseBody); err == nil {
		for _, k := range []string{"message", "detail"} {
			if msg, ok := responseBody[k]; ok {
				return true, fmt.Errorf("failed to authenticate: bad status code %q: %s", resp.Status, msg)
			}
		}
	}
	return false, nil
}
