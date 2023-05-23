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

package server

import (
	"context"
	"errors"
	"sync"
	"time"

	"github.com/observiq/bindplane-op/store"
	"go.uber.org/zap"
)

// Scheduler schedules periodic background tasks
type Scheduler interface {
	// Start starts the scheduler which will begin executing background tasks.
	Start(context.Context)

	// Stop stops the scheduler
	Stop(context.Context) error
}

// defaultScheduler is the default implementation of scheduler
type defaultScheduler struct {
	store    store.Store
	logger   *zap.Logger
	interval time.Duration

	ctx    context.Context
	cancel context.CancelCauseFunc
	wg     sync.WaitGroup
}

// NewScheduler creates a new scheduler
func NewScheduler(s store.Store, logger *zap.Logger, interval time.Duration) Scheduler {
	return &defaultScheduler{
		store:    s,
		logger:   logger,
		interval: interval,
	}
}

// Start starts the scheduler
func (s *defaultScheduler) Start(ctx context.Context) {
	s.ctx, s.cancel = context.WithCancelCause(ctx)
	s.wg.Add(1)
	go func() {
		defer s.wg.Done()

		updateRolloutsTicker := time.NewTicker(s.interval)
		defer updateRolloutsTicker.Stop()
		for {
			select {
			case <-s.ctx.Done():
				return
			case <-updateRolloutsTicker.C:
				if err := s.store.UpdateAllRollouts(s.ctx); err != nil {
					s.logger.Error("failed to update rollouts", zap.Error(err))
				}
			}
		}
	}()
}

// Stop stops the scheduler and all running tasks
func (s *defaultScheduler) Stop(ctx context.Context) error {
	if s.cancel == nil {
		return errors.New("scheduler was not started")
	}
	s.cancel(errors.New("stop called"))

	doneChan := make(chan struct{})
	go func() {
		s.wg.Wait()
		close(doneChan)
	}()

	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-doneChan:
		return nil
	}
}
