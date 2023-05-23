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

package tracer

import (
	"context"
	"errors"
	"testing"

	"github.com/observiq/bindplane-op/config"
	"github.com/stretchr/testify/require"
)

func TestGoogleCloudStart(t *testing.T) {
	testCases := []struct {
		name     string
		cfg      *config.GoogleCloudTracing
		expected error
	}{
		{
			name: "missing project id",
			cfg: &config.GoogleCloudTracing{
				ProjectID: "",
			},
			expected: errors.New("failed to create span exporter"),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			g := NewGoogleCloud(tc.cfg, 0, nil)
			err := g.Start(context.Background())
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

func TestGoogleCloudShutdown(t *testing.T) {
	testCases := []struct {
		name     string
		cfg      *config.GoogleCloudTracing
		setup    func(t Tracer) error
		expected error
	}{
		{
			name:  "not started",
			cfg:   &config.GoogleCloudTracing{},
			setup: func(t Tracer) error { return nil },
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			g := NewGoogleCloud(tc.cfg, 0, nil)
			err := tc.setup(g)
			require.NoError(t, err)

			err = g.Shutdown(context.Background())
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
