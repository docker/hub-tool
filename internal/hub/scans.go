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
	"strconv"
	"time"
)

const (
	// ScansURL path to the Hub API listing the scans
	// POST /reports/{scanner}/{registry}/{account}/{repo}
	ScansURL = "/api/scan/v1/reports/snyk/docker.io/%s/"
	// ScanURL path to the Hub API describing a scan
	// GET /reports/{scanner}/{registry}/{account}/{repo}/{digest}/{scanned_at}
	ScanURL = "/api/scan/v1/reports/snyk/docker.io/%s/%s/%s/newest"
)

//ScanReportSummary  summaries how many vulnerabilities were found
type ScanReportSummary struct {
	High      int
	Medium    int
	Low       int
	Total     int
	ScannedAt time.Time
}

//GetScanSummaries calls the hub repo API and returns all the scan summaries
func (c *Client) GetScanSummaries(repository string, digests ...string) (map[string]ScanReportSummary, error) {
	digests = deduplicateDigests(digests)
	repoPath, err := getRepoPath(repository)
	if err != nil {
		return nil, err
	}
	u, err := url.Parse(c.domain + fmt.Sprintf(ScansURL, repoPath))
	if err != nil {
		return nil, err
	}
	input := hubScanReportSummaryBulkInput{Digests: digests}
	data, err := json.Marshal(input)
	if err != nil {
		return nil, err
	}
	body := bytes.NewBuffer(data)

	req, err := http.NewRequest("POST", u.String(), body)
	if err != nil {
		return nil, err
	}
	response, err := c.doRequest(req, withHubToken(c.token))
	if err != nil {
		return nil, err
	}
	var hubResponse hubScanReportSummaryBulkOutput
	if err := json.Unmarshal(response, &hubResponse); err != nil {
		return nil, err
	}
	return toScanSummary(hubResponse), nil
}

type hubScanReportSummaryBulkInput struct {
	Digests []string `json:"digests"`
}

type hubScanReportSummaryBulkOutput struct {
	Reports map[string]hubScanReportSummary `json:"Reports"`
}

type hubScanReportSummary struct {
	Registry        string              `json:"registry"`
	Account         string              `json:"account"`
	Repo            string              `json:"repo"`
	Digest          string              `json:"digest"`
	ScannedAt       string              `json:"scannedAt"`
	Scanner         string              `json:"scanner"`
	Vulnerabilities *hubScanVulnSummary `json:"vulnerabilities,omitempty"`
	Error           *string             `json:"error,omitempty"`
}

type hubScanVulnSummary struct {
	High   int `json:"high"`
	Medium int `json:"medium"`
	Low    int `json:"low"`
	Total  int `json:"total"`
}

func toScanSummary(reports hubScanReportSummaryBulkOutput) map[string]ScanReportSummary {
	summaries := map[string]ScanReportSummary{}
	for sha, summary := range reports.Reports {
		if summary.Error == nil && summary.Vulnerabilities != nil {
			scannedAt, err := strconv.Atoi(summary.ScannedAt)
			if err != nil {
				continue
			}
			summaries[sha] = ScanReportSummary{
				High:      summary.Vulnerabilities.High,
				Medium:    summary.Vulnerabilities.Medium,
				Low:       summary.Vulnerabilities.Low,
				Total:     summary.Vulnerabilities.Total,
				ScannedAt: time.Unix(int64(scannedAt), 0),
			}
		}
	}
	return summaries
}

func deduplicateDigests(digests []string) []string {
	dedup := map[string]bool{}
	for _, d := range digests {
		dedup[d] = true
	}
	var result []string
	for digest := range dedup {
		result = append(result, digest)
	}
	return result
}
