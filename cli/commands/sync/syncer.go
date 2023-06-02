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

package sync

import (
	"context"

	"github.com/observiq/bindplane-op/client"
	"github.com/observiq/bindplane-op/model"
)

// Syncer is an interface for syncing resources.
type Syncer interface {
	// SyncAgentVersions syncs agent versions.
	SyncAgentVersions(ctx context.Context, version string) ([]*model.AnyResourceStatus, error)
}

// Builder is an interface for building a Syncer.
type Builder interface {
	// Build returns a new Syncer.
	BuildSyncer(ctx context.Context) (Syncer, error)
}

// NewSyncer returns a new Syncer.
func NewSyncer(client client.BindPlane) Syncer {
	return &defaultSyncer{
		client: client,
	}
}

// defaultSyncer is the default implementation of the Syncer interface.
type defaultSyncer struct {
	client client.BindPlane
}

// SyncAgentVersions syncs agent versions.
func (s *defaultSyncer) SyncAgentVersions(ctx context.Context, version string) ([]*model.AnyResourceStatus, error) {
	return s.client.SyncAgentVersions(ctx, version)
}
