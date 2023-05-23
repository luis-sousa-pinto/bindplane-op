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

package agent

import (
	"testing"

	"github.com/stretchr/testify/require"
)

// TODO: Find a way to test this without getting false positives in CI

func TestGithubVersion(t *testing.T) {
	t.Skip("Not hitting live GitHub API")
	github := newGithub()
	version, err := github.Version("v1.4.0")
	require.NoError(t, err)
	require.Equal(t, "v1.4.0", version.Version())
	require.True(t, version.Public(), "should be public")
	require.Equal(t, "https://github.com/observIQ/observiq-otel-collector/releases/download/v1.4.0/observiq-otel-collector-v1.4.0-darwin-amd64.tar.gz", version.Download("darwin/amd64").URL)
	require.Equal(t, "c7129c5dc69ec9c3fe3ae6864ff3e9960ae69ccf86f1cd5bcdacff5cf107ab87", version.Download("darwin/amd64").Hash)
	require.Equal(t, "https://github.com/observIQ/observiq-otel-collector/releases/download/v1.4.0/install_macos.sh", version.Installer("darwin/amd64").URL)
}

func TestGithubLatestVersion(t *testing.T) {
	t.Skip("Not hitting live GitHub API")
	github := newGithub()

	// since the latest version will change over time, we just want to make sure that we get reasonable results.
	version, err := github.LatestVersion()
	require.NoError(t, err)
	require.Contains(t, version.Version(), "v")
	require.Contains(t, version.Download("darwin/amd64").URL, "https://github.com/observIQ/observiq-otel-collector/releases/download/")
	require.Contains(t, version.Installer("darwin/amd64").URL, "https://github.com/observIQ/observiq-otel-collector/releases/download/")
}
