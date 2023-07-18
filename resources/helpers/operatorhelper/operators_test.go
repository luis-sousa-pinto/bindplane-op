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

package operatorhelper

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestBpRenderStandardParsingOperator(t *testing.T) {
	testCases := []struct {
		name                  string
		parseFormat           string
		parseTo               string
		regexPattern          string
		parseTimestamp        bool
		timestampField        string
		timezone              string
		parseTimestampFormat  string
		epochTimestampFormat  string
		manualTimestampFormat string
		parseSeverity         bool
		severityField         string
		expected              string
	}{
		{
			name:        "json parsing only",
			parseFormat: "json",
			parseTo:     "body",
			expected:    `{"parse_to":"body","type":"json_parser"}`,
		},
		{
			name:                 "json + timestamp parsing (ISO8601)",
			parseFormat:          "json",
			parseTo:              "body",
			parseTimestamp:       true,
			timestampField:       "timestamp",
			timezone:             "UTC",
			parseTimestampFormat: "ISO8601",
			expected:             `{"parse_to":"body","timestamp":{"layout":"%Y-%m-%dT%H:%M:%S.%f","location":"UTC","parse_from":"body.timestamp"},"type":"json_parser"}`,
		},
		{
			name:                 "json + timestamp parsing (RFC3339)",
			parseFormat:          "json",
			parseTo:              "body",
			parseTimestamp:       true,
			timestampField:       "timestamp",
			timezone:             "UTC",
			parseTimestampFormat: "RFC3339",
			expected:             `{"parse_to":"body","timestamp":{"layout":"2006-01-02T15:04:05.999999999Z07:00","layout_type":"gotime","location":"UTC","parse_from":"body.timestamp"},"type":"json_parser"}`,
		},
		{
			name:                  "json + timestamp parsing (manual)",
			parseFormat:           "json",
			parseTo:               "body",
			parseTimestamp:        true,
			timestampField:        "timestamp",
			timezone:              "UTC",
			parseTimestampFormat:  "Manual",
			manualTimestampFormat: "%Y-%m-%d",
			expected:              `{"parse_to":"body","timestamp":{"layout":"%Y-%m-%d","location":"UTC","parse_from":"body.timestamp"},"type":"json_parser"}`,
		},
		{
			name:                 "json + timestamp parsing (epoch)",
			parseFormat:          "json",
			parseTo:              "body",
			parseTimestamp:       true,
			timestampField:       "timestamp",
			timezone:             "UTC",
			parseTimestampFormat: "Epoch",
			epochTimestampFormat: "s",
			expected:             `{"parse_to":"body","timestamp":{"layout":"s","layout_type":"epoch","location":"UTC","parse_from":"body.timestamp"},"type":"json_parser"}`,
		},
		{
			name:          "json + severity",
			parseFormat:   "json",
			parseTo:       "body",
			parseSeverity: true,
			severityField: "level",
			expected:      `{"parse_to":"body","severity":{"parse_from":"body.level"},"type":"json_parser"}`,
		},
		{
			name:                 "json + timestamp + severity",
			parseFormat:          "json",
			parseTo:              "body",
			parseTimestamp:       true,
			timestampField:       "timestamp",
			timezone:             "UTC",
			parseTimestampFormat: "ISO8601",
			parseSeverity:        true,
			severityField:        "level",
			expected:             `{"parse_to":"body","severity":{"parse_from":"body.level"},"timestamp":{"layout":"%Y-%m-%dT%H:%M:%S.%f","location":"UTC","parse_from":"body.timestamp"},"type":"json_parser"}`,
		},
		// REGEXP
		{
			name:         "regexp parsing only",
			parseFormat:  "regex",
			regexPattern: "^(?P<timestamp>[^ ]*)",
			parseTo:      "body",
			expected:     `{"parse_to":"body","regex":"^(?P<timestamp>[^ ]*)","type":"regex_parser"}`,
		},
		{
			name:                 "regex + timestamp parsing (ISO8601)",
			parseFormat:          "regex",
			regexPattern:         "^(?P<timestamp>[^ ]*)",
			parseTo:              "body",
			parseTimestamp:       true,
			timestampField:       "timestamp",
			timezone:             "UTC",
			parseTimestampFormat: "ISO8601",
			expected:             `{"parse_to":"body","regex":"^(?P<timestamp>[^ ]*)","timestamp":{"layout":"%Y-%m-%dT%H:%M:%S.%f","location":"UTC","parse_from":"body.timestamp"},"type":"regex_parser"}`,
		},
		{
			name:                 "regex + timestamp parsing (RFC3339)",
			parseFormat:          "regex",
			regexPattern:         "^(?P<timestamp>[^ ]*)",
			parseTo:              "body",
			parseTimestamp:       true,
			timestampField:       "timestamp",
			timezone:             "UTC",
			parseTimestampFormat: "RFC3339",
			expected:             `{"parse_to":"body","regex":"^(?P<timestamp>[^ ]*)","timestamp":{"layout":"2006-01-02T15:04:05.999999999Z07:00","layout_type":"gotime","location":"UTC","parse_from":"body.timestamp"},"type":"regex_parser"}`,
		},
		{
			name:                  "regex + timestamp parsing (manual)",
			parseFormat:           "regex",
			regexPattern:          "^(?P<timestamp>[^ ]*)",
			parseTo:               "body",
			parseTimestamp:        true,
			timestampField:        "timestamp",
			timezone:              "UTC",
			parseTimestampFormat:  "Manual",
			manualTimestampFormat: "%Y-%m-%d",
			expected:              `{"parse_to":"body","regex":"^(?P<timestamp>[^ ]*)","timestamp":{"layout":"%Y-%m-%d","location":"UTC","parse_from":"body.timestamp"},"type":"regex_parser"}`,
		},
		{
			name:                 "regex + timestamp parsing (epoch)",
			parseFormat:          "regex",
			regexPattern:         "^(?P<timestamp>[^ ]*)",
			parseTo:              "body",
			parseTimestamp:       true,
			timestampField:       "timestamp",
			timezone:             "UTC",
			parseTimestampFormat: "Epoch",
			epochTimestampFormat: "s",
			expected:             `{"parse_to":"body","regex":"^(?P<timestamp>[^ ]*)","timestamp":{"layout":"s","layout_type":"epoch","location":"UTC","parse_from":"body.timestamp"},"type":"regex_parser"}`,
		},
		{
			name:          "regex + severity",
			parseFormat:   "regex",
			regexPattern:  "^(?P<timestamp>[^ ]*)",
			parseTo:       "body",
			parseSeverity: true,
			severityField: "level",
			expected:      `{"parse_to":"body","regex":"^(?P<timestamp>[^ ]*)","severity":{"parse_from":"body.level"},"type":"regex_parser"}`,
		},
		{
			name:                 "regex + timestamp + severity",
			parseFormat:          "regex",
			regexPattern:         "^(?P<timestamp>[^ ]*)",
			parseTo:              "body",
			parseTimestamp:       true,
			timestampField:       "timestamp",
			timezone:             "UTC",
			parseTimestampFormat: "ISO8601",
			parseSeverity:        true,
			severityField:        "level",
			expected:             `{"parse_to":"body","regex":"^(?P<timestamp>[^ ]*)","severity":{"parse_from":"body.level"},"timestamp":{"layout":"%Y-%m-%dT%H:%M:%S.%f","location":"UTC","parse_from":"body.timestamp"},"type":"regex_parser"}`,
		},
		{
			name:         "regexp parsing (quote, single quote)",
			parseFormat:  "regex",
			regexPattern: `^(?P<timestamp>[^ ]*)'"`,
			parseTo:      "body",
			expected:     `{"parse_to":"body","regex":"^(?P<timestamp>[^ ]*)'\"","type":"regex_parser"}`,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			out := BpRenderStandardParsingOperator(
				tc.parseFormat,
				tc.parseTo,
				tc.regexPattern,
				tc.parseTimestamp,
				tc.timestampField,
				tc.timezone,
				tc.parseTimestampFormat,
				tc.epochTimestampFormat,
				tc.manualTimestampFormat,
				tc.parseSeverity,
				tc.severityField,
			)

			require.Equal(t, tc.expected, out)
		})
	}
}
