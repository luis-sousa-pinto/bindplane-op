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

// Package stopqueue contains a structure for running stop functions in a specific order
package stopqueue

import (
	"context"
	"errors"
	"fmt"
)

// StopFunc is a function that defines how to stop
type StopFunc func(context.Context) error

// NewStopQueue creates a new stop queue
func NewStopQueue() *Queue {
	return &Queue{
		stopFuncs: make([]StopFunc, 0),
	}
}

// Queue is a stack of stop functions that executes in LIFO order
type Queue struct {
	stopFuncs []StopFunc
}

// Add appends the stop func onto the queue
func (s *Queue) Add(stopFunc StopFunc) {
	s.stopFuncs = append(s.stopFuncs, stopFunc)
}

// StopAll runs all stop funcs and clears queue.
// Returns all errors encountered
func (s *Queue) StopAll(ctx context.Context) error {
	var errs error

	for i, stopFunc := range s.stopFuncs {
		if err := stopFunc(ctx); err != nil {
			errs = errors.Join(
				errs,
				fmt.Errorf("stopFunc at index %d err: %w", i, err),
			)
		}
	}

	// Clear out stop funcs
	s.stopFuncs = make([]StopFunc, 0)

	return errs
}
