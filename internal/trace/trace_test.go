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

package trace

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSamplingRateOrDefault(t *testing.T) {
	one := 1.0
	zero := 0.0
	belowZero := -0.2
	aboveOne := 1.1

	cases := []struct {
		name     string
		provided *float64
		expected float64
	}{
		{
			"no sampling rate",
			nil,
			1,
		},
		{
			"sampling rate of 0",
			&zero,
			0.0,
		},
		{
			"sampling rate of 1",
			&one,
			1.0,
		},
		{
			"sampling rate above 1",
			&aboveOne,
			1.0,
		},
		{
			"sampling rate below 0",
			&belowZero,
			1.0,
		},
	}

	for _, tc := range cases {
		require.Equal(t, tc.expected, samplingRateOrDefault(tc.provided))
	}
}
