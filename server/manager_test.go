// Copyright  observIQ, Inc.
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

	"github.com/observiq/bindplane-op/server/protocol"
	protocolMocks "github.com/observiq/bindplane-op/server/protocol/mocks"
	"github.com/observiq/bindplane-op/store"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

// what we want to test:
//
// add some agents with labels
// add some configurations with matchLabels
// * send labels changes (and associated configuration) to the agent
// * delete a configuration in use by an agent, revert to sample config
// * agent report an unknown configuration, no change

var (
	logger       = zap.NewNop()
	testMapstore = store.NewMapStore(context.Background(), store.Options{
		SessionsSecret:   "super-secret-key",
		MaxEventsToMerge: 1,
	}, logger)
	testProtocol = &protocolMocks.MockProtocol{}
	testManager  = &DefaultManager{
		Storage:   testMapstore,
		Logger:    logger,
		Protocols: []protocol.Protocol{testProtocol},
	}
)

func TestManagerVerifySecretKey(t *testing.T) {
	tests := []struct {
		name             string
		managerSecretKey string
		agentSecretKey   string
		expect           bool
	}{
		{
			name:             "no manager key, no agent key",
			managerSecretKey: "",
			agentSecretKey:   "",
			expect:           true,
		},
		{
			name:             "no manager key, any agent key",
			managerSecretKey: "",
			agentSecretKey:   "any",
			expect:           true,
		},
		{
			name:             "manager key, no agent key",
			managerSecretKey: "test",
			agentSecretKey:   "",
			expect:           false,
		},
		{
			name:             "manager key matches agent key",
			managerSecretKey: "test",
			agentSecretKey:   "test",
			expect:           true,
		},
		{
			name:             "manager key doesn't match agent key",
			managerSecretKey: "test",
			agentSecretKey:   "something else",
			expect:           false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			testManager := &DefaultManager{
				SecretKey: test.managerSecretKey,
			}
			_, ok := testManager.VerifySecretKey(context.TODO(), test.agentSecretKey)
			require.Equal(t, test.expect, ok)
		})
	}
}
