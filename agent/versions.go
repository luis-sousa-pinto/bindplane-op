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
	"context"
	"fmt"
	"time"

	"github.com/observiq/bindplane-op/eventbus"
	"github.com/observiq/bindplane-op/model"
	"github.com/observiq/bindplane-op/store"
	"github.com/observiq/bindplane-op/util"
	"go.uber.org/zap"
)

const (
	// VersionLatest can be used in requests instead of an actual version
	VersionLatest = "latest"
)

// Versions manages versions of agents that are used during install and upgrade. The versions are stored in the Store as
// agent-version resources, but Versions provides quick access to the latest version.
//
//go:generate mockery --name Versions --filename mock_versions.go --structname MockVersions
type Versions interface {
	// LatestVersionString returns the semver version string of the latest agent version
	LatestVersionString(ctx context.Context) string
	// LatestVersion returns the latest agent version
	LatestVersion(ctx context.Context) (*model.AgentVersion, error)
	// Version returns the agent version for the given semver
	Version(ctx context.Context, version string) (*model.AgentVersion, error)
	SyncVersion(version string) (*model.AgentVersion, error)
	SyncVersions() ([]*model.AgentVersion, error)
}

// VersionsSettings is configuration for a Versions implementation
type VersionsSettings struct {
	Logger *zap.Logger

	// SyncAgentVersionsInterval is the interval at which SyncVersions() will be called to ensure the agent-versions are
	// in sync with GitHub and new releases are available.
	SyncAgentVersionsInterval time.Duration

	// Offline is true if the server is in offline mode and should not contact GitHub automatically. Sync methods called
	// by 'bindplane sync' commands will still attempt to contact GitHub.
	Offline bool
}

// The latest version cache keeps the latest version in memory to avoid hitting the store to get the latest version.
const (
	latestVersionCacheDuration = 15 * time.Minute

	// MinSyncAgentVersionsInterval is the minimum value for the SyncAgentVersionsInterval setting, currently 1 hour. 0
	// can also be specified to disable background periodic syncing.
	MinSyncAgentVersionsInterval = 1 * time.Hour
)

type versions struct {
	client        VersionClient
	store         store.Store
	latestVersion util.Remember[model.AgentVersion]
	logger        *zap.Logger
}

var _ Versions = (*versions)(nil)

// NewVersions creates an implementation of Versions using the specified client, cache, and settings. To disable
// caching, pass nil for the Cache.
func NewVersions(ctx context.Context, client VersionClient, store store.Store, settings VersionsSettings) Versions {
	v := &versions{
		client:        client,
		store:         store,
		latestVersion: util.NewRemember[model.AgentVersion](latestVersionCacheDuration),
		logger:        settings.Logger,
	}
	if settings.SyncAgentVersionsInterval > 0 {
		interval := settings.SyncAgentVersionsInterval
		if interval < MinSyncAgentVersionsInterval {
			interval = MinSyncAgentVersionsInterval
		}
		go v.syncAgentVersions(ctx, interval)
	}
	go v.watchAgentVersionUpdates(ctx)
	return v
}

func (v *versions) LatestVersionString(ctx context.Context) string {
	version, err := v.LatestVersion(ctx)
	if err != nil || version == nil {
		return ""
	}
	return version.AgentVersion()
}

// LatestVersion returns the latest *model.AgentVersion.
func (v *versions) LatestVersion(ctx context.Context) (*model.AgentVersion, error) {
	// check if we have a remembered result
	if remembered := v.latestVersion.Get(); remembered != nil {
		return remembered, nil
	}

	// find the latest public version
	agentVersions, err := v.store.AgentVersions(ctx)
	if err != nil {
		return nil, err
	}
	model.SortAgentVersionsLatestFirst(agentVersions)

	var found *model.AgentVersion
	for _, agentVersion := range agentVersions {
		if agentVersion.Public() {
			found = agentVersion
			break
		}
	}

	// cache it before returning
	if found != nil {
		v.latestVersion.Update(found)
	}

	return found, nil
}

// Version returns the specified agent version. If the version is invalid or does not exist, it returns an error. If
// version is "latest", it returns the latest version.
func (v *versions) Version(ctx context.Context, version string) (*model.AgentVersion, error) {
	if version == VersionLatest {
		return v.LatestVersion(ctx)
	}

	name := fmt.Sprintf("%s-%s", model.AgentTypeNameObservIQOtelCollector, version)

	found, err := v.store.AgentVersion(ctx, name)
	if err != nil {
		return nil, err
	}

	return found, nil
}

func (v *versions) SyncVersion(version string) (*model.AgentVersion, error) {
	return v.client.Version(version)
}

func (v *versions) SyncVersions() ([]*model.AgentVersion, error) {
	return v.client.Versions()
}

// ----------------------------------------------------------------------

func (v *versions) syncAgentVersions(ctx context.Context, interval time.Duration) {
	// sync once immediately
	v.syncAgentVersionsOnce(ctx)

	// sync at regular intervals
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			v.syncAgentVersionsOnce(ctx)
		}
	}
}

func (v *versions) syncAgentVersionsOnce(ctx context.Context) {
	agentVersions, err := v.SyncVersions()
	if err != nil {
		v.logger.Error("error during syncAgentVersions SyncVersions", zap.Error(err))
		return
	}

	// assemble the model.Resource array for Apply
	var resources []model.Resource
	for _, agentVersion := range agentVersions {
		resources = append(resources, agentVersion)
	}

	resourceStatuses, err := v.store.ApplyResources(ctx, resources)
	if err != nil {
		v.logger.Error("error during syncAgentVersions ApplyResources", zap.Error(err))
		return
	}

	var messages []string
	for _, resourceStatus := range resourceStatuses {
		messages = append(messages, resourceStatus.String())
	}
	v.logger.Debug("syncAgentVersions", zap.Strings("statuses", messages))
}

func (v *versions) watchAgentVersionUpdates(ctx context.Context) {
	channel, unsubscribe := eventbus.SubscribeWithFilter(ctx, v.store.Updates(ctx), func(u store.BasicEventUpdates) (store.BasicEventUpdates, bool) {
		return u, len(u.AgentVersions()) > 0
	})
	defer unsubscribe()

	for {
		select {
		case <-ctx.Done():
			return
		case <-channel:
			// clear the latest version whenever we see any AgentVersion changes
			v.latestVersion.Forget()
		}
	}
}
