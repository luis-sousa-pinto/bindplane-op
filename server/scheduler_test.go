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
	"testing"
	"time"

	"github.com/observiq/bindplane-op/config"
	storemocks "github.com/observiq/bindplane-op/store/mocks"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestScheduler(t *testing.T) {
	mockStore := storemocks.NewMockStore(t)
	mockStore.On("UpdateAllRollouts", mock.Anything).Return(nil)

	// Set this before calling start
	updateRolloutsInterval := 100 * time.Millisecond

	scheduler := NewScheduler(mockStore, zap.NewNop(), updateRolloutsInterval)

	scheduler.Start(context.Background())

	require.Eventually(t, func() bool {
		return mockStore.AssertCalled(t, "UpdateAllRollouts", mock.Anything)
	}, 2*time.Second, 250*time.Millisecond)

	require.Eventually(t, func() bool {
		scheduler.Stop(context.Background())
		return true
	}, 1*time.Second, 100*time.Millisecond)
}

func TestStopSchedulerNoStart(t *testing.T) {
	mockStore := storemocks.NewMockStore(t)
	scheduler := NewScheduler(mockStore, zap.NewNop(), config.DefaultRolloutsInterval)
	err := scheduler.Stop(context.Background())
	require.ErrorContains(t, err, "scheduler was not started")
}
