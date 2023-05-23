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

// Package broadcast contains the Broadcast interface and a simple local implementation.
package broadcast

import (
	"context"

	"github.com/observiq/bindplane-op/eventbus"
)

// Broadcast resembles both the noun and the verb of the word "broadcast". It is used to send a message to pub/sub to be
// received by all nodes in the cluster and it is used to receive messages from other nodes in the cluster.
type Broadcast[T any] interface {
	// Producer is used to send a message to pub/sub to be received by all nodes in the cluster.
	Producer() eventbus.Receiver[T]

	// Consumer returns the source can be subscribed to to receive messages from other nodes in the cluster.
	Consumer() eventbus.Source[T]
}

// RelayProducer relays messages from the source to the destination using the specified filter. It uses the
// broadcastOptions to determine if the source should be merged and if the channel should be unbounded.
func RelayProducer[T, R any](ctx context.Context, src eventbus.Source[T], filter eventbus.SubscriptionFilter[T, R], dst eventbus.Receiver[R], opts Options[T]) {
	if mc := opts.mergeConfig; mc != nil {
		merged := eventbus.NewSource[T]()
		// src => merged
		eventbus.RelayWithMerge[T](ctx, src, mc.merge, merged, mc.maxLatency, mc.maxEventsToMerge, ProducerOpts[T, T](opts)...)
		// merged => dst
		eventbus.RelayWithFilter(ctx, merged, filter, dst)
	} else {
		// src => dst
		eventbus.RelayWithFilter(ctx, src, filter, dst, ProducerOpts[T, R](opts)...)
	}
}
