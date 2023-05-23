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
	"path/filepath"
	"testing"

	"github.com/observiq/bindplane-op/model/version"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEnsureAPIVersion(t *testing.T) {
	tests := []struct {
		version string
		expect  string
	}{
		{
			version.V1,
			version.V1,
		},
		{
			version.V1Alpha,
			version.V1,
		},
		{
			version.V1Beta,
			version.V1,
		},
		{
			"",
			version.V1,
		},
	}

	for _, test := range tests {
		result := reg.ensureAPIVersion(test.version, KindSource)
		require.Equal(t, test.expect, result)
		result2 := reg.ensureAPIVersion(test.version, KindUnknown)
		require.Equal(t, test.expect, result2)
	}
}

func TestParseSourceTypeStrict_ExtraKeys(t *testing.T) {
	resources, err := ResourcesFromFile(filepath.Join("testfiles", "sourcetype-macos-extra-spec.yaml"))
	assert.NoError(t, err)

	_, err = ParseResourcesStrict(resources)
	require.ErrorContains(t, err, "failed to decode definition:")
}
