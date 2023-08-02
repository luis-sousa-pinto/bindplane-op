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

package model

import (
	"testing"

	"github.com/observiq/bindplane-op/model/version"
	"github.com/stretchr/testify/require"
)

func TestNewSourceType(t *testing.T) {
	name := "TestSourceType"
	parameters := []ParameterDefinition{
		{
			Name:        "TestParameter",
			Description: "This is a test parameter",
			Required:    true,
		},
	}
	supportedPlatforms := []string{"Platform1", "Platform2"}

	sourceType := NewSourceType(name, parameters, supportedPlatforms)

	// Validate the created source type
	require.Equal(t, version.V1, sourceType.APIVersion)
	require.Equal(t, KindSourceType, sourceType.Kind)
	require.Equal(t, name, sourceType.Metadata.Name)

	// Validate the spec
	require.Equal(t, parameters, sourceType.Spec.Parameters)
	require.Equal(t, supportedPlatforms, sourceType.Spec.SupportedPlatforms)
}

func TestNewSourceTypeWithSpec(t *testing.T) {
	name := "TestSourceType"
	spec := ResourceTypeSpec{
		Parameters: []ParameterDefinition{
			{
				Name:        "TestParameter",
				Description: "This is a test parameter",
				Required:    true,
			},
		},
		SupportedPlatforms: []string{"Platform1", "Platform2"},
	}

	sourceType := NewSourceTypeWithSpec(name, spec)

	// Validate the created source type
	require.Equal(t, version.V1, sourceType.APIVersion)
	require.Equal(t, KindSourceType, sourceType.Kind)
	require.Equal(t, name, sourceType.Metadata.Name)

	// Validate the spec
	require.Equal(t, spec, sourceType.Spec)
}
