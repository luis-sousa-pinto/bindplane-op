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
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"testing"

	"github.com/jarcoal/httpmock"
	jsoniter "github.com/json-iterator/go"
	"github.com/observiq/bindplane-op/model"
	"github.com/stretchr/testify/require"
)

func TestGithubVersion(t *testing.T) {
	t.Run("happy path", func(t *testing.T) {
		github := newGithub()

		httpmock.ActivateNonDefault(github.client.GetClient())
		defer httpmock.DeactivateAndReset()
		setMockResponders(t)

		expectedVersion := readJSONFile[*model.AgentVersion](t, filepath.Join("testfiles", "agent-version-v1.30.0.json"))

		version, err := github.Version("v1.30.0")
		require.NoError(t, err)

		require.Equal(t, expectedVersion, version)
	})

	t.Run("not json", func(t *testing.T) {
		github := newGithub()

		httpmock.ActivateNonDefault(github.client.GetClient())
		defer httpmock.DeactivateAndReset()
		setMockResponders(t)

		httpmock.RegisterResponder(
			"GET",
			"https://api.github.com/repos/observIQ/bindplane-agent/releases/tags/v1.30.0",
			jsonResponder(200, []byte("not json")),
		)

		_, err := github.Version("v1.30.0")
		require.ErrorContains(t, err, "get release:")
	})

	t.Run("release not found", func(t *testing.T) {
		github := newGithub()

		httpmock.ActivateNonDefault(github.client.GetClient())
		defer httpmock.DeactivateAndReset()
		setMockResponders(t)

		httpmock.RegisterResponder(
			"GET",
			"https://api.github.com/repos/observIQ/bindplane-agent/releases/tags/v1.30.0",
			jsonResponder(404, []byte{}),
		)

		version, err := github.Version("v1.30.0")
		require.Nil(t, version)
		require.ErrorIs(t, err, ErrVersionNotFound)
	})

	t.Run("release returns 500", func(t *testing.T) {
		github := newGithub()

		httpmock.ActivateNonDefault(github.client.GetClient())
		defer httpmock.DeactivateAndReset()
		setMockResponders(t)
		httpmock.RegisterResponder(
			"GET",
			"https://api.github.com/repos/observIQ/bindplane-agent/releases/tags/v1.30.0",
			jsonResponder(500, []byte{}),
		)

		version, err := github.Version("v1.30.0")
		require.Nil(t, version)
		require.ErrorContains(t, err, "unable to get version v1.30.0:")
	})

	t.Run("failure getting sums", func(t *testing.T) {
		github := newGithub()

		httpmock.ActivateNonDefault(github.client.GetClient())
		defer httpmock.DeactivateAndReset()
		setMockResponders(t)
		httpmock.RegisterResponder(
			"GET",
			"https://github.com/observIQ/bindplane-agent/releases/download/v1.30.0/observiq-otel-collector-v1.30.0-SHA256SUMS",
			jsonResponder(404, []byte{}),
		)

		version, err := github.Version("v1.30.0")
		require.Nil(t, version)
		require.ErrorContains(t, err, "get sha256 sums:")
	})

}

func TestGithubLatestVersion(t *testing.T) {
	t.Run("happy path", func(t *testing.T) {
		github := newGithub()

		httpmock.ActivateNonDefault(github.client.GetClient())
		defer httpmock.DeactivateAndReset()
		setMockResponders(t)

		expectedVersion := readJSONFile[*model.AgentVersion](t, filepath.Join("testfiles", "agent-version-v1.30.0.json"))

		version, err := github.LatestVersion()
		require.NoError(t, err)

		require.Equal(t, expectedVersion, version)
	})

	t.Run("not json", func(t *testing.T) {
		github := newGithub()

		httpmock.ActivateNonDefault(github.client.GetClient())
		defer httpmock.DeactivateAndReset()
		setMockResponders(t)

		httpmock.RegisterResponder(
			"GET",
			"https://api.github.com/repos/observIQ/bindplane-agent/releases/latest",
			jsonResponder(200, []byte("not json")),
		)

		version, err := github.LatestVersion()
		require.Nil(t, version)
		require.ErrorContains(t, err, "get release:")
	})

	t.Run("release not found", func(t *testing.T) {
		github := newGithub()

		httpmock.ActivateNonDefault(github.client.GetClient())
		defer httpmock.DeactivateAndReset()
		setMockResponders(t)

		httpmock.RegisterResponder(
			"GET",
			"https://api.github.com/repos/observIQ/bindplane-agent/releases/latest",
			jsonResponder(404, []byte{}),
		)

		version, err := github.LatestVersion()
		require.Nil(t, version)
		require.ErrorIs(t, err, ErrVersionNotFound)
	})

	t.Run("release returns 500", func(t *testing.T) {
		github := newGithub()

		httpmock.ActivateNonDefault(github.client.GetClient())
		defer httpmock.DeactivateAndReset()
		setMockResponders(t)
		httpmock.RegisterResponder(
			"GET",
			"https://api.github.com/repos/observIQ/bindplane-agent/releases/latest",
			jsonResponder(500, []byte{}),
		)

		version, err := github.LatestVersion()
		require.Nil(t, version)
		require.ErrorContains(t, err, "unable to get version latest:")
	})

	t.Run("failure getting sums", func(t *testing.T) {
		github := newGithub()

		httpmock.ActivateNonDefault(github.client.GetClient())
		defer httpmock.DeactivateAndReset()
		setMockResponders(t)
		httpmock.RegisterResponder(
			"GET",
			"https://github.com/observIQ/bindplane-agent/releases/download/v1.30.0/observiq-otel-collector-v1.30.0-SHA256SUMS",
			jsonResponder(404, []byte{}),
		)

		version, err := github.LatestVersion()
		require.Nil(t, version)
		require.ErrorContains(t, err, "get sha256 sums:")
	})

}

func TestGithubVersions(t *testing.T) {
	t.Run("happy path", func(t *testing.T) {
		github := newGithub()

		httpmock.ActivateNonDefault(github.client.GetClient())
		defer httpmock.DeactivateAndReset()
		setMockResponders(t)

		expectedVersions := readJSONFile[[]*model.AgentVersion](t, filepath.Join("testfiles", "agent-versions.json"))

		versions, err := github.Versions()
		require.NoError(t, err)
		require.Equal(t, expectedVersions, versions)
	})

	t.Run("not json", func(t *testing.T) {
		github := newGithub()

		httpmock.ActivateNonDefault(github.client.GetClient())
		defer httpmock.DeactivateAndReset()
		setMockResponders(t)

		httpmock.RegisterResponder(
			"GET",
			"https://api.github.com/repos/observIQ/bindplane-agent/releases",
			jsonResponder(200, []byte("not json")),
		)

		versions, err := github.Versions()
		require.Nil(t, versions)
		require.ErrorContains(t, err, "get releases:")
	})

	t.Run("releases returns 500", func(t *testing.T) {
		github := newGithub()

		httpmock.ActivateNonDefault(github.client.GetClient())
		defer httpmock.DeactivateAndReset()
		setMockResponders(t)
		httpmock.RegisterResponder(
			"GET",
			"https://api.github.com/repos/observIQ/bindplane-agent/releases",
			jsonResponder(500, []byte{}),
		)

		version, err := github.Versions()
		require.Nil(t, version)
		require.ErrorContains(t, err, "unable to get versions:")
	})

	t.Run("failure getting sums", func(t *testing.T) {
		github := newGithub()

		httpmock.ActivateNonDefault(github.client.GetClient())
		defer httpmock.DeactivateAndReset()
		setMockResponders(t)
		httpmock.RegisterResponder(
			"GET",
			"https://github.com/observIQ/bindplane-agent/releases/download/v1.30.0/observiq-otel-collector-v1.30.0-SHA256SUMS",
			jsonResponder(404, []byte{}),
		)

		version, err := github.Versions()
		require.Nil(t, version)
		require.ErrorContains(t, err, "get sha256 sums:")
	})
}

func setMockResponders(t *testing.T) {
	releases := readFile(t, filepath.Join("testfiles", "github-releases-response.json"))
	latest := readFile(t, filepath.Join("testfiles", "github-latest-response.json"))
	tag := readFile(t, filepath.Join("testfiles", "github-tag-response.json"))
	sums := readFile(t, filepath.Join("testfiles", "github-v1.30.0-sums-response.sums"))

	httpmock.RegisterResponder(
		"GET",
		"https://api.github.com/repos/observIQ/bindplane-agent/releases",
		jsonResponder(200, releases),
	)
	httpmock.RegisterResponder(
		"GET",
		"https://api.github.com/repos/observIQ/bindplane-agent/releases/latest",
		jsonResponder(200, latest),
	)
	httpmock.RegisterResponder(
		"GET",
		"https://api.github.com/repos/observIQ/bindplane-agent/releases/tags/v1.30.0",
		jsonResponder(200, tag),
	)
	httpmock.RegisterResponder(
		"GET",
		"https://github.com/observIQ/bindplane-agent/releases/download/v1.30.0/observiq-otel-collector-v1.30.0-SHA256SUMS",
		httpmock.NewBytesResponder(200, sums),
	)
	httpmock.RegisterResponder(
		"GET",
		"https://github.com/observIQ/bindplane-agent/releases/download/v1.29.1/observiq-otel-collector-v1.29.1-SHA256SUMS",
		httpmock.NewBytesResponder(200, sums),
	)
	httpmock.RegisterResponder(
		"GET",
		"https://github.com/observIQ/bindplane-agent/releases/download/v1.29.0/observiq-otel-collector-v1.29.0-SHA256SUMS",
		httpmock.NewBytesResponder(200, sums),
	)
}

func readFile(t *testing.T, path string) []byte {
	b, err := os.ReadFile(path)
	require.NoError(t, err)
	return b
}

func jsonResponder(status int, body []byte) httpmock.Responder {
	return httpmock.ResponderFromResponse(
		&http.Response{
			Status:     strconv.Itoa(status),
			StatusCode: status,
			Body:       httpmock.NewRespBodyFromBytes(body),
			Header: http.Header{
				// We need this for resty to properly unmarshal the response
				"Content-Type": []string{"application/json"},
			},
			ContentLength: -1,
		},
	)
}

// readJSONFile generically reads the generic type from a filepath where the file contains json data.
func readJSONFile[T any](t *testing.T, path string) T {
	b, err := os.ReadFile(path)
	require.NoError(t, err)

	var o T
	err = jsoniter.Unmarshal(b, &o)
	require.NoError(t, err)

	return o
}

// writeJSONfile is a helper function can be used to write test files as JSON.
func writeJSONFile(t *testing.T, path string, o any) {
	b, err := jsoniter.Marshal(o)
	require.NoError(t, err)

	err = os.WriteFile(path, b, 0666)
	require.NoError(t, err)
}
