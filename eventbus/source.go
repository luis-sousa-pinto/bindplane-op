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

// Package eventbus provides a simple event bus implementation.
package eventbus

import (
	"context"
	"sync"
	"time"

	"github.com/observiq/bindplane-op/util"
	"go.opentelemetry.io/otel"
)

const subscriberChannelBufferSize = 10

var tracer = otel.Tracer("eventbus")

// UnsubscribeFunc is a function that allows a subscriber to unsubscribe
type UnsubscribeFunc func()

// Subscriber can be notified of events of type T. Instead of using this interface directly, use one of the eventbus.Subscribe
// functions to receive a channel of events.
//
//go:generate mockery --name Subscriber --filename mock_subscriber.go --structname MockSubscriber
type Subscriber[T any] interface {
	// Channel will be returned to Subscribe calls to receive events
	Channel() <-chan T

	// Receive will be called when an event is available
	Receive(event T)

	// Close will be called when the subscriber is unsubscribed
	Close()
}

// Receiver receives events of type T
//
//go:generate mockery --name Receiver --filename mock_receiver.go --structname MockReceiver
type Receiver[T any] interface {
	// Send the event to this receiver
	Send(ctx context.Context, event T)
}

// Source is a source of events.
//
//go:generate mockery --name Source --filename mock_source.go --structname MockSource
type Source[T any] interface {
	// Send the event to this receiver
	Send(ctx context.Context, event T)

	// Subscribe adds a subscriber to the source and automatically unsubscribes when the context is done. If the
	// context is nil, the unsubscribe function must be called to unsubscribe. Instead of using this method to subscribe
	// to a source, use one of the eventbus.Subscribe functions to receive a channel of events.
	Subscribe(ctx context.Context, subscriber Subscriber[T], onUnsubscribe func()) UnsubscribeFunc

	// Subscribers returns the current number of subscribers (used for testing)
	Subscribers() int
}

// SubscriptionFilter can filter on events and map from an event to another type. It can also ignore events. If accept
// is false, the result is ignored and not sent to subscribers.
type SubscriptionFilter[T, R any] func(event T) (result R, accept bool)

var exists = struct{}{}

// implementation of EventBus
type source[T any] struct {
	// subscribers is the set of current subscribers, implemented as a map to an empty struct
	subscribers map[Subscriber[T]]struct{}
	mtx         sync.RWMutex
}

var _ Source[any] = (*source[any])(nil)

// ----------------------------------------------------------------------
// generic subscriptions

type subscription[T any] struct {
	channel chan T
	ctx     context.Context
	cancel  context.CancelFunc
}

func newSubscription[T any](options []SubscriptionOption[T]) Subscriber[T] {
	opts := makeSubscriptionOptions(options)
	if opts.unbounded {
		return newUnboundedSubscription[T](opts.unboundedInterval)
	}
	channel := opts.channel
	if channel == nil {
		channel = make(chan T, subscriberChannelBufferSize)
	}

	ctx, cancel := context.WithCancel(context.Background())
	return &subscription[T]{
		channel: channel,
		ctx:     ctx,
		cancel:  cancel,
	}
}

func (s *subscription[T]) Receive(event T) {
	select {
	case <-s.ctx.Done():
		close(s.channel)
	case s.channel <- event:
	}
}

func (s *subscription[T]) Channel() <-chan T {
	return s.channel
}

func (s *subscription[T]) Close() {
	s.cancel()
}

var _ Subscriber[int] = (*subscription[int])(nil)

// ----------------------------------------------------------------------
// filter subscriptions

type filterSubscription[T, R any] struct {
	subscription Subscriber[R]
	filter       SubscriptionFilter[T, R]
}

func newFilterSubscription[T, R any](_ context.Context, filter SubscriptionFilter[T, R], options []SubscriptionOption[R]) *filterSubscription[T, R] {
	return &filterSubscription[T, R]{
		subscription: newSubscription(options),
		filter:       filter,
	}
}

func (s *filterSubscription[T, R]) Channel() <-chan T {
	panic("filterSubscription Channel() should not be called because it won't receive anything")
}
func (s *filterSubscription[T, R]) FilterChannel() <-chan R {
	return s.subscription.Channel()
}

func (s *filterSubscription[T, R]) Receive(event T) {
	filtered, accept := s.filter(event)
	if accept {
		s.subscription.Receive(filtered)
	}
}

func (s *filterSubscription[T, R]) Close() {
	s.subscription.Close()
}

var _ Subscriber[int] = (*filterSubscription[int, int])(nil)

// ----------------------------------------------------------------------
// unbounded

type unboundedSubscription[T any] struct {
	channel util.UnboundedChan[T]
}

const unboundedMinimumInterval = 50 * time.Millisecond

func newUnboundedSubscription[T any](interval time.Duration) *unboundedSubscription[T] {
	if interval < unboundedMinimumInterval {
		interval = unboundedMinimumInterval
	}
	return &unboundedSubscription[T]{
		channel: util.NewUnboundedChan[T](interval),
	}
}

func (s *unboundedSubscription[T]) Receive(event T) {
	s.channel.In() <- event
}

func (s *unboundedSubscription[T]) Channel() <-chan T {
	return s.channel.Out()
}

func (s *unboundedSubscription[T]) Close() {
	s.channel.Close()
}

var _ Subscriber[int] = (*unboundedSubscription[int])(nil)

// ----------------------------------------------------------------------
// package methods

// NewSource returns a new Source implementation for the specified event type T
func NewSource[T any]() Source[T] {
	return &source[T]{
		subscribers: make(map[Subscriber[T]]struct{}),
	}
}

// Subscribe subscribes to events on the bus and returns a channel to receive events and an unsubscribe
// function. It automatically unsubscribes when the context is done.
func Subscribe[T any](ctx context.Context, bus Source[T], options ...SubscriptionOption[T]) (<-chan T, UnsubscribeFunc) {
	subscription := newSubscription(options)
	opts := makeSubscriptionOptions(options)
	unsubscribe := bus.Subscribe(ctx, subscription, opts.unsubscribeHook)
	return subscription.Channel(), unsubscribe
}

// SubscribeWithFilter TODO
func SubscribeWithFilter[T, R any](ctx context.Context, source Source[T], filter SubscriptionFilter[T, R], options ...SubscriptionOption[R]) (<-chan R, UnsubscribeFunc) {
	subscription := newFilterSubscription(ctx, filter, options)
	opts := makeSubscriptionOptions(options)
	unsubscribe := source.Subscribe(ctx, subscription, opts.unsubscribeHook)
	return subscription.FilterChannel(), unsubscribe
}

// SubscribeUntilDone adds the subscriber and returns a cancel function.
func (s *source[T]) Subscribe(ctx context.Context, subscriber Subscriber[T], unsubscribeHook func()) UnsubscribeFunc {
	ctx, span := tracer.Start(ctx, "source/Subscribe")
	defer span.End()

	s.mtx.Lock()
	defer s.mtx.Unlock()

	// create a new subscription and add it
	s.subscribers[subscriber] = exists

	// avoid unsubscribe more than once
	unsubscribed := false

	// create the unsubscribe function
	unsubscribe := func() {
		s.mtx.Lock()
		// note: we manually unlock to ensure that the unsubscribeHook() is executed outside of the source being locked.
		// This allows routingSource to check Subscribers without causing a reentrant lock situation.
		if !unsubscribed {
			unsubscribed = true
			delete(s.subscribers, subscriber)
			s.mtx.Unlock()

			subscriber.Close()
			if unsubscribeHook != nil {
				// defer to avoid being locked if the source needs to be accessed
				unsubscribeHook()
			}
		} else {
			s.mtx.Unlock()
		}
	}

	// wait for context done and auto-unsubscribe
	if ctx != nil {
		go func() {
			<-ctx.Done()
			unsubscribe()
		}()
	}

	return unsubscribe
}

// Send the event to all of the subscribers
func (s *source[T]) Send(ctx context.Context, event T) {
	_, span := tracer.Start(ctx, "source/Send")
	defer span.End()

	for _, sub := range s.subscriberList() {
		sub.Receive(event)
	}
}

func (s *source[T]) subscriberList() []Subscriber[T] {
	s.mtx.RLock()
	defer s.mtx.RUnlock()
	subscribers := make([]Subscriber[T], 0, len(s.subscribers))
	for sub := range s.subscribers {
		subscribers = append(subscribers, sub)
	}
	return subscribers
}

// Subscribers returns the current number of subscribers
func (s *source[T]) Subscribers() int {
	s.mtx.RLock()
	defer s.mtx.RUnlock()
	return len(s.subscribers)
}

// Relay will relay from source to destination. It runs a separate goroutine, consuming events from the source and
// sending events to the destination. When the supplied context is Done, the relay is automatically unsubscribed from
// the source and the destination will no longer receive events.
func Relay[T any](ctx context.Context, source Source[T], destination Receiver[T], options ...SubscriptionOption[T]) {
	ctx, span := tracer.Start(ctx, "source/Relay")
	defer span.End()

	channel, unsubscribe := Subscribe(ctx, source, options...)
	go relay(ctx, channel, unsubscribe, destination)
}

// RelayWithFilter will relay from source to destination with the specified filter. It runs a separate goroutine,
// consuming events from the source, running the filter, and sending events to the destination. When the supplied
// context is Done, the relay is automatically unsubscribed from the source and the destination will no longer receive
// events.
func RelayWithFilter[T, R any](ctx context.Context, source Source[T], filter SubscriptionFilter[T, R], destination Receiver[R], options ...SubscriptionOption[R]) {
	ctx, span := tracer.Start(ctx, "source/RelayWithFilter")
	defer span.End()

	channel, unsubscribe := SubscribeWithFilter(ctx, source, filter, options...)
	go relay(ctx, channel, unsubscribe, destination)
}

func relay[T any](ctx context.Context, channel <-chan T, unsubscribe UnsubscribeFunc, destination Receiver[T]) {
	ctx, span := tracer.Start(ctx, "source/relay")
	defer span.End()

	defer unsubscribe()
	for {
		select {
		case event, ok := <-channel:
			if ok {
				destination.Send(ctx, event)
			} else {
				return
			}
		case <-ctx.Done():
			return
		}
	}
}

// SubscriptionMerger merges the second event into the first and returns true if the merge was successful. It should
// return false if a merge was not possible and the two individual events will be preserved and dispatched separately.
type SubscriptionMerger[T any] func(into, single T) bool

// RelayWithMerge will relay from source to destination, merging events before sending them to the destination. This can
// be used when there are lots of small individual events that can be more efficiently processed as a few larger events.
func RelayWithMerge[T any](ctx context.Context, source Source[T], merge SubscriptionMerger[T], destination Receiver[T], maxLatency time.Duration, maxEventsToMerge int, options ...SubscriptionOption[T]) {
	// constrain max events to at least 1
	if maxEventsToMerge < 1 {
		maxEventsToMerge = 1
	}
	channel, unsubscribe := Subscribe(ctx, source, options...)
	go func() {
		defer unsubscribe()

		maxLatencyTicker := time.NewTicker(maxLatency)
		defer maxLatencyTicker.Stop()

		var buffer []T
		mergeAndSend := func() {
			if len(buffer) == 0 {
				return
			}

			// merge: insert the first element merge into it
			prev := buffer[0]
			merged := []T{prev}

			for _, item := range buffer[1:] {
				if !merge(prev, item) {
					// unable to merge, append and start merging into this item
					merged = append(merged, item)
					prev = item
				}
			}

			// send the merged items
			for _, item := range merged {
				destination.Send(ctx, item)
			}

			// reset the buffer
			buffer = nil
		}

		// drain anything remaining when finished
		defer mergeAndSend()
		for {
			select {
			case event, ok := <-channel:
				if !ok {
					return
				}
				buffer = append(buffer, event)
				if len(buffer) >= maxEventsToMerge {
					mergeAndSend()
				}

			case <-maxLatencyTicker.C:
				// periodically drain the buffer to limit latency
				mergeAndSend()

			case <-ctx.Done():
				// send anything left in the buffer before stopping
				return
			}
		}
	}()
}

// ----------------------------------------------------------------------

// RoutingKey is a function that returns a route key for a specified event. Events will only be dispatched to
// subscribers that subscribe with a Context where SubscriptionKey returns this key.
type RoutingKey[T any, K comparable] func(event T) K

// SubscriptionKey is a function that returns a route key for a specified context. Subscribers will only receive events
// where RoutingKey returns this key.
type SubscriptionKey[K comparable] func(ctx context.Context) K

// RoutingSource is a special Source that can route events to subscribers based on the Context used to subscribe.
type RoutingSource[K, T any] interface {
	Source[T]

	// Subscribers returns the current number of routes
	Routes() int

	// HasRoute returns true if there is a route for the specified key
	HasRoute(K) bool
}

type routingSource[K comparable, T any] struct {
	routingKey      RoutingKey[T, K]
	subscriptionKey SubscriptionKey[K]
	routes          map[K]Source[T]
	mtx             sync.RWMutex
}

var _ RoutingSource[string, any] = (*routingSource[string, any])(nil)

// NewRoutingSource creates a new source that routes events using the specified RoutingKey and SubscriptionKey functions.
func NewRoutingSource[K comparable, T any](routingKey RoutingKey[T, K], subscriptionKey SubscriptionKey[K]) RoutingSource[K, T] {
	return &routingSource[K, T]{
		routingKey:      routingKey,
		subscriptionKey: subscriptionKey,
		routes:          map[K]Source[T]{},
	}
}

func (s *routingSource[K, T]) Send(ctx context.Context, event T) {
	s.mtx.RLock()
	defer s.mtx.RUnlock()

	key := s.routingKey(event)
	if route, ok := s.routes[key]; ok {
		route.Send(ctx, event)
	}
	// ignore events without a route
}

func (s *routingSource[K, T]) Subscribe(ctx context.Context, subscriber Subscriber[T], _ func()) UnsubscribeFunc {
	s.mtx.Lock()
	defer s.mtx.Unlock()
	key := s.subscriptionKey(ctx)
	route, ok := s.routes[key]
	if !ok {
		// create a new route
		route = NewSource[T]()
		s.routes[key] = route
	}
	return route.Subscribe(ctx, subscriber, func() {
		// cleanup this route if there are no more subscribers.
		s.mtx.Lock()
		defer s.mtx.Unlock()
		if route.Subscribers() == 0 {
			delete(s.routes, key)
		}
	})
}

// Subscribers returns the current number of routes
func (s *routingSource[K, T]) Routes() int {
	s.mtx.RLock()
	defer s.mtx.RUnlock()
	return len(s.routes)
}

func (s *routingSource[K, T]) HasRoute(key K) bool {
	s.mtx.RLock()
	defer s.mtx.RUnlock()
	_, ok := s.routes[key]
	return ok
}

// Subscribers returns the current number of subscribers
func (s *routingSource[K, T]) Subscribers() int {
	s.mtx.RLock()
	defer s.mtx.RUnlock()
	var count = 0
	for _, route := range s.routes {
		count += route.Subscribers()
	}
	return count
}
