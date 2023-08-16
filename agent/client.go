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

// Package agent provides a client for fetching information about agent versions from Github.
package agent

import (
	"errors"

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
	// Version returns the agent version for the given version string
	Version(version string) (*model.AgentVersion, error)
	// Versions returns all available agent versions
	Versions() ([]*model.AgentVersion, error)
	// LatestVersion returns the latest agent version
	LatestVersion() (*model.AgentVersion, error)
}

// ----------------------------------------------------------------------

// NewGitHubVersionClient creates a new VersionClient for github versions
func NewGitHubVersionClient() VersionClient {
	return newGithub()
}

// NewNoopClient creates a new client that does not connect to other services
func NewNoopClient() VersionClient {
	return newNoopClient()
}
