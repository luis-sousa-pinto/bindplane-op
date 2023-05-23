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

package sync

import (
	"context"
	"errors"
	"testing"

	"github.com/observiq/bindplane-op/client"
	"github.com/observiq/bindplane-op/client/mocks"
	"github.com/observiq/bindplane-op/model"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestSyncAgentVersions(t *testing.T) {
	testCases := []struct {
		name           string
		version        string
		clientFunc     func() client.BindPlane
		expectedStatus []*model.AnyResourceStatus
		expectedError  error
	}{
		{
			name:    "success",
			version: "1.0.0",
			clientFunc: func() client.BindPlane {
				c := mocks.NewMockBindPlane(t)
				c.On("SyncAgentVersions", mock.Anything, "1.0.0").Return([]*model.AnyResourceStatus{}, nil)
				return c
			},
			expectedStatus: []*model.AnyResourceStatus{},
			expectedError:  nil,
		},
		{
			name:    "error",
			version: "1.0.0",
			clientFunc: func() client.BindPlane {
				c := mocks.NewMockBindPlane(t)
				c.On("SyncAgentVersions", mock.Anything, "1.0.0").Return(nil, errors.New("error"))
				return c
			},
			expectedStatus: nil,
			expectedError:  errors.New("error"),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			s := NewSyncer(tc.clientFunc())
			status, err := s.SyncAgentVersions(context.Background(), tc.version)
			require.Equal(t, tc.expectedStatus, status)
			require.Equal(t, tc.expectedError, err)
		})
	}
}
