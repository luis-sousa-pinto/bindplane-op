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

package version

import (
	"context"
	"errors"
	"testing"

	"github.com/observiq/bindplane-op/client"
	"github.com/observiq/bindplane-op/client/mocks"
	"github.com/observiq/bindplane-op/version"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestGetServerVersion(t *testing.T) {
	testCases := []struct {
		name            string
		clientFunc      func() client.BindPlane
		expectedVersion version.Version
		expectedErr     error
	}{
		{
			name: "successful request",
			clientFunc: func() client.BindPlane {
				mockClient := mocks.NewMockBindPlane(t)
				mockClient.On("Version", mock.Anything).Return(version.Version{Tag: "1.0.0"}, nil)
				return mockClient
			},
			expectedVersion: version.Version{Tag: "1.0.0"},
			expectedErr:     nil,
		},
		{
			name: "unsuccessful request",
			clientFunc: func() client.BindPlane {
				mockClient := mocks.NewMockBindPlane(t)
				mockClient.On("Version", mock.Anything).Return(version.Version{}, errors.New("error"))
				return mockClient
			},
			expectedVersion: version.Version{},
			expectedErr:     errors.New("error"),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			v := NewVersioner(tc.clientFunc())
			version, err := v.GetServerVersion(context.Background())
			require.Equal(t, tc.expectedVersion, version)
			require.Equal(t, tc.expectedErr, err)
		})
	}
}

func TestGetClientVersion(t *testing.T) {
	versioner := NewVersioner(nil)
	v, err := versioner.GetClientVersion(context.Background())
	require.Equal(t, version.NewVersion(), v)
	require.NoError(t, err)
}
