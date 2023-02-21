// Copyright  observIQ, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package agent

import (
	"errors"
	"io"

	"github.com/go-resty/resty/v2"
	"github.com/observiq/bindplane-op/model"
)

var (
	// ErrVersionNotFound is returned when the agent versions service returns a 404 for a version
	ErrVersionNotFound = errors.New("agent version not found")
)

// VersionClient is an interface for retrieving available agent versions
//
//go:generate mockery --name VersionClient --filename mock_version_client.go --structname MockVersionClient
type VersionClient interface {
	Version(version string) (*model.AgentVersion, error)
	Versions() ([]*model.AgentVersion, error)
	LatestVersion() (*model.AgentVersion, error)
}

// ----------------------------------------------------------------------

// NewVersionClient creates a new VersionClient
func NewVersionClient() VersionClient {
	return newGithub()
}

func reader(client *resty.Client, url string) (io.ReadCloser, error) {
	response, err := client.R().SetDoNotParseResponse(true).Get(url)
	if err != nil {
		return nil, err
	}
	return response.RawBody(), nil
}
