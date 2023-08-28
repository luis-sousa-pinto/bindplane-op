// Copyright observIQ, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//	http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package store

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/observiq/bindplane-op/eventbus"
	"github.com/observiq/bindplane-op/model"
	"github.com/observiq/bindplane-op/store/storetest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

// eventRecorder helps with verifying that Store update events are dispatched as expected. It records events in a buffer
// and uses require.Eventually to confirm that the events exist.
type eventRecorder struct {
	startTime time.Time
	events    []BasicEventUpdates
	cancel    context.CancelFunc
	mtx       sync.Mutex
}

func recordStoreUpdates(ctx context.Context, store Store) *eventRecorder {
	recorder := &eventRecorder{
		startTime: time.Now(),
	}
	// cancelling this context will auto-unsubscribe and stop recording events
	ctx, recorder.cancel = context.WithCancel(ctx)

	// subscribe and start recording
	updates := store.Updates(ctx)
	ch, _ := eventbus.Subscribe(ctx, updates)
	recorder.record(ctx, ch)

	return recorder
}

func (r *eventRecorder) record(ctx context.Context, c <-chan BasicEventUpdates) {
	go func() {
		for {
			select {
			case update := <-c:
				r.mtx.Lock()
				r.events = append(r.events, update)
				r.mtx.Unlock()
			case <-ctx.Done():
				return
			}
		}
	}()
}

// assertEvents follows the pattern of the assert library
func (r *eventRecorder) assertEvents(t *testing.T, expectedEvents []BasicEventUpdates, msgAndArgs ...interface{}) bool {
	t.Helper()
	defer r.cancel()
	// minRecordTime ensures that we wait long enough to receive events. we need to ensure that we don't receive any extra
	// events. part of the reason we need to wait is that merging within updates operates with a (currently fixed) 100ms
	// delay.
	minRecordTime := 200 * time.Millisecond
	elapsed := time.Now().Sub(r.startTime)
	if elapsed < minRecordTime {
		wait := time.Duration(minRecordTime - elapsed)
		time.Sleep(wait)
	}

	// merge all update for deterministic comparison
	expectedEvents = mergeAllUpdates(expectedEvents)

	// make sure we eventually get the right number of events
	result := assert.Eventually(t, func() bool {
		r.mtx.Lock()
		defer r.mtx.Unlock()

		r.events = mergeAllUpdates(r.events)

		if len(expectedEvents) != len(r.events) {
			return false
		}
		for i, expectedEvent := range expectedEvents {
			if !assert.ObjectsAreEqual(expectedEvent, r.events[i]) {
				return false
			}
		}
		return true
	}, time.Second, 10*time.Millisecond, msgAndArgs...)
	if !result {
		return assert.Equal(t, expectedEvents, r.events)
	}
	return result
}

// requireEvents follows the pattern of the require library and fails immediately
func (r *eventRecorder) requireEvents(t *testing.T, expectedEvents []BasicEventUpdates, msgAndArgs ...interface{}) {
	if r.assertEvents(t, expectedEvents, msgAndArgs...) {
		return
	}
	t.FailNow()
}

func TestDelayedEvent(t *testing.T) {
	db, err := storetest.InitTestBboltDB(t, testBuckets)
	require.NoError(t, err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	store := NewBoltStore(ctx, db, testOptions, zap.NewNop())
	defer store.Close()

	// start recording
	r := recordStoreUpdates(ctx, store)

	// delay an agent event
	go func() {
		time.Sleep(50 * time.Millisecond)
		_, err := store.UpsertAgent(ctx, "a", func(a *model.Agent) {
			a.Name = "testAgent"
		})
		require.NoError(t, err)
	}()

	// we expect this test to fail, so we need to create our own T and then ensure that it failed. it would pass if we
	// didn't wait for the event.
	metaT := &testing.T{}
	r.assertEvents(metaT, nil)
	require.True(t, metaT.Failed(), "the inner test should fail (Failed should be True) because it asserts that there are no events and there is an event after 50ms")
}
