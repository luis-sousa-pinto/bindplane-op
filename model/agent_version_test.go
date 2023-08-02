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
	"encoding/hex"
	"testing"

	"github.com/stretchr/testify/require"
)

// TestAgentVersionSpec tests the AgentVersionSpec methods
func TestAgentVersionMethods(t *testing.T) {
	spec := AgentVersionSpec{
		Type:            "testType",
		Version:         "1.0.0",
		ReleaseNotesURL: "https://example.com",
		Draft:           false,
		Prerelease:      true,
		Installer: map[string]AgentInstaller{
			"linux": {
				URL: "https://example.com/installer",
			},
		},
		Download: map[string]AgentDownload{
			"linux": {
				URL:  "https://example.com/download",
				Hash: "123456",
			},
		},
		ReleaseDate: "2023-07-25T00:00:00Z",
	}

	agentVersion := NewAgentVersion(spec)

	require.Equal(t, agentVersion.GetSpec(), spec, "Incorrect spec")

	require.Equal(t, KindAgentVersion, agentVersion.GetKind(), "Incorrect kind")

	require.Equal(t, "testType", agentVersion.AgentType(), "Incorrect type")

	require.Equal(t, "1.0.0", agentVersion.AgentVersion(), "Incorrect version")

	require.Equal(t, false, agentVersion.Public(), "Incorrect public status")

	installer := agentVersion.Installer("linux")
	require.NotNil(t, installer, "Installer not found")
	require.Equal(t, "https://example.com/installer", installer.URL, "Incorrect installer URL")

	download := agentVersion.Download("linux")
	require.NotNil(t, download, "Download not found")
	require.Equal(t, "https://example.com/download", download.URL, "Incorrect download URL")

	semanticVersion := agentVersion.SemanticVersion()
	require.Equal(t, "1.0.0", semanticVersion.String(), "Incorrect semantic version")

	specDownload := spec.Download["linux"]
	expectedHashBytes, _ := hex.DecodeString(download.Hash)
	require.Equal(t, expectedHashBytes, specDownload.HashBytes(), "Incorrect hash bytes")

	t.Run("testInstallerMethodSplit", func(t *testing.T) {
		installer := agentVersion.Installer("linux/amd64")
		require.NotNil(t, installer, "Installer not found for 'linux/amd64'")
		require.Equal(t, "https://example.com/installer", installer.URL, "Incorrect installer URL for 'linux/amd64'")

		installer = agentVersion.Installer("windows")
		require.Nil(t, installer, "Installer should not be found for 'windows'")

		spec.Installer["windows"] = AgentInstaller{URL: "https://example.com/installer-windows"}
		agentVersion = NewAgentVersion(spec)

		installer = agentVersion.Installer("windows/amd64")
		require.NotNil(t, installer, "Installer not found for 'windows/amd64'")
		require.Equal(t, "https://example.com/installer-windows", installer.URL, "Incorrect installer URL for 'windows/amd64'")
	})
}

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

func TestAgentVersionPrintableMethods(t *testing.T) {
	spec := AgentVersionSpec{
		Type:            "test",
		Version:         "1.0.0",
		ReleaseNotesURL: "https://example.com/release",
		Draft:           false,
		Prerelease:      false,
	}

	t.Run("with release date", func(t *testing.T) {
		spec.ReleaseDate = "2023-07-01T00:00:00Z"
		agentVersion := NewAgentVersion(spec)

		expectedTitles := []string{"Name", "Type", "Version", "Public", "Date", "URL"}

		titles := agentVersion.PrintableFieldTitles()
		require.Equal(t, expectedTitles, titles, "PrintableFieldTitles did not return the expected result")

		expectedValues := map[string]string{
			"Name":    "test-1.0.0",
			"Type":    "test",
			"Version": "1.0.0",
			"Public":  "true",
			"Date":    "07-01-2023",
			"URL":     "https://example.com/release",
		}

		for _, title := range titles {
			value := agentVersion.PrintableFieldValue(title)
			expectedValue, ok := expectedValues[title]
			require.True(t, ok, "Unexpected title: %s", title)
			require.Equal(t, expectedValue, value, "PrintableFieldValue for title %s did not return the expected result", title)
		}
	})

	t.Run("without release date", func(t *testing.T) {
		spec.ReleaseDate = ""
		agentVersion := NewAgentVersion(spec)

		expectedValues := map[string]string{
			"Date": "unknown",
		}

		for title, expectedValue := range expectedValues {
			value := agentVersion.PrintableFieldValue(title)
			require.Equal(t, expectedValue, value, "PrintableFieldValue for title %s did not return the expected result", title)
		}
	})

	t.Run("unknown title", func(t *testing.T) {
		spec.ReleaseDate = "2023-07-01T00:00:00Z"
		agentVersion := NewAgentVersion(spec)

		title := "Unknown"
		value := agentVersion.PrintableFieldValue(title)
		expectedValue := agentVersion.ResourceMeta.PrintableFieldValue(title)

		require.Equal(t, expectedValue, value, "PrintableFieldValue for title %s did not return the expected result", title)
	})
}

func TestSortAgentVersionsLatestFirst(t *testing.T) {
	agentVersion1 := &AgentVersion{Spec: AgentVersionSpec{Version: "1.0.0"}}
	agentVersion2 := &AgentVersion{Spec: AgentVersionSpec{Version: "2.0.0"}}
	agentVersion3 := &AgentVersion{Spec: AgentVersionSpec{Version: "0.5.0"}}

	agentVersions := []*AgentVersion{agentVersion1, agentVersion2, agentVersion3}

	SortAgentVersionsLatestFirst(agentVersions)

	// Check if the agent versions are correctly sorted
	require.Equal(t, "2.0.0", agentVersions[0].AgentVersion())
	require.Equal(t, "1.0.0", agentVersions[1].AgentVersion())
	require.Equal(t, "0.5.0", agentVersions[2].AgentVersion())
}
