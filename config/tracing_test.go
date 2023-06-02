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

func TestTracingValidate(t *testing.T) {
	testCases := []struct {
		name     string
		tracing  Tracing
		expected error
	}{
		{
			name: "valid",
			tracing: Tracing{
				Type: TracerTypeNop,
			},
		},
		{
			name: "invalid type",
			tracing: Tracing{
				Type: "invalid",
			},
			expected: errors.New("invalid tracing type: invalid"),
		},
		{
			name: "invalid sampling rate",
			tracing: Tracing{
				Type:         TracerTypeOTLP,
				SamplingRate: 2,
				OTLP: OTLPTracing{
					Endpoint: "http://localhost:4317",
				},
			},
			expected: errors.New("tracing sampling rate must be between 0 and 1"),
		},
		{
			name: "invalid otlp endpoint",
			tracing: Tracing{
				Type: TracerTypeOTLP,
				OTLP: OTLPTracing{},
			},
			expected: errors.New("OTLP endpoint must be set"),
		},
		{
			name: "invalid google cloud project id",
			tracing: Tracing{
				Type:        TracerTypeGoogleCloud,
				GoogleCloud: GoogleCloudTracing{},
			},
			expected: errors.New("project ID must be set"),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.tracing.Validate()
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
