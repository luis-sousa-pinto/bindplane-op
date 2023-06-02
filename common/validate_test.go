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

package common

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestValidatePort(t *testing.T) {
	testCases := []struct {
		desc          string
		port          string
		expectedError error
	}{
		{
			desc:          "Not int",
			port:          "not_int",
			expectedError: errors.New("port must be an integer"),
		},
		{
			desc:          "port below min",
			port:          "-1",
			expectedError: errors.New("port must be between"),
		},
		{
			desc:          "port above max",
			port:          "99999",
			expectedError: errors.New("port must be between"),
		},
		{
			desc:          "Valid port",
			port:          "1234",
			expectedError: nil,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {
			err := ValidatePort(tc.port)
			if tc.expectedError != nil {
				require.ErrorContains(t, err, tc.expectedError.Error())
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestValidateURL(t *testing.T) {
	testCases := []struct {
		desc          string
		urlString     string
		schemes       []string
		expectedError error
	}{
		{
			desc:          "Invalid URL",
			urlString:     "invalid\n",
			schemes:       ValidHTTPSchemes,
			expectedError: errors.New("invalid control character in URL"),
		},
		{
			desc:          "Invalid Scheme",
			urlString:     "bad://my.url.com",
			schemes:       ValidHTTPSchemes,
			expectedError: errors.New("scheme 'bad' is invalid"),
		},
		{
			desc:          "Valid URL",
			urlString:     "http://my.url.com",
			schemes:       ValidHTTPSchemes,
			expectedError: nil,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {
			err := ValidateURL(tc.urlString, tc.schemes)
			if tc.expectedError != nil {
				require.ErrorContains(t, err, tc.expectedError.Error())
			} else {
				require.NoError(t, err)
			}
		})
	}
}
