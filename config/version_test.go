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

package config

import (
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestAgentVersionsValidate(t *testing.T) {
	testCases := []struct {
		name     string
		versions AgentVersions
		expected error
	}{
		{
			name: "valid",
			versions: AgentVersions{
				SyncInterval: time.Hour,
			},
		},
		{
			name: "invalid sync interval",
			versions: AgentVersions{
				SyncInterval: 0,
			},
			expected: errors.New("sync interval must be at least 1 hour"),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.versions.Validate()
			switch tc.expected {
			case nil:
				require.NoError(t, err)
			default:
				require.Error(t, err)
				require.Contains(t, err.Error(), tc.expected.Error())
			}
		})
	}
}
