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
	"sort"
	"testing"

	"github.com/stretchr/testify/require"
)

type TestPrintable struct {
	FieldTitles map[string]string
}

func (tp *TestPrintable) PrintableKindSingular() string {
	return "TestKind"
}

func (tp *TestPrintable) PrintableKindPlural() string {
	return "TestKinds"
}

func (tp *TestPrintable) PrintableFieldTitles() []string {
	titles := make([]string, 0, len(tp.FieldTitles))
	for title := range tp.FieldTitles {
		titles = append(titles, title)
	}
	return titles
}

func (tp *TestPrintable) PrintableFieldValue(title string) string {
	return tp.FieldTitles[title]
}

func TestPrintableFieldValues(t *testing.T) {
	fieldTitles := map[string]string{
		"Title1": "Value1",
		"Title2": "Value2",
		"Title3": "Value3",
	}
	p := &TestPrintable{FieldTitles: fieldTitles}

	// Sort the keys so that the order is predictable
	keys := make([]string, 0, len(fieldTitles))
	for k := range fieldTitles {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	// Get the values in the same order as the sorted keys
	expectedValues := make([]string, 0, len(keys))
	for _, k := range keys {
		expectedValues = append(expectedValues, fieldTitles[k])
	}

	// Test PrintableFieldValues
	require.ElementsMatch(t, expectedValues, PrintableFieldValues(p))

	// Test PrintableFieldValuesForTitles
	selectedTitles := []string{"Title1", "Title3"}
	expectedSelectedValues := []string{"Value1", "Value3"}
	require.ElementsMatch(t, expectedSelectedValues, PrintableFieldValuesForTitles(p, selectedTitles))
}
