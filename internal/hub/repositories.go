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

import "time"

//Repository represents a Docker Hub repository
type Repository struct {
	Name        string
	Description string
	LastUpdated time.Time
	PullCount   int
	StarCount   int
	IsPrivate   bool
}

type hubRepositoryResponse struct {
	Count    int                   `json:"count"`
	Next     string                `json:"next,omitempty"`
	Previous string                `json:"previous,omitempty"`
	Results  []hubRepositoryResult `json:"results,omitempty"`
}

type hubRepositoryResult struct {
	Name           string         `json:"name"`
	Namespace      string         `json:"namespace"`
	PullCount      int            `json:"pull_count"`
	StarCount      int            `json:"star_count"`
	RepositoryType RepositoryType `json:"repository_type"`
	CanEdit        bool           `json:"can_edit"`
	Description    string         `json:"description,omitempty"`
	IsAutomated    bool           `json:"is_automated"`
	IsMigrated     bool           `json:"is_migrated"`
	IsPrivate      bool           `json:"is_private"`
	LastUpdated    time.Time      `json:"last_updated"`
	Status         int            `json:"status"`
	User           string         `json:"user"`
}

//RepositoryType lists all the different repository types handled by the Docker Hub
type RepositoryType string

const (
	//ImageType is the classic image type
	ImageType = RepositoryType("image")
)
