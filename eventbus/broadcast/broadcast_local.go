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

package broadcast

import (
	"context"

	"github.com/observiq/bindplane-op/eventbus"
	"go.uber.org/zap"
)

type localBroadcast[T any] struct {
	producerBuffer eventbus.Source[T]
	consumer       eventbus.Source[T]
	opts           Options[T]
	logger         *zap.Logger
}

var _ Broadcast[any] = (*localBroadcast[any])(nil)
var _ eventbus.Receiver[any] = (*localBroadcast[any])(nil)

// NewLocalBroadcast returns a new broadcast interface. This broadcast interface is used when pub/sub is not enabled. It
// does not require a message type because it does not use pub/sub.
func NewLocalBroadcast[T any](ctx context.Context, logger *zap.Logger, options ...Option[T]) Broadcast[T] {
	opts := MakeBroadcastOptions(options)

	b := &localBroadcast[T]{
		logger:         logger,
		opts:           opts,
		consumer:       InitConsumer(opts),
		producerBuffer: eventbus.NewSource[T](),
	}

	// producer => consumer
	RelayProducer[T, T](ctx, b.producerBuffer, b.consumerFilter(ctx), b.consumer, opts)

	return b
}

// Producer returns the producer which can be used to send messages to pub/sub.
func (b *localBroadcast[T]) Producer() eventbus.Receiver[T] {
	return b
}

// Consumer returns the source which can be subscribed to to receive messages from other nodes in the cluster.
func (b *localBroadcast[T]) Consumer() eventbus.Source[T] {
	return b.consumer
}

// Send sends a message to pub/sub to be received by all nodes in the cluster. It implements Receiver[*T] for use with
// RelayWithMerge.
func (b *localBroadcast[T]) Send(ctx context.Context, msg T) {
	// for local broadcasts, we can just send immediately, but we still need to use filtering if configured
	b.producerBuffer.Send(ctx, msg)
}

// ----------------------------------------------------------------------
// filter for processing messages

func (b *localBroadcast[T]) consumerFilter(ctx context.Context) eventbus.SubscriptionFilter[T, T] {
	return func(msg T) (T, bool) {
		// we can ignore the messageType because local broadcasts are sent directly
		//
		// we can ignore the ordering key because local broadcasts are sent immediately

		// build message attributes for filtering
		attrs := MessageAttributes{}
		b.opts.AddAttributes(ctx, msg, attrs)

		// if routing is configured, only process messages that have a route
		if !b.opts.HasRoute(attrs) {
			return msg, false
		}

		// if attribute filtering is configured, only process messages that pass the filter
		if !b.opts.AcceptMessage(attrs) {
			return msg, false
		}

		return msg, true
	}
}
