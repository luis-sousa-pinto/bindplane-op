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

package label

import (
	"context"
	"fmt"
	"testing"

	"github.com/observiq/bindplane-op/client"
	"github.com/observiq/bindplane-op/client/mocks"
	"github.com/observiq/bindplane-op/model"
	"github.com/stretchr/testify/require"
)

func TestGetAgentLabels(t *testing.T) {
	testCases := []struct {
		name           string
		clientFunc     func() client.BindPlane
		id             string
		expectedLabels *model.Labels
		expectedErr    error
	}{
		{
			name: "success",
			clientFunc: func() client.BindPlane {
				c := mocks.NewMockBindPlane(t)
				labels := &model.Labels{Set: map[string]string{"foo": "bar"}}
				c.On("AgentLabels", context.Background(), "agent-id").Return(labels, nil)
				return c
			},
			id:             "agent-id",
			expectedLabels: &model.Labels{Set: map[string]string{"foo": "bar"}},
		},
		{
			name: "error",
			clientFunc: func() client.BindPlane {
				c := mocks.NewMockBindPlane(t)
				c.On("AgentLabels", context.Background(), "agent-id").Return(nil, fmt.Errorf("error"))
				return c
			},
			id:          "agent-id",
			expectedErr: fmt.Errorf("error"),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			l := NewLabeler(tc.clientFunc())
			labels, err := l.GetAgentLabels(context.Background(), tc.id)
			switch tc.expectedErr {
			case nil:
				require.NoError(t, err)
				require.Equal(t, tc.expectedLabels, labels)
			default:
				require.Error(t, err)
				require.Contains(t, err.Error(), tc.expectedErr.Error())
			}
		})
	}
}

func TestApplyAgentLabels(t *testing.T) {
	testCases := []struct {
		name           string
		clientFunc     func() client.BindPlane
		id             string
		labels         map[string]string
		overwrite      bool
		expectedLabels *model.Labels
		expectedErr    error
	}{
		{
			name: "success",
			clientFunc: func() client.BindPlane {
				c := mocks.NewMockBindPlane(t)
				labels := &model.Labels{Set: map[string]string{"foo": "bar"}}
				c.On("ApplyAgentLabels", context.Background(), "agent-id", labels, false).Return(labels, nil)
				return c
			},
			id:             "agent-id",
			labels:         map[string]string{"foo": "bar"},
			overwrite:      false,
			expectedLabels: &model.Labels{Set: map[string]string{"foo": "bar"}},
		},
		{
			name: "client error",
			clientFunc: func() client.BindPlane {
				c := mocks.NewMockBindPlane(t)
				labels := &model.Labels{Set: map[string]string{"foo": "bar"}}
				c.On("ApplyAgentLabels", context.Background(), "agent-id", labels, false).Return(nil, fmt.Errorf("error"))
				return c
			},
			id:          "agent-id",
			labels:      map[string]string{"foo": "bar"},
			overwrite:   false,
			expectedErr: fmt.Errorf("error"),
		},
		{
			name: "invalid label",
			clientFunc: func() client.BindPlane {
				c := mocks.NewMockBindPlane(t)
				return c
			},
			id:          "agent-id",
			labels:      map[string]string{"": "..."},
			overwrite:   false,
			expectedErr: fmt.Errorf("failed to create labels"),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			l := NewLabeler(tc.clientFunc())
			labels, err := l.ApplyAgentLabels(context.Background(), tc.id, tc.labels, tc.overwrite)
			switch tc.expectedErr {
			case nil:
				require.NoError(t, err)
				require.Equal(t, tc.expectedLabels, labels)
			default:
				require.Error(t, err)
				require.Contains(t, err.Error(), tc.expectedErr.Error())
			}
		})
	}
}
