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

package printer

import (
	"bytes"
	"io"
	"strings"
	"testing"

	"github.com/observiq/bindplane-op/model"
	"github.com/stretchr/testify/require"
)

type testPrintable struct {
	titles []string
	values map[string]string
}

var _ model.Printable = (*testPrintable)(nil)

func (tp *testPrintable) PrintableKindSingular() string {
	return "test"
}
func (tp *testPrintable) PrintableKindPlural() string {
	return "tests"
}
func (tp *testPrintable) PrintableFieldTitles() []string {
	return tp.titles
}
func (tp *testPrintable) PrintableFieldValue(title string) string {
	return tp.values[title]
}

func makeTestPrintables(t *testing.T, titles []string, values []map[string]string) []model.Printable {
	t.Helper()
	var printables []model.Printable
	for _, v := range values {
		printables = append(printables, &testPrintable{titles: titles, values: v})
	}
	return printables
}

func TestPrintEmpty(t *testing.T) {
	titles := []string{"a", "b", "c"}
	values := []map[string]string{}

	printables := makeTestPrintables(t, titles, values)

	buf := &bytes.Buffer{}
	p := NewTablePrinter(buf)
	p.PrintResources(printables)

	require.Equal(t, "No matching resources found.\n", buf.String())
}

func TestPrintSingle(t *testing.T) {
	titles := []string{"a", "b", "c"}
	values := []map[string]string{
		{"a": "1", "b": "2", "c": "3"},
	}

	printables := makeTestPrintables(t, titles, values)

	buf := &bytes.Buffer{}
	p := NewTablePrinter(buf)
	p.PrintResource(printables[0])

	require.Equal(t, "A\tB\tC \n1\t2\t3\t\n", buf.String())
}

func TestTableSort(t *testing.T) {
	titles := []string{"a", "b", "c"}
	values := [][]string{
		{"1", "z", "d"}, // a
		{"2", "y", "f"}, // b
		{"3", "x", "E"}, // c
	}

	valueMaps := []map[string]string{}
	for _, v := range values {
		valueMap := map[string]string{}
		for i, title := range titles {
			valueMap[title] = v[i]
		}
		valueMaps = append(valueMaps, valueMap)
	}

	// spelled out for readability
	tab := "\t"
	nl := "\n"

	tests := []struct {
		name        string
		titles      []string
		expectTable string
	}{
		{
			name:   "already sorted",
			titles: []string{"a", "b", "c"},
			expectTable: "A" + tab + "B" + tab + "C " + nl +
				"1" + tab + "z" + tab + "d" + tab + nl +
				"2" + tab + "y" + tab + "f" + tab + nl +
				"3" + tab + "x" + tab + "E" + tab + nl,
		},
		{
			name:   "lowercase sort",
			titles: []string{"b", "a", "c"},
			expectTable: "B" + tab + "A" + tab + "C " + nl +
				"x" + tab + "3" + tab + "E" + tab + nl +
				"y" + tab + "2" + tab + "f" + tab + nl +
				"z" + tab + "1" + tab + "d" + tab + nl,
		},
		{
			name:   "mixed case sort",
			titles: []string{"c", "b", "a"},
			expectTable: "C" + tab + "B" + tab + "A " + nl +
				"d" + tab + "z" + tab + "1" + tab + nl +
				"E" + tab + "x" + tab + "3" + tab + nl +
				"f" + tab + "y" + tab + "2" + tab + nl,
		},
	}

	for _, test := range tests {
		t.Run(strings.Join(test.titles, "-"), func(t *testing.T) {
			printables := makeTestPrintables(t, test.titles, valueMaps)

			// capture the result in a buffer
			buf := new(bytes.Buffer)
			writer := io.Writer(buf)

			// print the table
			table := NewTablePrinter(writer)
			table.PrintResources(printables)

			// compare the result
			actual := buf.String()
			require.Equal(t, test.expectTable, actual)
		})
	}
}
