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
	"go.opentelemetry.io/otel"
	"go.uber.org/zap"
)

var tracer = otel.Tracer("broadcast")

type localBroadcast[T any] struct {
	producerBuffer eventbus.Source[T]
	consumer       eventbus.Source[T]
	logger         *zap.Logger
}

// NewLocalBroadcast returns a new local broadcast with no routing or filtering.
func NewLocalBroadcast[T any](ctx context.Context, logger *zap.Logger, options ...Option[T]) Broadcast[T] {
	opts := MakeBroadcastOptions(options)
	b := &localBroadcast[T]{
		logger:         logger,
		consumer:       eventbus.NewSource[T](),
		producerBuffer: eventbus.NewSource[T](),
	}

	// producer => consumer
	RelayProducer[T, T](ctx, b.producerBuffer, b.consumerFilter(ctx), b.consumer, opts)

	return b
}

// Producer returns the producer for this broadcast
func (b *localBroadcast[T]) Producer() eventbus.Receiver[T] {
	return b
}

// Consumer returns the consumer for this broadcast
func (b *localBroadcast[T]) Consumer() eventbus.Source[T] {
	return b.consumer
}

// Send sends a message to the broadcast's producer
func (b *localBroadcast[T]) Send(ctx context.Context, msg T) {
	ctx, span := tracer.Start(ctx, "localBroadcast/Send")
	defer span.End()
	b.producerBuffer.Send(ctx, msg)
}

func (b *localBroadcast[T]) consumerFilter(_ context.Context) eventbus.SubscriptionFilter[T, T] {
	return func(msg T) (T, bool) {
		// we can ignore the messageType because local broadcasts are sent directly
		// we can ignore the ordering `key because local broadcasts are sent immediately
		// we can ignore any routing or filtering because local broadcasts don't route or filter
		return msg, true
	}
}
