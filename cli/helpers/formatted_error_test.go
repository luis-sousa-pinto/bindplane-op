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

package helpers

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestFormatError(t *testing.T) {
	testCases := []struct {
		desc        string
		err         error
		expectedErr error
		expected    string
	}{
		{
			desc:        "nil error",
			err:         nil,
			expectedErr: nil,
			expected:    "",
		},
		{
			desc:        "Single Error",
			err:         errors.New("one error"),
			expectedErr: &formattedError{errs: []error{errors.New("one error")}},
			expected:    "1 error occurred:\n\t* one error\n\n",
		},
		{
			desc:        "Joined Error",
			err:         errors.Join(errors.New("one error"), errors.New("two error")),
			expectedErr: &formattedError{errs: []error{errors.New("one error"), errors.New("two error")}},
			expected:    "2 errors occurred:\n\t* one error\n\t* two error\n\n",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {
			err := FormatError(tc.err)
			require.Equal(t, tc.expectedErr, err)
			if err != nil {
				require.Equal(t, tc.expected, err.Error())
			}
		})
	}
}
