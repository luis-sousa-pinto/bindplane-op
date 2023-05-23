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

	"github.com/stretchr/testify/require"
)

func TestStoreValidate(t *testing.T) {
	testCases := []struct {
		name     string
		store    Store
		expected error
	}{
		{
			name: "valid",
			store: Store{
				Type:      StoreTypeMap,
				MaxEvents: 100,
			},
		},
		{
			name: "invalid type",
			store: Store{
				Type: "invalid",
			},
			expected: errors.New("invalid store type: invalid"),
		},
		{
			name: "invalid max events",
			store: Store{
				Type:      StoreTypeMap,
				MaxEvents: 0,
			},
			expected: errors.New("maxEvents must be greater than 0"),
		},
		{
			name: "invalid bbolt path",
			store: Store{
				Type:  StoreTypeBBolt,
				BBolt: BBolt{},
			},
			expected: errors.New("bbolt path must be set"),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.store.Validate()
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
