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
	"testing"

	"github.com/stretchr/testify/require"
	"k8s.io/apimachinery/pkg/labels"
)

func TestLabelsFromMap(t *testing.T) {
	tests := []struct {
		name      string
		labels    map[string]string
		result    map[string]string
		errorMsgs []string
	}{
		{
			name:   "empty",
			labels: map[string]string{},
			result: map[string]string{},
		},
		{
			name:   "nil",
			labels: nil,
			result: map[string]string{},
		},
		{
			name: "no errors",
			labels: map[string]string{
				"name": "value",
				"x":    "y",
			},
			result: map[string]string{
				"name": "value",
				"x":    "y",
			},
		},
		{
			name: "no errors",
			labels: map[string]string{
				"name":  "value",
				"na-me": "value",
				"n8me":  "value",
				"1name": "value",
				"name-": "value",
				"name_": "value",
			},
			result: map[string]string{
				"name":  "value",
				"na-me": "value",
				"n8me":  "value",
				"1name": "value",
			},
			errorMsgs: []string{
				"name- is not a valid label name",
				"name_ is not a valid label name",
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			labels, err := LabelsFromMap(test.labels)
			require.Equal(t, test.result, labels.AsMap())
			for _, errorMsg := range test.errorMsgs {
				require.Contains(t, err.Error(), errorMsg)
			}
		})
	}
}

func TestLabelsFromMerge(t *testing.T) {

	testCases := []struct {
		name      string
		oldLabels Labels
		newLabels Labels
		expected  Labels
	}{
		{
			name:      "empty labels",
			oldLabels: MakeLabels(),
			newLabels: MakeLabels(),
			expected:  MakeLabels(),
		},
		{
			name: "overwriting old labels",
			oldLabels: Labels{
				labels.Set{
					"configuration": "old-config",
					"platform":      "Linux",
				},
			},
			newLabels: Labels{
				labels.Set{
					"configuration": "new-config",
				},
			},
			expected: Labels{
				labels.Set{
					"configuration": "new-config",
					"platform":      "Linux",
				},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			require.Equal(t, tc.expected, LabelsFromMerge(tc.oldLabels, tc.newLabels))
		})
	}
}
