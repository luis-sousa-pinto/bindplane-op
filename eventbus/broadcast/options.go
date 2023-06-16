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
	"errors"
	"time"

	"github.com/observiq/bindplane-op/eventbus"
)

type mergeConfig[T any] struct {
	merge            eventbus.SubscriptionMerger[T]
	maxLatency       time.Duration
	maxEventsToMerge int
}

type unboundedConfig[T any] struct {
	interval time.Duration
}

// Options are options for a broadcast. They are used to configure the broadcast's producer and consumer.
type Options[T any] struct {
	// routing
	routingKey      eventbus.RoutingKey[T, string]
	subscriptionKey eventbus.SubscriptionKey[string]

	// routingSource will be nil if there is no routing configured
	routingSource eventbus.RoutingSource[string, T]

	orderingKey     func(T) string
	addAttributes   func(T, MessageAttributes)
	attributeFilter func(attrs MessageAttributes) bool
	mergeConfig     *mergeConfig[T]
	parseFunc       func([]byte) (T, error)
	unboundedConfig *unboundedConfig[T]
}

// ParseTo parses bytes to type T
// Returns and error if no parseFunc is set
func (b *Options[T]) ParseTo(msg []byte) (T, error) {
	if b.parseFunc == nil {
		var zero T
		return zero, errors.New("no parse func specified")
	}
	return b.parseFunc(msg)
}

// RoutingKey returns the routing key for the message. Returns an empty string if the routingKey is nil.
func (b *Options[T]) RoutingKey(_ context.Context, msg T) string {
	if b.routingKey != nil {
		return b.routingKey(msg)
	}
	return ""
}

// HasRoute returns true if the message should be routed. Returns true if the routingSource is nil.
func (b *Options[T]) HasRoute(attrs MessageAttributes) bool {
	return b.routingSource == nil || b.routingSource.HasRoute(attrs[AttributeRoutingKey])
}

// AddAttributes adds attributes to the message. Does nothing if the addAttributes is nil.
func (b *Options[T]) AddAttributes(_ context.Context, msg T, attrs MessageAttributes) {
	if b.addAttributes != nil {
		b.addAttributes(msg, attrs)
	}
	// also add the routing key as an attribute
	if b.routingKey != nil {
		attrs[AttributeRoutingKey] = b.routingKey(msg)
	}
}

// OrderingKey returns the ordering key for the message. Returns an empty string if the orderingKey is nil.
func (b *Options[T]) OrderingKey(msg T) string {
	if b.orderingKey != nil {
		return b.orderingKey(msg)
	}
	if b.routingKey != nil {
		return b.routingKey(msg)
	}
	return ""
}

// AcceptMessage returns true if the message should be processed. Returns true if the attributeFilter is nil.
func (b *Options[T]) AcceptMessage(attrs MessageAttributes) bool {
	if b.attributeFilter != nil {
		return b.attributeFilter(attrs)
	}
	return true
}

// MakeBroadcastOptions creates a broadcastOptions from the given options.
func MakeBroadcastOptions[T any](options []Option[T]) Options[T] {
	opts := Options[T]{}
	for _, option := range options {
		option(&opts)
	}
	return opts
}

// InitConsumer initializes the consumer and returns the routing source if routing is configured.
func InitConsumer[T any](opts Options[T]) eventbus.Source[T] {
	if opts.routingSource != nil {
		return opts.routingSource
	}
	return eventbus.NewSource[T]()
}

// ProducerOpts returns the producer options for the broadcast.
func ProducerOpts[T, R any](opts Options[T]) []eventbus.SubscriptionOption[R] {
	var producerOpts []eventbus.SubscriptionOption[R]
	if uc := opts.unboundedConfig; uc != nil {
		producerOpts = append(producerOpts, eventbus.WithUnboundedChannel[R](uc.interval))
	}
	return producerOpts
}

// Option is used to configure a Broadcast.
type Option[T any] func(*Options[T])

// WithParseFunc returns a BroadcastOption that adds a parsing function
func WithParseFunc[T any](parseFunc func([]byte) (T, error)) Option[T] {
	return func(o *Options[T]) {
		o.parseFunc = parseFunc
	}
}

// WithRouting returns a BroadcastOption that configures the routing key and subscription key for the broadcast.
func WithRouting[T any](routingKey eventbus.RoutingKey[T, string], subscriptionKey eventbus.SubscriptionKey[string]) Option[T] {
	return func(o *Options[T]) {
		o.routingKey = routingKey
		o.subscriptionKey = subscriptionKey
		o.routingSource = eventbus.NewRoutingSource(routingKey, subscriptionKey)
	}
}

// WithOrderingKey specifies the ordering key to use for the message. It should be unique for each message type and must
// be less than 1KB with the messageType: prefix. If WithRouting is specified without WithOrderingKey, the
// RoutingKey of the message will be used as the ordering key.
func WithOrderingKey[T any](orderingKey func(T) string) Option[T] {
	return func(o *Options[T]) {
		o.orderingKey = orderingKey
	}
}

// WithAttributeProcessor applies a filter to messages before they are processed. It takes a function to add attributes
// to outgoing messages and a filter to process incoming messages. If the filter returns false, the message will be
// skipped.
func WithAttributeProcessor[T any](addAttributes func(m T, attrs MessageAttributes), attributeFilter func(attrs MessageAttributes) bool) Option[T] {
	return func(o *Options[T]) {
		o.addAttributes = addAttributes
		o.attributeFilter = attributeFilter
	}
}

// WithMerge configures the broadcast to merge events before sending them to the consumer. The merge function is called
// to merge events and the maxLatency and maxEventsToMerge are used to determine when to send the merged events.
func WithMerge[T any](merge eventbus.SubscriptionMerger[T], maxLatency time.Duration, maxEventsToMerge int) Option[T] {
	return func(o *Options[T]) {
		o.mergeConfig = &mergeConfig[T]{
			merge:            merge,
			maxLatency:       maxLatency,
			maxEventsToMerge: maxEventsToMerge,
		}
	}
}

// WithUnboundedChannel configures the broadcast to use an unbounded channel for sending messages to the consumer. This
// is useful for bursts of messages that need to be processed as quickly as possible.
func WithUnboundedChannel[T any](interval time.Duration) Option[T] {
	return func(o *Options[T]) {
		o.unboundedConfig = &unboundedConfig[T]{
			interval: interval,
		}
	}
}
