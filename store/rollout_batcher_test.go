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
	"sync/atomic"
	"testing"
	"time"

	"github.com/observiq/bindplane-op/model"
	"github.com/oklog/ulid/v2"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestNopRolloutBatcher(t *testing.T) {
	batcher := NewNopRolloutBatcher()
	require.NoError(t, batcher.Batch(context.Background(), nil))
	require.NoError(t, batcher.Shutdown(context.Background()))
}

func TestNewDefaultRolloutBatcher(t *testing.T) {
	storeMock := newMockStore(t)

	// create a context here so we can verify that on cancel the batcher cleans up goroutines.
	// The tests that the batcher derives it's lifecycle context from the passed in one.
	ctx, cancel := context.WithCancel(context.Background())

	batcher := NewDefaultBatcher(ctx, zap.NewNop(), DefaultRolloutBatchFlushInterval, storeMock)

	require.Equal(t, storeMock, batcher.s)
	require.Equal(t, DefaultRolloutBatchFlushInterval, batcher.flushInterval)

	cancel()

	require.Eventually(t, func() bool {
		batcher.wg.Wait()
		return true
	}, 100*time.Millisecond, 10*time.Millisecond)
}

func TestNewDefaultRolloutBatcher_Batch(t *testing.T) {
	storeMock := newMockStore(t)

	var counter atomic.Int32
	storeMock.EXPECT().UpdateRollout(mock.Anything, "test").
		Return(nil, nil).
		Run(func(ctx context.Context, configurationName string) {
			counter.Add(1)
		})
	storeMock.EXPECT().UpdateRollout(mock.Anything, "foo").
		Return(nil, nil).
		Run(func(ctx context.Context, configurationName string) {
			counter.Add(1)
		})

	// Create update data
	updates := NewEventUpdates()
	a1 := &model.Agent{
		ID: ulid.Make().String(),
		ConfigurationStatus: model.ConfigurationVersions{
			Current: "test:1",
			Pending: "test:2",
			Future:  "test:2",
		},
	}
	a2 := &model.Agent{
		ID: ulid.Make().String(),
		ConfigurationStatus: model.ConfigurationVersions{
			Current: "test:2",
			Pending: "",
			Future:  "test:2",
		},
	}
	a3 := &model.Agent{
		ID: ulid.Make().String(),
		ConfigurationStatus: model.ConfigurationVersions{
			Current: "foo:1",
			Pending: "foo:2",
			Future:  "foo:2",
		},
	}
	a4 := &model.Agent{
		ID: ulid.Make().String(),
		ConfigurationStatus: model.ConfigurationVersions{
			Current: "foo:2",
			Pending: "",
			Future:  "foo:3",
		},
	}
	updates.IncludeAgent(a1, EventTypeUpdate)
	updates.IncludeAgent(a2, EventTypeUpdate)
	updates.IncludeAgent(a3, EventTypeUpdate)
	updates.IncludeAgent(a4, EventTypeUpdate)

	batcher := NewDefaultBatcher(context.Background(), zap.NewNop(), DefaultRolloutBatchFlushInterval, storeMock)
	err := batcher.Batch(context.Background(), updates.Agents())
	require.NoError(t, err)

	require.Eventually(t, func() bool {
		return counter.Load() == 2
	}, time.Second, 100*time.Millisecond)

	err = batcher.Shutdown(context.Background())
	require.NoError(t, err)
}

func TestNewDefaultRolloutBatcher_Batch_Error(t *testing.T) {
	storeMock := newMockStore(t)

	// Create update data
	updates := NewEventUpdates()
	a1 := &model.Agent{
		ID: ulid.Make().String(),
		ConfigurationStatus: model.ConfigurationVersions{
			Current: "test:1",
			Pending: "test:2",
			Future:  "test:2",
		},
	}
	updates.IncludeAgent(a1, EventTypeUpdate)

	batcher := NewDefaultBatcher(context.Background(), zap.NewNop(), DefaultRolloutBatchFlushInterval, storeMock)

	// Shutdown to ensure nothing will read off the channel
	err := batcher.Shutdown(context.Background())
	require.Eventually(t, func() bool {
		batcher.wg.Wait()
		return true
	}, 100*time.Millisecond, 10*time.Millisecond)

	// Recreate event chan as unbuffered so we can ensure the context cancel happens.
	batcher.eventChan = make(chan []model.ConfigurationVersions)

	// cancel the context before passing it in
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	err = batcher.Batch(ctx, updates.Agents())
	require.ErrorIs(t, err, context.Canceled)
}
