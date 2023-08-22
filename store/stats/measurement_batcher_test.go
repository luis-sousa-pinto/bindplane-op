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

package stats

import (
	"context"
	"testing"
	"time"

	"github.com/observiq/bindplane-op/otlp/record"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestNewDefaultBatcher(t *testing.T) {
	measurements := newMockMeasurements(t)

	// create a context here so we can verify that on cancel the batcher cleans up goroutines.
	// The tests that the batcher derives it's lifecycle context from the passed in one.
	ctx, cancel := context.WithCancel(context.Background())

	batcher := NewDefaultBatcher(ctx, zap.NewNop(), measurements)

	require.Equal(t, measurements, batcher.measurements)

	cancel()

	require.Eventually(t, func() bool {
		batcher.wg.Wait()
		return true
	}, 100*time.Millisecond, 10*time.Millisecond)
}

func TestDefaultBatcher_AcceptMetrics(t *testing.T) {
	measurements := newMockMeasurements(t)
	MeasurementBatchFlushInterval = 2 * time.Second

	payloadSize := 10
	// Create a slice of empty Metrics since we don't actually care about the content of the metrics just the number of them
	payload := make([]*record.Metric, payloadSize)

	doneChan := make(chan struct{})

	measurements.On("SaveAgentMetrics", mock.Anything, mock.Anything).Return(nil).Run(func(args mock.Arguments) {
		metrics, ok := args.Get(1).([]*record.Metric)
		require.True(t, ok)
		require.Len(t, metrics, 3*payloadSize)
		close(doneChan)
	})

	batcher := NewDefaultBatcher(context.Background(), zap.NewNop(), measurements)

	// Send three payloads of metrics
	for i := 0; i < 3; i++ {
		err := batcher.AcceptMetrics(context.Background(), payload)
		require.NoError(t, err)
	}

	require.Eventually(t, func() bool {
		<-doneChan
		return true
	}, MeasurementBatchFlushInterval+time.Second, 10*time.Millisecond)

	err := batcher.Shutdown(context.Background())
	require.NoError(t, err)
}

func TestDefaultBatcher_AcceptMetrics_Error(t *testing.T) {
	measurements := newMockMeasurements(t)

	batcher := NewDefaultBatcher(context.Background(), zap.NewNop(), measurements)

	// Shutdown to ensure nothing will read off the channel
	err := batcher.Shutdown(context.Background())
	require.NoError(t, err)

	// Ensure all workers are shut down
	require.Eventually(t, func() bool {
		batcher.wg.Wait()
		return true
	}, 100*time.Millisecond, 10*time.Millisecond)

	// Recreate metric chan as unbuffered so we can ensure the context cancel happens.
	batcher.metricChan = make(chan []*record.Metric)

	// cancel the context before passing it in
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	// call with canceled context and no goroutines reading off channel.
	err = batcher.AcceptMetrics(ctx, []*record.Metric{})
	require.ErrorIs(t, err, context.Canceled)
}

func TestDefaultBatcher_Shutdown_Error(t *testing.T) {
	measurements := newMockMeasurements(t)

	batcher := NewDefaultBatcher(context.Background(), zap.NewNop(), measurements)

	// cancel the context before passing it in
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	// Add one more to the waitgroup to ensure we don't shut down before we check the context
	batcher.wg.Add(1)

	err := batcher.Shutdown(ctx)
	require.ErrorIs(t, err, context.Canceled)

	// Ensure this tests doesn't leave any hanging goroutines
	batcher.wg.Done()
	require.Eventually(t, func() bool {
		batcher.wg.Wait()
		return true
	}, 100*time.Millisecond, 10*time.Millisecond)
}
