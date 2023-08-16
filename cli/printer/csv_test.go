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

package printer

import (
	"bytes"
	"testing"

	"github.com/observiq/bindplane-op/model"
	"go.uber.org/zap"
)

type mockPrintable struct {
	titles []string
	values map[string]string
}

func (m *mockPrintable) PrintableKindSingular() string {
	return "Item"
}

func (m *mockPrintable) PrintableKindPlural() string {
	return "Items"
}

func (m *mockPrintable) PrintableFieldTitles() []string {
	return m.titles
}

func (m *mockPrintable) PrintableFieldValue(title string) string {
	return m.values[title]
}

func TestPrintResources(t *testing.T) {
	buffer := &bytes.Buffer{}
	logger := zap.NewNop()
	cp := NewCSVPrinter(buffer, logger)

	item1 := &mockPrintable{
		titles: []string{"Title1", "Title2"},
		values: map[string]string{
			"Title1": "Value1",
			"Title2": "Value2",
		},
	}

	item2 := &mockPrintable{
		titles: []string{"Title1", "Title2"},
		values: map[string]string{
			"Title1": "Value3",
			"Title2": "Value4",
		},
	}

	cp.PrintResources([]model.Printable{item1, item2})

	expected := "Title1,Title2\nValue1,Value2\nValue3,Value4\n"
	if buffer.String() != expected {
		t.Errorf("PrintResources = %v, want %v", buffer.String(), expected)
	}
}

func TestPrintResource(t *testing.T) {
	buffer := &bytes.Buffer{}
	logger := zap.NewNop()
	cp := NewCSVPrinter(buffer, logger)

	item := &mockPrintable{
		titles: []string{"Title1", "Title2"},
		values: map[string]string{
			"Title1": "Value1",
			"Title2": "Value2",
		},
	}

	cp.PrintResource(item)

	expected := "Title1,Title2\nValue1,Value2\n"
	if buffer.String() != expected {
		t.Errorf("PrintResource = %v, want %v", buffer.String(), expected)
	}
}
