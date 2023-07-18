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

package legacy

import (
	"context"
	"testing"

	"github.com/observiq/bindplane-op/model"
	"github.com/observiq/bindplane-op/model/otel"
	bpserver "github.com/observiq/bindplane-op/server"
	servermocks "github.com/observiq/bindplane-op/server/mocks"
	"github.com/observiq/bindplane-op/server/protocol"
	protomocks "github.com/observiq/bindplane-op/server/protocol/mocks"
	"github.com/observiq/bindplane-op/store"
	storemocks "github.com/observiq/bindplane-op/store/mocks"
	"github.com/oklog/ulid/v2"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

var logger = zap.NewNop()

func TestUpdaterStop(t *testing.T) {
	stopCalled := false
	var stop context.CancelFunc = func() {
		stopCalled = true
	}
	updater := newUpdater(protomocks.NewMockProtocol(t), servermocks.NewMockManager(t), stop, logger)
	updater.Stop(context.Background())
	require.True(t, stopCalled)
}

func TestUpdaterHandleMessage(t *testing.T) {
	testCases := []struct {
		name    string
		message bpserver.Message
		setup   func(t *testing.T) (*servermocks.MockManager, *protomocks.MockProtocol)
	}{
		{
			name: "not a snapshot",
			message: bpserver.AgentMessage{
				AgentIDField: "agent-1",
				TypeField:    "not-snapshot",
			},
			setup: func(t *testing.T) (*servermocks.MockManager, *protomocks.MockProtocol) {
				p := protomocks.NewMockProtocol(t)
				p.EXPECT().Connected("agent-1").Return(true)
				return servermocks.NewMockManager(t), p
			},
		},
		{
			name: "request for logs snapshot",
			message: bpserver.AgentMessage{
				AgentIDField: "agentID",
				TypeField:    bpserver.AgentMessageTypeSnapshot,
				BodyField: map[string]any{
					"configuration": protocol.Report{
						Snapshot: protocol.Snapshot{
							PipelineType: otel.Logs,
						},
					},
				},
			},
			setup: func(t *testing.T) (*servermocks.MockManager, *protomocks.MockProtocol) {
				p := protomocks.NewMockProtocol(t)
				p.EXPECT().Connected("agentID").Return(true)
				p.On("RequestReport", mock.Anything, "agentID", protocol.Report{
					Snapshot: protocol.Snapshot{
						PipelineType: otel.Logs,
					},
				}).Return(nil)
				return servermocks.NewMockManager(t), p
			},
		},
	}

	logger := zap.NewNop()
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			manager, proto := tc.setup(t)

			updater := &updater{
				manager:  manager,
				logger:   logger,
				protocol: proto,
				stop:     func() {},
			}
			updater.handleMessage(context.Background(), tc.message)
		})
	}
}

func TestHandleUpdatesEmpty(t *testing.T) {
	testProto := protomocks.NewMockProtocol(t)
	manager := servermocks.NewMockManager(t)
	updater := updater{
		manager:  manager,
		logger:   logger,
		protocol: testProto,
		stop:     func() {},
	}
	updates := store.NewEventUpdates()
	updater.handleUpdates(context.Background(), updates)
}

func TestHandleUpdatesAgents(t *testing.T) {
	testProto := protomocks.NewMockProtocol(t)
	manager := servermocks.NewMockManager(t)

	testAgent := &model.Agent{ID: ulid.Make().String()}
	updates := store.NewEventUpdates()
	updates.IncludeAgent(testAgent, store.EventTypeRollout)
	labels := model.MakeLabels()
	configuration := model.NewConfiguration("config-name")
	agentUpdates := &protocol.AgentUpdates{
		Labels:        &labels,
		Configuration: configuration,
	}

	mockStore := storemocks.NewMockStore(t)
	mockStore.On("AgentConfiguration", mock.Anything, testAgent).Return(configuration, nil)
	manager.On("Store").Return(mockStore)
	testProto.
		On("Connected", testAgent.ID).Return(true).
		On("UpdateAgent", mock.Anything, testAgent, agentUpdates).Return(nil)
	updater := updater{
		manager:  manager,
		logger:   logger,
		protocol: testProto,
		stop:     func() {},
	}
	updater.handleUpdates(context.Background(), updates)
}

func TestHandleUpdatesTwoAgents(t *testing.T) {
	testProto := protomocks.NewMockProtocol(t)

	labels, err := model.LabelsFromSelector("x=y")
	require.NoError(t, err)
	testAgentA := &model.Agent{ID: "A", Labels: labels}
	testAgentB := &model.Agent{ID: "B", Labels: labels}

	manager := servermocks.NewMockManager(t)
	mockStore := storemocks.NewMockStore(t)
	mockStore.On("AgentConfiguration", mock.Anything, testAgentA).Return(nil, nil).
		On("AgentConfiguration", mock.Anything, testAgentB).Return(nil, nil)
	manager.On("Store").Return(mockStore)

	updates := store.NewEventUpdates()
	updates.IncludeAgent(testAgentA, store.EventTypeRollout)
	updates.IncludeAgent(testAgentB, store.EventTypeRollout)

	agentUpdates := &protocol.AgentUpdates{
		Labels: &labels,
	}
	testProto.
		On("Connected", testAgentA.ID).Return(true).
		On("Connected", testAgentB.ID).Return(true).
		On("UpdateAgent", mock.Anything, testAgentA, agentUpdates).Return(nil).
		On("UpdateAgent", mock.Anything, testAgentB, agentUpdates).Return(nil)

	updater := updater{
		manager:  manager,
		logger:   logger,
		protocol: testProto,
		stop:     func() {},
	}
	updater.handleUpdates(context.Background(), updates)
}

func TestHandleUpdatesAgentLabels(t *testing.T) {
	testProto := protomocks.NewMockProtocol(t)

	labels, err := model.LabelsFromMap(map[string]string{
		"w": "x",
		"y": "z",
	})
	require.NoError(t, err)

	testAgentA := &model.Agent{ID: "A", Labels: labels}
	updates := store.NewEventUpdates()
	updates.IncludeAgent(testAgentA, store.EventTypeRollout)

	manager := servermocks.NewMockManager(t)
	mockStore := storemocks.NewMockStore(t)
	mockStore.On("AgentConfiguration", mock.Anything, testAgentA).Return(nil, nil)
	manager.On("Store").Return(mockStore)

	agentUpdates := &protocol.AgentUpdates{
		Labels: &labels,
	}
	testProto.
		On("Connected", testAgentA.ID).Return(true).
		On("UpdateAgent", mock.Anything, testAgentA, agentUpdates).Return(nil)

	updater := updater{
		manager:  manager,
		logger:   logger,
		protocol: testProto,
		stop:     func() {},
	}
	updater.handleUpdates(context.Background(), updates)
}
