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

package store

import (
	"context"
	"time"

	jsoniter "github.com/json-iterator/go"
	"github.com/observiq/bindplane-op/eventbus/broadcast"
	"github.com/observiq/bindplane-op/model"
	"go.uber.org/zap"
)

// RolloutEventUpdates is a collection of rollout related events
type RolloutEventUpdates interface {
	// Updates retrieves the agent updates
	Updates() Events[*model.ConfigurationVersions]

	// Empty returns true if the updates are empty.
	Empty() bool
	// Merge merges another set of updates into this one, returns true
	// if it was able to merge any updates.
	Merge(other RolloutEventUpdates) bool
}

// RolloutUpdates is a basic implementation of the RolloutEventUpdates interface
type RolloutUpdates struct {
	UpdatesField Events[*model.ConfigurationVersions] `json:"updates"`
}

// NewRolloutUpdates creates a new RolloutUpdates
func NewRolloutUpdates(_ context.Context, agentEvents Events[*model.Agent]) RolloutEventUpdates {
	events := NewEvents[*model.ConfigurationVersions]()

	for _, agentEvent := range agentEvents {
		configStatus := agentEvent.Item.ConfigurationStatus
		events.Include(&configStatus, EventTypeRollout)
	}

	return &RolloutUpdates{
		UpdatesField: events,
	}
}

// Updates retrieves the agent updates
func (r *RolloutUpdates) Updates() Events[*model.ConfigurationVersions] {
	return r.UpdatesField
}

// Empty returns true if the updates are empty.
func (r *RolloutUpdates) Empty() bool {
	return r.UpdatesField.Empty()
}

// Merge merges another set of updates into this one, returns true
// if it was able to merge any updates.
func (r *RolloutUpdates) Merge(other RolloutEventUpdates) bool {
	if !r.UpdatesField.CanSafelyMerge(other.Updates()) {
		return false
	}

	r.UpdatesField.Merge(other.Updates())
	return true
}

// BuildRolloutEventBroadcast returns a BroadCastBuilder that builds a broadcast.Broadcast[RolloutEventUpdates] using routing and broadcast options for oss.
func BuildRolloutEventBroadcast() BroadCastBuilder[RolloutEventUpdates] {
	return func(ctx context.Context, options Options, logger *zap.Logger, maxEventsToMerge int) broadcast.Broadcast[RolloutEventUpdates] {
		return broadcast.NewLocalBroadcast(ctx, logger,
			broadcast.WithUnboundedChannel[RolloutEventUpdates](100*time.Millisecond),
			broadcast.WithParseFunc(func(data []byte) (RolloutEventUpdates, error) {
				var updates RolloutUpdates
				err := jsoniter.Unmarshal(data, &updates)
				return &updates, err
			}),
			broadcast.WithMerge(func(into, single RolloutEventUpdates) bool {
				return into.Merge(single)
			}, 100*time.Millisecond, maxEventsToMerge),
		)
	}
}
