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

package metrics

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/observiq/bindplane-op/config"
	"github.com/stretchr/testify/require"
)

func TestNewOTLP(t *testing.T) {
	testCases := []struct {
		name     string
		cfg      *config.OTLPMetrics
		expected error
	}{
		{
			name: "invalid endpoint",
			cfg: &config.OTLPMetrics{
				Endpoint: "localhost",
				Insecure: true,
			},
			expected: errors.New("failed to parse gRPC endpoint"),
		},
		{
			name: "not insecure",
			cfg: &config.OTLPMetrics{
				Endpoint: "localhost:4317",
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := NewOTLP(tc.cfg, 1*time.Minute, nil)
			switch tc.expected {
			case nil:
				require.NoError(t, err)
			default:
				require.Error(t, err)
				require.ErrorContains(t, err, tc.expected.Error())
			}
		})
	}
}

func TestOTLPStart(t *testing.T) {
	testCases := []struct {
		name     string
		cfg      *config.OTLPMetrics
		expected error
	}{
		{
			name: "plain",
			cfg: &config.OTLPMetrics{
				Endpoint: "localhost:4317",
				Insecure: true,
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mp, err := NewOTLP(tc.cfg, 1*time.Minute, nil)

			err = mp.Start(context.Background())
			switch tc.expected {
			case nil:
				require.NoError(t, err)
			default:
				require.Error(t, err)
				require.ErrorContains(t, err, tc.expected.Error())
			}
		})
	}
}

func TestOTLPShutdown(t *testing.T) {
	testCases := []struct {
		name     string
		cfg      *config.OTLPMetrics
		setup    func(Provider) error
		expected error
	}{
		{
			name: "not started",
			cfg: &config.OTLPMetrics{
				Endpoint: "localhost:4317",
				Insecure: true,
			},
			setup: func(_ Provider) error { return nil },
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mp, err := NewOTLP(tc.cfg, 1*time.Minute, nil)
			require.NoError(t, err)

			require.NoError(t, tc.setup(mp))
			err = mp.Shutdown(context.Background())
			switch tc.expected {
			case nil:
				require.NoError(t, err)
			default:
				require.Error(t, err)
				require.ErrorContains(t, err, tc.expected.Error())
			}
		})
	}
}
