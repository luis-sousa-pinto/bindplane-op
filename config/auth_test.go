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

func TestAuthValidate(t *testing.T) {
	testCases := []struct {
		name     string
		auth     Auth
		expected error
	}{
		{
			name: "valid",
			auth: Auth{
				Username:      "user",
				Password:      "pass",
				SecretKey:     "secret",
				SessionSecret: "session",
			},
		},
		{
			name: "missing username",
			auth: Auth{
				Password:      "pass",
				SecretKey:     "secret",
				SessionSecret: "session",
			},
			expected: errors.New("username must be set"),
		},
		{
			name: "missing password",
			auth: Auth{
				Username:      "user",
				SecretKey:     "secret",
				SessionSecret: "session",
			},
			expected: errors.New("password must be set"),
		},
		{
			name: "missing secret key",
			auth: Auth{
				Username:      "user",
				Password:      "pass",
				SessionSecret: "session",
			},
			expected: errors.New("secret key must be set"),
		},
		{
			name: "missing session secret",
			auth: Auth{
				Username:  "user",
				Password:  "pass",
				SecretKey: "secret",
			},
			expected: errors.New("session secret must be set"),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.auth.Validate()
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
