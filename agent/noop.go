// Copyright observIQ, Inc.
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
	"fmt"

	"github.com/observiq/bindplane-op/model"
)

type noopClient struct{}

func newNoopClient() noopClient {
	return noopClient{}
}

// Version always returns ErrVersionNotFound
func (noopClient) Version(_ string) (*model.AgentVersion, error) {
	return nil, fmt.Errorf("noop client: %w", ErrVersionNotFound)
}

// Versions returns an empty list of versions
func (noopClient) Versions() ([]*model.AgentVersion, error) {
	return []*model.AgentVersion{}, nil
}

// LatestVersion always returns ErrVersionNotFound
func (n noopClient) LatestVersion() (*model.AgentVersion, error) {
	return n.Version(VersionLatest)
}
