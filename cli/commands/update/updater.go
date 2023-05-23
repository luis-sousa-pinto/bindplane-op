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

package update

import (
	"context"

	"github.com/observiq/bindplane-op/client"
)

// Updater is an interface for updating BindPlane resources.
type Updater interface {
	// UpdateAgent updates the agent with the given id to the given version.
	UpdateAgent(ctx context.Context, id, version string) error
}

// Builder is an interface for building an Updater.
type Builder interface {
	// Build returns a new Updater.
	BuildUpdater(ctx context.Context) (Updater, error)
}

// NewUpdater returns a new Updater.
func NewUpdater(client client.BindPlane) Updater {
	return &defaultUpdater{
		client: client,
	}
}

// defaultUpdater is the default implementation of the Updater interface.
type defaultUpdater struct {
	client client.BindPlane
}

// UpdateAgent updates the agent with the given id to the given version.
func (u *defaultUpdater) UpdateAgent(ctx context.Context, id, version string) error {
	return u.client.AgentUpgrade(ctx, id, version)
}
