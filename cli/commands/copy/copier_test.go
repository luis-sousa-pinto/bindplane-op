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

package copy

import (
	"context"
	"errors"
	"testing"

	"github.com/observiq/bindplane-op/client"
	"github.com/observiq/bindplane-op/client/mocks"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestCopyConfig(t *testing.T) {
	testCases := []struct {
		name        string
		clientFunc  func() client.BindPlane
		cfgName     string
		newCfgName  string
		expectedErr error
	}{
		{
			name: "valid",
			clientFunc: func() client.BindPlane {
				c := mocks.NewMockBindPlane(t)
				c.On("CopyConfig", mock.Anything, "foo", "bar").Return(nil)
				return c
			},
			cfgName:    "foo",
			newCfgName: "bar",
		},
		{
			name: "invalid",
			clientFunc: func() client.BindPlane {
				c := mocks.NewMockBindPlane(t)
				c.On("CopyConfig", mock.Anything, "foo", "bar").Return(errors.New("failed to copy"))
				return c
			},
			cfgName:     "foo",
			newCfgName:  "bar",
			expectedErr: errors.New("failed to copy"),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			c := NewCopier(tc.clientFunc())
			err := c.CopyConfig(context.Background(), tc.cfgName, tc.newCfgName)
			switch tc.expectedErr {
			case nil:
				require.NoError(t, err)
			default:
				require.Error(t, err)
				require.Contains(t, err.Error(), tc.expectedErr.Error())
			}
		})
	}
}
