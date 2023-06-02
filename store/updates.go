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
	"github.com/observiq/bindplane-op/model"
	"go.uber.org/zap"
)

type key int

var updatesContextKey key

// UpdatesForContext returns the Updates for this Context. If there is no Updates on the Context, it creates a new
// Updates and adds it to the Context. It returns either the existing Context or a child Context with new Updates as
// appropriate. It returns true if the Updates were already found on the context.
//
// Keeping the Updates on the context allows for recursive calls to ApplyResources without creating multiple Updates
// which would result in multiple Updates events. Store implementations should use this function instead of NewUpdates
// and should only notify if shouldNotify is true.
func UpdatesForContext(ctx context.Context) (updates BasicEventUpdates, newContext context.Context, shouldNotify bool) {
	updates, ok := ctx.Value(updatesContextKey).(BasicEventUpdates)
	if !ok {
		updates = NewEventUpdates()
		ctx = context.WithValue(ctx, updatesContextKey, updates)
	}
	return updates, ctx, !ok
}

// Updates is a wrapped event bus for store updates.
type Updates struct {
	broadcast            broadcast.Broadcast[BasicEventUpdates]
	rolloutBroadcast     broadcast.Broadcast[RolloutEventUpdates]
	rolloutUpdateCreator RolloutUpdateCreator
	cancel               context.CancelFunc
}

// NewUpdates creates a new UpdatesEventBus.
func NewUpdates(ctx context.Context, options Options, logger *zap.Logger,
	basicEventBroadcaster BroadCastBuilder[BasicEventUpdates],
	rolloutBroadcaster BroadCastBuilder[RolloutEventUpdates],
	rolloutUpdateCreator RolloutUpdateCreator,
) *Updates {
	ctx, cancel := context.WithCancel(ctx)
	maxEventsToMerge := options.MaxEventsToMerge
	if maxEventsToMerge == 0 {
		maxEventsToMerge = 100
	}

	return &Updates{
		broadcast:            basicEventBroadcaster(ctx, options, logger, maxEventsToMerge),
		rolloutBroadcast:     rolloutBroadcaster(ctx, options, logger, maxEventsToMerge),
		rolloutUpdateCreator: rolloutUpdateCreator,
		cancel:               cancel,
	}
}

// Updates returns the external channel that can be provided to external clients.
func (s *Updates) Updates() eventbus.Source[BasicEventUpdates] {
	return s.broadcast.Consumer()
}

// RolloutUpdates returns the external channel that can be provided to external clients
func (s *Updates) RolloutUpdates() eventbus.Source[RolloutEventUpdates] {
	return s.rolloutBroadcast.Consumer()
}

// Send adds an Updates event to the internal channel where it can be merged and relayed to the external channel.
func (s *Updates) Send(ctx context.Context, updates BasicEventUpdates) {
	s.broadcast.Producer().Send(ctx, updates)

	if !updates.Agents().Empty() {
		rolloutEvent := s.rolloutUpdateCreator(ctx, updates.Agents())
		s.rolloutBroadcast.Producer().Send(ctx, rolloutEvent)
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

// RolloutUpdateCreator is a function that creates a RolloutEventUpdates from a set of agent events.
type RolloutUpdateCreator func(context.Context, Events[*model.Agent]) RolloutEventUpdates
