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

// Package operatorhelper contains the helper functions for stanza operators
package operatorhelper

import (
	"bytes"
	"fmt"
	"strings"

	jsoniter "github.com/json-iterator/go"
	"go.opentelemetry.io/collector/pdata/plog"
)

// Use the a special config for jsoniter, because
// we want the keys to be sorted so the output is stable for tests
var operatorMarshalConfig = jsoniter.Config{
	SortMapKeys: true,
}.Froze()

// Parse formats
const (
	jsonParseFormat  = "json"
	regexParseFormat = "regex"
)

// timestamp formats
const (
	iso8601TimestampFormat = "ISO8601"
	rfc3339TimestampFormat = "RFC3339"
	manualTimestampFormat  = "Manual"
	epochTimestampFormat   = "Epoch"
)

// BpRenderStandardParsingOperator renders a parsing operator for stanza-based OTEL receivers as a line of inline YAML.
// This helper function should only be used when parseFormat is not "none", otherwise an empty object is returned
func BpRenderStandardParsingOperator(
	parseFormat string,
	parseTo string,
	regexPattern string,
	parseTimestamp bool,
	timestampField string,
	timezone string,
	parseTimestampFormat string,
	epochTimestampFormat string,
	manualTimestampFormat string,
	parseSeverity bool,
	severityField string,
) string {
	operator := map[string]any{}

	switch parseFormat {
	case jsonParseFormat:
		operator = map[string]any{
			"type":     "json_parser",
			"parse_to": parseTo,
		}
		addSeverityParsingConfig(operator, parseSeverity, parseTo, severityField)
		addTimestampParsingConfig(operator, parseTimestamp, timezone, parseTo, timestampField, parseTimestampFormat, epochTimestampFormat, manualTimestampFormat)
	case regexParseFormat:
		operator = map[string]any{
			"type":     "regex_parser",
			"regex":    regexPattern,
			"parse_to": parseTo,
		}
		addSeverityParsingConfig(operator, parseSeverity, parseTo, severityField)
		addTimestampParsingConfig(operator, parseTimestamp, timezone, parseTo, timestampField, parseTimestampFormat, epochTimestampFormat, manualTimestampFormat)
	}

	// We marshal as JSON - This seems wrong, because we are using this in a YAML
	// file, but YAML is actually a superset of JSON, and marshalling json allows for this to be represented on a single
	// line of a YAML file, so we don't need to worry about indentation.
	buf := &bytes.Buffer{}
	_ = operatorMarshalConfig.NewEncoder(buf).Encode(operator)

	bufStr := buf.String()

	if len(bufStr) == 0 {
		return ""
	}

	// Remove trailing newline
	return bufStr[:len(bufStr)-1]
}

func addSeverityParsingConfig(
	operator map[string]any,
	parseSeverity bool,
	parseTo,
	severityField string,
) {
	if !parseSeverity {
		return
	}

	operator["severity"] = map[string]string{
		"parse_from": fmt.Sprintf("%s.%s", parseTo, severityField),
	}
}

func addTimestampParsingConfig(
	operator map[string]any,
	parseTimestamp bool,
	timezone,
	parseTo,
	timestampField,
	parseTimestampFormat,
	epochFormat,
	manualFormat string,
) {
	if !parseTimestamp {
		return
	}

	timestampConfig := map[string]string{
		"parse_from": fmt.Sprintf("%s.%s", parseTo, timestampField),
	}
	switch parseTimestampFormat {
	case iso8601TimestampFormat:
		timestampConfig["layout"] = "%Y-%m-%dT%H:%M:%S.%f%z"
	case rfc3339TimestampFormat:
		timestampConfig["layout"] = "%Y-%m-%dT%H:%M:%S.%f"
	case manualTimestampFormat:
		timestampConfig["layout"] = manualFormat
	case epochTimestampFormat:
		timestampConfig["layout_type"] = "epoch"
		timestampConfig["layout"] = epochFormat
	}

	timestampConfig["location"] = timezone
	operator["timestamp"] = timestampConfig
}

// BpMapSeverityNameToNumber maps from severity strings to plog SeverityNumbers
func BpMapSeverityNameToNumber(severity string) int {

	s := strings.ToUpper(severity)

	switch s {
	case "TRACE":
		return int(plog.SeverityNumberTrace)
	case "DEBUG":
		return int(plog.SeverityNumberDebug)
	case "INFO":
		return int(plog.SeverityNumberInfo)
	case "WARN", "WARNING":
		return int(plog.SeverityNumberWarn)
	case "ERROR", "ERR":
		return int(plog.SeverityNumberError)
	case "FATAL":
		return int(plog.SeverityNumberFatal)
	}
	return int(plog.SeverityNumberUnspecified)
}
