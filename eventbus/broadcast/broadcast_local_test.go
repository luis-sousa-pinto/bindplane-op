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

package broadcast

import (
	"context"
	"testing"
	"time"

	"github.com/observiq/bindplane-op/eventbus"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

type testMessage struct {
	Value int
}

func TestSendReceive(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	b := NewLocalBroadcast[testMessage](ctx, zap.NewNop())

	ch, unsubscribe := eventbus.Subscribe(ctx, b.Consumer())
	defer unsubscribe()

	producer := b.Producer()
	expect := 0
	for _, i := range []int{1, 2, 3} {
		producer.Send(ctx, testMessage{Value: i})
		expect += i
	}

	total := sumMessages(ch, 3)
	require.Equal(t, expect, total)
}

func sumMessages(channel <-chan testMessage, messageCount int) int {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	total := 0
	count := 0
	for {
		select {
		case msg := <-channel:
			count++
			total += msg.Value
			if count == messageCount {
				return total
			}
		case <-ctx.Done():
			return total
		}
	}
}
