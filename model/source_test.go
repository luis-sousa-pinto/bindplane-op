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

	"github.com/stretchr/testify/require"
)

func TestSourcePrintableFields(t *testing.T) {
	source := &Source{
		ResourceMeta: ResourceMeta{
			Metadata: Metadata{
				Name:        "TestName",
				Description: "TestDescription",
			},
		},
		Spec: ParameterizedSpec{
			Type: "TestType",
		},
	}

	expectedTitles := []string{"Name", "Type", "Description"}
	actualTitles := source.PrintableFieldTitles()
	require.Equal(t, expectedTitles, actualTitles)

	for _, title := range actualTitles {
		var expectedValue string
		switch title {
		case "Name":
			expectedValue = source.Metadata.Name
		case "Type":
			expectedValue = source.Spec.Type
		case "Description":
			expectedValue = source.Metadata.Description
		default:
			t.Fatalf("Unexpected field title: %s", title)
		}

		actualValue := source.PrintableFieldValue(title)
		require.Equal(t, expectedValue, actualValue, "Field value for title %s does not match expected", title)
	}
}
