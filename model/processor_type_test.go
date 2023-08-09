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

func TestNewProcessorType(t *testing.T) {
	name := "TestProcessorType"
	parameters := []ParameterDefinition{
		{
			Name: "Param1",
			Type: "string",
		},
		{
			Name: "Param2",
			Type: "integer",
		},
	}
	processorType := NewProcessorType(name, parameters)

	require.NotNil(t, processorType)
	require.Equal(t, name, processorType.Metadata.Name)
	require.Equal(t, version.V1, processorType.APIVersion)
	require.Equal(t, KindProcessorType, processorType.Kind)
	require.Equal(t, parameters, processorType.Spec.Parameters)
}

func TestNewProcessorTypeWithSpec(t *testing.T) {
	name := "TestProcessorTypeWithSpec"
	spec := ResourceTypeSpec{
		Version: "1.0",
		Parameters: []ParameterDefinition{
			{
				Name: "Param1",
				Type: "string",
			},
			{
				Name: "Param2",
				Type: "integer",
			},
		},
	}
	processorTypeWithSpec := NewProcessorTypeWithSpec(name, spec)

	require.NotNil(t, processorTypeWithSpec)
	require.Equal(t, name, processorTypeWithSpec.Metadata.Name)
	require.Equal(t, version.V1, processorTypeWithSpec.APIVersion)
	require.Equal(t, KindProcessorType, processorTypeWithSpec.Kind)
	require.Equal(t, spec, processorTypeWithSpec.Spec)
}
