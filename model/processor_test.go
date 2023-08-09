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

func TestProcessor_PrintableFieldTitles(t *testing.T) {
	processor := Processor{}

	expectedTitles := []string{"Name", "Type", "Description"}
	require.Equal(t, expectedTitles, processor.PrintableFieldTitles())
}

func TestProcessor_PrintableFieldValue(t *testing.T) {
	processor := Processor{
		ResourceMeta: ResourceMeta{
			Metadata: Metadata{
				Name:        "Processor1",
				Description: "Test Processor",
			},
		},
		Spec: ParameterizedSpec{
			Type: "TestType",
		},
	}

	testCases := []struct {
		Title           string
		ExpectedOutcome string
	}{
		{
			Title:           "Name",
			ExpectedOutcome: "Processor1",
		},
		{
			Title:           "Type",
			ExpectedOutcome: "TestType",
		},
		{
			Title:           "Description",
			ExpectedOutcome: "Test Processor",
		},
		{
			Title:           "Invalid Title",
			ExpectedOutcome: "-",
		},
	}

	for _, tc := range testCases {
		require.Equal(t, tc.ExpectedOutcome, processor.PrintableFieldValue(tc.Title))
	}
}
