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
	"errors"
	"sync"
	"time"

	"github.com/observiq/bindplane-op/model"
	"go.uber.org/zap"
)

// DefaultRolloutBatchFlushInterval the default interval at which a batch is flushed
const DefaultRolloutBatchFlushInterval = 100 * time.Millisecond

// RolloutBatcher batches rollout updates before saving them in storage
//
//go:generate mockery --name RolloutBatcher --filename mock_rollout_batcher.go --structname MockRolloutBatcher
type RolloutBatcher interface {
	// Batch adds the incoming events to the batch
	Batch(ctx context.Context, agentEvents Events[*model.Agent]) error

	// Shutdown stops the batcher
	Shutdown(ctx context.Context) error
}

// NopRolloutBatcher is a nop Rollout batcher
type NopRolloutBatcher struct{}

// NewNopRolloutBatcher creates a new NopRolloutBatcher
func NewNopRolloutBatcher() *NopRolloutBatcher {
	return &NopRolloutBatcher{}
}

// Batch does nothing returns nil
func (n *NopRolloutBatcher) Batch(_ context.Context, _ Events[*model.Agent]) error {
	return nil
}

// Shutdown does nothing returns nil
func (n *NopRolloutBatcher) Shutdown(_ context.Context) error {
	return nil
}

// RolloutEventBatch batches rollout updates by unique configs
type RolloutEventBatch map[string]struct{}

// BatchConfig batches the config name trimmed of version
func (r RolloutEventBatch) BatchConfig(configName string) {
	if trimmedName := model.TrimVersion(configName); trimmedName != "" {
		r[trimmedName] = struct{}{}
	}
}

// DefaultRolloutBatcher is the default implementation of the RolloutBatcher
type DefaultRolloutBatcher struct {
	s             Store
	eventChan     chan []model.ConfigurationVersions
	logger        *zap.Logger
	flushInterval time.Duration

	wg     sync.WaitGroup
	ctx    context.Context
	cancel context.CancelCauseFunc
}

// NewDefaultBatcher creates a new default rollout batcher
func NewDefaultBatcher(ctx context.Context, logger *zap.Logger, flushInterval time.Duration, s Store) *DefaultRolloutBatcher {
	batcherCtx, cancel := context.WithCancelCause(ctx)

	batcher := &DefaultRolloutBatcher{
		s:             s,
		eventChan:     make(chan []model.ConfigurationVersions, 100),
		logger:        logger.Named("rollout_batcher"),
		flushInterval: flushInterval,
		ctx:           batcherCtx,
		cancel:        cancel,
	}

	// Spin off workers
	batcher.wg.Add(1)
	go batcher.batchWorker()

	return batcher
}

// Batch accepts events to be batched to be batched
func (d *DefaultRolloutBatcher) Batch(ctx context.Context, agentEvents Events[*model.Agent]) error {
	configStatuses := make([]model.ConfigurationVersions, 0, len(agentEvents))

	for _, agentEvent := range agentEvents {
		configStatuses = append(configStatuses, agentEvent.Item.ConfigurationStatus)
	}

	select {
	case <-ctx.Done():
		d.logger.Error("Context error while batching events", zap.Error(ctx.Err()))
		return ctx.Err()
	case d.eventChan <- configStatuses:
		return nil
	}
}

// Shutdown shuts down the batcher
func (d *DefaultRolloutBatcher) Shutdown(ctx context.Context) error {
	doneChan := make(chan struct{})

	go func() {
		defer close(doneChan)
		d.cancel(errors.New("shutdown"))
		d.wg.Wait()
	}()

	select {
	case <-ctx.Done():
		d.logger.Error("Error while shutting down", zap.Error(ctx.Err()))
		return ctx.Err()
	case <-doneChan:
		return nil
	}
}

func (d *DefaultRolloutBatcher) batchWorker() {
	defer d.wg.Done()

	ticker := time.NewTicker(d.flushInterval)
	defer ticker.Stop()

	batch := make(RolloutEventBatch)
	for {
		select {
		case <-d.ctx.Done():
			return
		case configStatuses := <-d.eventChan:
			for _, configStatus := range configStatuses {
				batch.BatchConfig(configStatus.Current)
				batch.BatchConfig(configStatus.Pending)
				batch.BatchConfig(configStatus.Future)
			}
		case <-ticker.C:
			// If there's nothing in the batch then continue on
			if len(batch) == 0 {
				continue
			}

			for configName := range batch {
				if _, err := d.s.UpdateRollout(d.ctx, configName); err != nil {
					d.logger.Error("Failed to update rollout",
						zap.Error(err),
						zap.String("config", configName),
					)
				}
				// Remove after processing to clean up map
				delete(batch, configName)
			}
		}
	}

}
