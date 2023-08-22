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
	"errors"
	"sync"
	"time"

	"github.com/observiq/bindplane-op/otlp/record"
	"go.uber.org/zap"
)

// MeasurementBatchFlushInterval is the number of seconds to wait before flushing batch.
var MeasurementBatchFlushInterval = 5 * time.Second

// MeasurementBatcher the metric batcher accepts metrics and batches them before saving them in storage
//
//go:generate mockery --name MeasurementBatcher --filename mock_measurement_batcher.go --structname MockMeasurementBatcher
type MeasurementBatcher interface {
	// AcceptMetrics adds the metrics to the batcher that will eventually be saved in storage
	AcceptMetrics(ctx context.Context, metrics []*record.Metric) error

	// Shutdown stops the batcher
	Shutdown(ctx context.Context) error
}

// DefaultBatcher is the default implementation of the MeasurementBatcher
type DefaultBatcher struct {
	measurements Measurements
	metricChan   chan []*record.Metric
	batchChan    chan []*record.Metric
	logger       *zap.Logger

	wg     sync.WaitGroup
	ctx    context.Context
	cancel context.CancelCauseFunc
}

// NewDefaultBatcher creates a new default MeasurementBatcher
func NewDefaultBatcher(ctx context.Context, logger *zap.Logger, measurements Measurements) *DefaultBatcher {
	batcherCtx, cancel := context.WithCancelCause(ctx)

	batcher := &DefaultBatcher{
		measurements: measurements,
		metricChan:   make(chan []*record.Metric, 100),
		batchChan:    make(chan []*record.Metric, 10),
		logger:       logger.Named("measurement_batcher"),
		ctx:          batcherCtx,
		cancel:       cancel,
	}

	// Spin off workers
	batcher.wg.Add(2)
	go batcher.saveWorker()
	go batcher.acceptWorker()

	return batcher
}

// AcceptMetrics accepts metrics to be batched
func (d *DefaultBatcher) AcceptMetrics(ctx context.Context, metrics []*record.Metric) error {
	select {
	case <-ctx.Done():
		d.logger.Error("Context error while accepting metrics", zap.Error(ctx.Err()))
		return ctx.Err()
	case d.metricChan <- metrics:
		return nil
	}
}

// Shutdown shuts down the batcher
func (d *DefaultBatcher) Shutdown(ctx context.Context) error {
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

func (d *DefaultBatcher) acceptWorker() {
	defer d.wg.Done()

	ticker := time.NewTicker(MeasurementBatchFlushInterval)
	defer ticker.Stop()

	batch := make([]*record.Metric, 0)
	for {
		select {
		case <-d.ctx.Done():
			return
		case metrics := <-d.metricChan:
			batch = append(batch, metrics...)
		case <-ticker.C:
			if len(batch) > 0 {
				// Send to save worker
				d.batchChan <- batch

				// Reset buffer so the old one gets GCed
				batch = make([]*record.Metric, 0)
			}
		}
	}
}

// saveWorker handles actually saving measurements.
// This worker is to offload the saving from the accepting path
func (d *DefaultBatcher) saveWorker() {
	defer d.wg.Done()

	for {
		select {
		case <-d.ctx.Done():
			return
		case batch := <-d.batchChan:
			if err := d.measurements.SaveAgentMetrics(d.ctx, batch); err != nil {
				d.logger.Error("Error while saving agent metrics", zap.Error(err))
			}
		}
	}
}
