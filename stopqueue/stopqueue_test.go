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

package stopqueue

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
	"golang.org/x/net/context"
)

func TestNewStopQueue(t *testing.T) {
	expectecd := &Queue{
		stopFuncs: make([]StopFunc, 0),
	}

	actual := NewStopQueue()
	require.Equal(t, expectecd, actual)
}

func TestAddAndStopAll(t *testing.T) {
	oneCalled := false
	stopFunc1 := func(_ context.Context) error {
		oneCalled = true
		return nil
	}

	expectedErr := errors.New("my error")
	twoCalled := false
	stopFunc2 := func(_ context.Context) error {
		twoCalled = true
		return expectedErr
	}

	queue := NewStopQueue()
	queue.Add(stopFunc1)
	queue.Add(stopFunc2)

	require.Len(t, queue.stopFuncs, 2)

	err := queue.StopAll(context.Background())
	require.True(t, oneCalled)
	require.True(t, twoCalled)
	require.ErrorIs(t, err, expectedErr)
}
