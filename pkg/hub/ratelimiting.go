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
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
)

// RateLimits ...
type RateLimits struct {
	Limit           *int    `json:",omitempty"`
	LimitWindow     *int    `json:",omitempty"`
	Remaining       *int    `json:",omitempty"`
	RemainingWindow *int    `json:",omitempty"`
	Source          *string `json:",omitempty"`
}

var (
	first        = "https://auth.docker.io/token?service=registry.docker.io&scope=repository:ratelimitpreview/test:pull"
	second       = "https://registry-1.docker.io/v2/ratelimitpreview/test/manifests/latest"
	defaultValue = -1
)

// SetURLs change the base urls used to check ratelimiting values
func SetURLs(newFirst, newSecond string) {
	first = newFirst
	second = newSecond
}

// GetRateLimits returns the rate limits for the user
func (c *Client) GetRateLimits() (*RateLimits, error) {
	token, err := tryGetToken(c)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("HEAD", second, nil)
	if err != nil {
		return nil, err
	}

	resp, err := c.doRawRequest(req, withHubToken(token))
	if err != nil {
		return nil, err
	}

	limitHeader := resp.Header.Get("Ratelimit-Limit")
	remainingHeader := resp.Header.Get("Ratelimit-Remaining")
	source := resp.Header.Get("docker-Ratelimit-Source")

	if limitHeader == "" || remainingHeader == "" {
		return &RateLimits{
			Limit:           &defaultValue,
			LimitWindow:     &defaultValue,
			Remaining:       &defaultValue,
			RemainingWindow: &defaultValue,
			Source:          &source,
		}, nil
	}

	limit, limitWindow, err := parseLimitHeader(limitHeader)
	if err != nil {
		return nil, err
	}

	remaining, remainingWindow, err := parseLimitHeader(remainingHeader)
	if err != nil {
		return nil, err
	}

	return &RateLimits{
		Limit:           &limit,
		LimitWindow:     &limitWindow,
		Remaining:       &remaining,
		RemainingWindow: &remainingWindow,
		Source:          &source,
	}, nil
}

func tryGetToken(c *Client) (string, error) {
	token, err := c.getToken("", true)
	if err != nil {
		token, err = c.getToken(c.password, false)
		if err != nil {
			token, err = c.getToken(c.refreshToken, false)
			if err != nil {
				token, err = c.getToken(c.token, false)
				if err != nil {
					return "", err
				}
			}
		}
	}
	return token, nil
}

func (c *Client) getToken(password string, anonymous bool) (string, error) {
	req, err := http.NewRequest("GET", first, nil)
	if err != nil {
		return "", err
	}

	if !anonymous {
		req.Header.Add("Authorization", "Basic "+basicAuth(c.account, password))
	}
	resp, err := c.doRawRequest(req)
	if err != nil {
		return "", err
	}
	if resp.Body != nil {
		defer resp.Body.Close() //nolint:errcheck
	}

	if resp.StatusCode != http.StatusOK {
		return "", errors.New("unable to get authorization token")
	}

	buf, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	var t tokenResponse
	if err := json.Unmarshal(buf, &t); err != nil {
		return "", err
	}

	return t.Token, nil
}

func parseLimitHeader(value string) (int, int, error) {
	parts := strings.Split(value, ";")
	if len(parts) != 2 {
		return 0, 0, fmt.Errorf("bad limit header %s", value)
	}

	v, err := strconv.Atoi(parts[0])
	if err != nil {
		return 0, 0, err
	}

	windowParts := strings.Split(parts[1], "=")
	if len(windowParts) != 2 {
		return 0, 0, fmt.Errorf("bad limit header %s", value)
	}
	w, err := strconv.Atoi(windowParts[1])
	if err != nil {
		return 0, 0, err
	}

	return v, w, nil
}

func basicAuth(username, password string) string {
	auth := username + ":" + password
	return base64.StdEncoding.EncodeToString([]byte(auth))
}
