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

package opamp

import (
	"testing"

	"github.com/golang/protobuf/proto"
	jsoniter "github.com/json-iterator/go"
	"github.com/observiq/bindplane-op/model"
	"github.com/open-telemetry/opamp-go/protobufs"
	"github.com/stretchr/testify/require"
)

func TestSerializeState(t *testing.T) {

	tests := []struct {
		name   string
		state  *AgentState
		expect *AgentState
	}{
		{
			name:  "nil state",
			state: nil,
			expect: &AgentState{
				SequenceNum: 0,
				Status:      protobufs.AgentToServer{},
			},
		},
		{
			name:  "empty state",
			state: &AgentState{},
			expect: &AgentState{
				SequenceNum: 0,
				Status:      protobufs.AgentToServer{},
			},
		},
		{
			name: "empty status",
			state: &AgentState{
				SequenceNum: 1,
				Status:      protobufs.AgentToServer{},
			},
		},
		{
			name: "full status",
			state: &AgentState{
				SequenceNum: 1,
				Status: protobufs.AgentToServer{
					AgentDescription: &protobufs.AgentDescription{
						IdentifyingAttributes: []*protobufs.KeyValue{
							{
								Key: "id",
								Value: &protobufs.AnyValue{
									Value: &protobufs.AnyValue_StringValue{StringValue: "c1bfe746-82f2-473e-8106-70e8f8e48f9f"},
								},
							},
						},
					},
					EffectiveConfig: &protobufs.EffectiveConfig{
						ConfigMap: &protobufs.AgentConfigMap{
							ConfigMap: map[string]*protobufs.AgentConfigFile{
								"collector.yaml": {Body: []byte("config")},
							},
						},
					},
					RemoteConfigStatus: &protobufs.RemoteConfigStatus{
						Status: protobufs.RemoteConfigStatuses_RemoteConfigStatuses_APPLIED,
					},
					PackageStatuses: &protobufs.PackageStatuses{},
				},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			agentBefore := model.Agent{
				State: EncodeState(test.state),
			}
			// simulate storage with serialization and deserialization
			data, err := jsoniter.Marshal(agentBefore)
			require.NoError(t, err)

			agentAfter := &model.Agent{}
			err = jsoniter.Unmarshal(data, agentAfter)
			require.NoError(t, err)

			actual, err := DecodeState(agentAfter.State)
			require.NoError(t, err)

			expect := test.expect
			if expect == nil {
				expect = test.state
			}

			require.Equal(t, expect.SequenceNum, actual.SequenceNum)
			require.True(t, proto.Equal(&expect.Status, &actual.Status), "protobufs must be equal\nexpect: %v\nactual: %v\n", &expect.Status, &actual.Status)
		})
	}

}
