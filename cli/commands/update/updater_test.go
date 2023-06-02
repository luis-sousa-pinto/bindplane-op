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

package update

import (
	"context"
	"errors"
	"testing"

	"github.com/observiq/bindplane-op/client"
	"github.com/observiq/bindplane-op/client/mocks"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestUpgradeAgent(t *testing.T) {
	testCases := []struct {
		name       string
		id         string
		version    string
		clientFunc func() client.BindPlane
		expected   error
	}{
		{
			name:    "success",
			id:      "agent-id",
			version: "1.0.0",
			clientFunc: func() client.BindPlane {
				c := mocks.NewMockBindPlane(t)
				c.On("AgentUpgrade", mock.Anything, "agent-id", "1.0.0").Return(nil)
				return c
			},
		},
		{
			name:    "error",
			id:      "agent-id",
			version: "1.0.0",
			clientFunc: func() client.BindPlane {
				c := mocks.NewMockBindPlane(t)
				c.On("AgentUpgrade", mock.Anything, "agent-id", "1.0.0").Return(errors.New("error"))
				return c
			},
			expected: errors.New("error"),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			u := NewUpdater(tc.clientFunc())
			err := u.UpdateAgent(context.Background(), tc.id, tc.version)
			require.Equal(t, tc.expected, err)
		})
	}
}
