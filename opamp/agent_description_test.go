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

package opamp

import (
	"testing"

	"github.com/open-telemetry/opamp-go/protobufs"
	"github.com/stretchr/testify/require"
)

func TestParseAgentDescription(t *testing.T) {
	testCases := []struct {
		name     string
		msg      *protobufs.AgentDescription
		expected *agentDescription
	}{
		{
			name: "plain description",
			msg: &protobufs.AgentDescription{
				IdentifyingAttributes: []*protobufs.KeyValue{
					{
						Key:   "service.instance.id",
						Value: &protobufs.AnyValue{Value: &protobufs.AnyValue_StringValue{StringValue: "id"}},
					},
					{
						Key:   "service.instance.name",
						Value: &protobufs.AnyValue{Value: &protobufs.AnyValue_StringValue{StringValue: "name"}},
					},
					{
						Key:   "service.name",
						Value: &protobufs.AnyValue{Value: &protobufs.AnyValue_StringValue{StringValue: "type"}},
					},
					{
						Key:   "service.version",
						Value: &protobufs.AnyValue{Value: &protobufs.AnyValue_StringValue{StringValue: "v1.1.1"}},
					},
				},
				NonIdentifyingAttributes: []*protobufs.KeyValue{
					{
						Key:   "service.labels",
						Value: &protobufs.AnyValue{Value: &protobufs.AnyValue_StringValue{StringValue: "label1=value1,label2=value2"}},
					},
					{
						Key:   "os.arch",
						Value: &protobufs.AnyValue{Value: &protobufs.AnyValue_StringValue{StringValue: "arm64"}},
					},
					{
						Key:   "os.details",
						Value: &protobufs.AnyValue{Value: &protobufs.AnyValue_StringValue{StringValue: "Ubuntu 22.04"}},
					},
					{
						Key:   "os.family",
						Value: &protobufs.AnyValue{Value: &protobufs.AnyValue_StringValue{StringValue: "Linux"}},
					},
					{
						Key:   "host.name",
						Value: &protobufs.AnyValue{Value: &protobufs.AnyValue_StringValue{StringValue: "hostname"}},
					},
					{
						Key:   "host.mac_address",
						Value: &protobufs.AnyValue{Value: &protobufs.AnyValue_StringValue{StringValue: "FF:FF:FF:FF:FF:FF"}},
					},
				},
			},
			expected: &agentDescription{
				AgentID:         "id",
				AgentName:       "name",
				AgentType:       "type",
				Version:         "v1.1.1",
				Labels:          "label1=value1,label2=value2",
				Architecture:    "arm64",
				OperatingSystem: "Ubuntu 22.04",
				Platform:        "Linux",
				Hostname:        "hostname",
				MacAddress:      "FF:FF:FF:FF:FF:FF",
			},
		},
		{
			name: "container-platform in labels",
			msg: &protobufs.AgentDescription{
				IdentifyingAttributes: []*protobufs.KeyValue{
					{
						Key:   "service.instance.id",
						Value: &protobufs.AnyValue{Value: &protobufs.AnyValue_StringValue{StringValue: "id"}},
					},
					{
						Key:   "service.instance.name",
						Value: &protobufs.AnyValue{Value: &protobufs.AnyValue_StringValue{StringValue: "name"}},
					},
					{
						Key:   "service.name",
						Value: &protobufs.AnyValue{Value: &protobufs.AnyValue_StringValue{StringValue: "type"}},
					},
					{
						Key:   "service.version",
						Value: &protobufs.AnyValue{Value: &protobufs.AnyValue_StringValue{StringValue: "v1.1.19"}},
					},
				},
				NonIdentifyingAttributes: []*protobufs.KeyValue{
					{
						Key:   "service.labels",
						Value: &protobufs.AnyValue{Value: &protobufs.AnyValue_StringValue{StringValue: "label1=value1,container-platform=kubernetes"}},
					},
					{
						Key:   "os.arch",
						Value: &protobufs.AnyValue{Value: &protobufs.AnyValue_StringValue{StringValue: "arm64"}},
					},
					{
						Key:   "os.details",
						Value: &protobufs.AnyValue{Value: &protobufs.AnyValue_StringValue{StringValue: "Ubuntu 22.04"}},
					},
					{
						Key:   "os.family",
						Value: &protobufs.AnyValue{Value: &protobufs.AnyValue_StringValue{StringValue: "Linux"}},
					},
					{
						Key:   "host.name",
						Value: &protobufs.AnyValue{Value: &protobufs.AnyValue_StringValue{StringValue: "hostname"}},
					},
					{
						Key:   "host.mac_address",
						Value: &protobufs.AnyValue{Value: &protobufs.AnyValue_StringValue{StringValue: "FF:FF:FF:FF:FF:EE"}},
					},
				},
			},
			expected: &agentDescription{
				AgentID:         "id",
				AgentName:       "name",
				AgentType:       "type",
				Version:         "v1.1.19",
				Labels:          "label1=value1,container-platform=kubernetes",
				Architecture:    "arm64",
				OperatingSystem: "Ubuntu 22.04",
				Platform:        "kubernetes",
				Hostname:        "hostname",
				MacAddress:      "FF:FF:FF:FF:FF:EE",
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			require.Equal(t, tc.expected, parseAgentDescription(tc.msg))
		})
	}
}
