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

package store

import (
	"context"

	"github.com/observiq/bindplane-op/eventbus"
	"github.com/observiq/bindplane-op/eventbus/broadcast"
	"go.uber.org/zap"
)

type key int

// UpdatesContextKey is the context key for updates
var UpdatesContextKey key

// UpdatesForContext returns the Updates for this Context. If there is no Updates on the Context, it creates a new
// Updates and adds it to the Context. It returns either the existing Context or a child Context with new Updates as
// appropriate. It returns true if the Updates were already found on the context.
//
// Keeping the Updates on the context allows for recursive calls to ApplyResources without creating multiple Updates
// which would result in multiple Updates events. Store implementations should use this function instead of NewUpdates
// and should only notify if shouldNotify is true.
func UpdatesForContext(ctx context.Context) (updates BasicEventUpdates, newContext context.Context, shouldNotify bool) {
	updates, ok := ctx.Value(UpdatesContextKey).(BasicEventUpdates)
	if !ok {
		updates = NewEventUpdates()
		ctx = context.WithValue(ctx, UpdatesContextKey, updates)
	}
	return updates, ctx, !ok
}

// Updates is a wrapped event bus for store updates.
type Updates struct {
	broadcast      broadcast.Broadcast[BasicEventUpdates]
	rolloutBatcher RolloutBatcher
	logger         *zap.Logger
	cancel         context.CancelFunc
}

// NewUpdates creates a new UpdatesEventBus.
func NewUpdates(ctx context.Context, options Options, logger *zap.Logger,
	rolloutBatcher RolloutBatcher,
	basicEventBroadcaster BroadCastBuilder[BasicEventUpdates],
) *Updates {
	ctx, cancel := context.WithCancel(ctx)
	maxEventsToMerge := options.MaxEventsToMerge
	if maxEventsToMerge == 0 {
		maxEventsToMerge = 100
	}

	return &Updates{
		broadcast:      basicEventBroadcaster(ctx, options, logger, maxEventsToMerge),
		rolloutBatcher: rolloutBatcher,
		logger:         logger.Named("updates"),
		cancel:         cancel,
	}
}

// Updates returns the external channel that can be provided to external clients.
func (s *Updates) Updates() eventbus.Source[BasicEventUpdates] {
	return s.broadcast.Consumer()
}

// Send adds an Updates event to the internal channel where it can be merged and relayed to the external channel.
func (s *Updates) Send(ctx context.Context, updates BasicEventUpdates) {
	s.broadcast.Producer().Send(ctx, updates)

	if !updates.Agents().Empty() {
		if err := s.rolloutBatcher.Batch(ctx, updates.Agents()); err != nil {
			s.logger.Error("Failed to batch rollout updates", zap.Error(err))
		}
	}
}

// Shutdown stops the event bus.
func (s *Updates) Shutdown(_ context.Context) {
	if s.cancel != nil {
		s.cancel()
	}
}

// BroadCastBuilder is a function that builds a broadcast.Broadcast[BasicUpdates] using routing and broadcast options.
type BroadCastBuilder[T any] func(ctx context.Context, options Options, logger *zap.Logger, maxEventsToMerge int) broadcast.Broadcast[T]

// mergeAllUpdates will merge all of the individual updates as much as possible. For example, if there are 4 updates,
// and the first 2 can be merged and the last 2 can be merged, there will be 2 updates in the result. Because the
// updates in the list will be merged into each other, the list should not be used again as it will contain duplicate
// update information. Because of this, the usage should typically look like:
//
//	updates = mergeAllUpdates(updates)
//
// This method is primarily used for testing to ensure that a list of updates is merged as much as possible and can be
// safely compared with a similar list of updates.
func mergeAllUpdates(list []BasicEventUpdates) []BasicEventUpdates {
	var result []BasicEventUpdates

	var prev BasicEventUpdates
	for _, cur := range list {
		// base case for the first iteration
		if prev != nil && MergeUpdates(prev, cur) {
			continue
		}
		// either first iteration (prev == nil) or we failed to merge and need to create a new base
		result = append(result, cur)
		prev = cur
	}

	return result
}
