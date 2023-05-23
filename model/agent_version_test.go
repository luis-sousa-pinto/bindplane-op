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

package model

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestAgentVersionDownload(t *testing.T) {
	tests := []struct {
		name       string
		platform   string
		expectURL  string
		expectHash string
	}{
		{
			name:       "specific release",
			platform:   "darwin/arm64",
			expectURL:  "https://github.com/observIQ/observiq-otel-collector/releases/download/v1.5.0/observiq-otel-collector-v1.5.0-darwin-arm64.tar.gz",
			expectHash: "576fe6d165e7e2a7c293aaceb67d952e5534c3e195927a59b27a08a948b375e5",
		},
		{
			name:       "missing release",
			platform:   "windows/arm64",
			expectURL:  "",
			expectHash: "",
		},
		{
			name:       "os release",
			platform:   "linux/mips",
			expectURL:  "https://github.com/observIQ/observiq-otel-collector/releases/download/v1.5.0/observiq-otel-collector-v1.5.0-amd64.tar.gz",
			expectHash: "f298212e08bfc54ca7dc02339a259375cf07149e186acc2f5803c0255c2391ab",
		},
	}

	version := testResource[*AgentVersion](t, "agentversion-observiq-otel-collector-v1.5.0.yaml")

	require.Equal(t, true, version.Public())

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			download := version.Download(test.platform)
			if download != nil {
				require.Equal(t, test.expectURL, download.URL)
				require.Equal(t, test.expectHash, download.Hash)
			} else {
				require.Equal(t, test.expectURL, "")
				require.Equal(t, test.expectHash, "")
			}
		})
	}
}

func TestAgentVersionBeta(t *testing.T) {
	version := testResource[*AgentVersion](t, "agentversion-observiq-otel-collector-v1.5.0-beta.yaml")
	require.Equal(t, true, version.Public())
}

func TestAgentVersionNoApiVersion(t *testing.T) {
	version := testResource[*AgentVersion](t, "agentversion-observiq-otel-collector-v1.5.0-no-api-version.yaml")
	require.Equal(t, true, version.Public())
}
