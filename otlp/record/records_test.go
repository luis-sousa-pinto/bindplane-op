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

package record

import (
	"context"
	"encoding/base64"
	"math"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"go.opentelemetry.io/collector/pdata/pcommon"
	"go.opentelemetry.io/collector/pdata/plog"
	"go.opentelemetry.io/collector/pdata/pmetric"
)

func TestConvertMetrics(t *testing.T) {
	testTimestamp := time.Date(2022, time.September, 15, 1, 1, 1, 1, time.UTC)
	startTestTimestamp := time.Date(2022, time.September, 15, 1, 12, 1, 1, time.UTC)
	testCases := []struct {
		name     string
		input    func() pmetric.Metrics
		expected []*Metric
	}{
		{
			name: "no metrics",
			input: func() pmetric.Metrics {
				return pmetric.NewMetrics()
			},
			expected: []*Metric{},
		},
		{
			name: "All Types",
			input: func() pmetric.Metrics {
				metrics := pmetric.NewMetrics()
				rm := metrics.ResourceMetrics().AppendEmpty()
				rm.Resource().Attributes().PutStr("resource", "one")
				sm := rm.ScopeMetrics().AppendEmpty()

				sumMetric := sm.Metrics().AppendEmpty()
				sumMetric.SetUnit("bytes")
				sumMetric.SetName("sum metric")
				sumDP := sumMetric.SetEmptySum().DataPoints().AppendEmpty()
				sumDP.SetIntValue(1234)
				sumDP.SetTimestamp(pcommon.NewTimestampFromTime(testTimestamp))
				sumDP.SetStartTimestamp(pcommon.NewTimestampFromTime(startTestTimestamp))
				sumDP.Attributes().PutBool("sum", true)

				gaugeMetric := sm.Metrics().AppendEmpty()
				gaugeMetric.SetUnit("psi")
				gaugeMetric.SetName("gauge metric")
				gaugeDP := gaugeMetric.SetEmptyGauge().DataPoints().AppendEmpty()
				gaugeDP.SetDoubleValue(0.1)
				gaugeDP.SetTimestamp(pcommon.NewTimestampFromTime(testTimestamp))
				gaugeDP.SetStartTimestamp(pcommon.NewTimestampFromTime(startTestTimestamp))
				gaugeDP.Attributes().PutDouble("double", 1.2)

				summaryMetric := sm.Metrics().AppendEmpty()
				summaryMetric.SetUnit("fish")
				summaryMetric.SetName("summary metric")
				summaryDP := summaryMetric.SetEmptySummary().DataPoints().AppendEmpty()
				summaryDP.QuantileValues().AppendEmpty().SetValue(0.5)
				summaryDP.SetTimestamp(pcommon.NewTimestampFromTime(testTimestamp))
				summaryDP.SetStartTimestamp(pcommon.NewTimestampFromTime(startTestTimestamp))

				return metrics
			},
			expected: []*Metric{
				{
					Name:           "sum metric",
					Timestamp:      testTimestamp,
					StartTimestamp: startTestTimestamp,
					Value:          int64(1234),
					Unit:           "bytes",
					Type:           pmetric.MetricTypeSum.String(),
					Attributes: map[string]any{
						"sum": true,
					},
					Resource: map[string]any{
						"resource": "one",
					},
				},
				{
					Name:           "gauge metric",
					Timestamp:      testTimestamp,
					StartTimestamp: startTestTimestamp,
					Value:          float64(0.1),
					Unit:           "psi",
					Type:           pmetric.MetricTypeGauge.String(),
					Attributes: map[string]any{
						"double": float64(1.2),
					},
					Resource: map[string]any{
						"resource": "one",
					},
				},
				{
					Name:           "summary metric",
					Timestamp:      testTimestamp,
					StartTimestamp: startTestTimestamp,
					Value: map[string]any{
						"0": float64(0.5),
					},
					Unit:       "fish",
					Type:       pmetric.MetricTypeSummary.String(),
					Attributes: map[string]any{},
					Resource: map[string]any{
						"resource": "one",
					},
				},
			},
		},
		{
			name: "Gauge NaN and Inf filtering",
			input: func() pmetric.Metrics {
				metrics := pmetric.NewMetrics()
				rm := metrics.ResourceMetrics().AppendEmpty()
				rm.Resource().Attributes().PutStr("resource", "one")
				sm := rm.ScopeMetrics().AppendEmpty()

				gaugeMetric := sm.Metrics().AppendEmpty()
				gaugeMetric.SetUnit("psi")
				gaugeMetric.SetName("gauge metric")
				gauge := gaugeMetric.SetEmptyGauge()
				gaugeDP := gauge.DataPoints().AppendEmpty()
				gaugeDP.SetDoubleValue(0.1)
				gaugeDP.SetTimestamp(pcommon.NewTimestampFromTime(testTimestamp))
				gaugeDP.SetStartTimestamp(pcommon.NewTimestampFromTime(startTestTimestamp))
				gaugeDP.Attributes().PutDouble("double", 1.2)

				nanDP := gauge.DataPoints().AppendEmpty()
				nanDP.SetDoubleValue(math.NaN())
				nanDP.SetTimestamp(pcommon.NewTimestampFromTime(testTimestamp))
				nanDP.SetStartTimestamp(pcommon.NewTimestampFromTime(startTestTimestamp))

				infDP := gauge.DataPoints().AppendEmpty()
				infDP.SetDoubleValue(math.Inf(1))
				infDP.SetTimestamp(pcommon.NewTimestampFromTime(testTimestamp))
				infDP.SetStartTimestamp(pcommon.NewTimestampFromTime(startTestTimestamp))

				negInfDP := gauge.DataPoints().AppendEmpty()
				negInfDP.SetDoubleValue(math.Inf(-1))
				negInfDP.SetTimestamp(pcommon.NewTimestampFromTime(testTimestamp))
				negInfDP.SetStartTimestamp(pcommon.NewTimestampFromTime(startTestTimestamp))

				return metrics
			},
			expected: []*Metric{
				{
					Name:           "gauge metric",
					Timestamp:      testTimestamp,
					StartTimestamp: startTestTimestamp,
					Value:          float64(0.1),
					Unit:           "psi",
					Type:           pmetric.MetricTypeGauge.String(),
					Attributes: map[string]any{
						"double": float64(1.2),
					},
					Resource: map[string]any{
						"resource": "one",
					},
				},
			},
		},
		{
			name: "Summary NaN and Inf filtering",
			input: func() pmetric.Metrics {
				metrics := pmetric.NewMetrics()
				rm := metrics.ResourceMetrics().AppendEmpty()
				rm.Resource().Attributes().PutStr("resource", "nineteen")
				sm := rm.ScopeMetrics().AppendEmpty()

				summaryMetric := sm.Metrics().AppendEmpty()
				summaryMetric.SetUnit("requests")
				summaryMetric.SetName("summary metric")
				summary := summaryMetric.SetEmptySummary()
				summaryDP := summary.DataPoints().AppendEmpty()
				summaryDP.SetCount(3)
				summaryDP.SetTimestamp(pcommon.NewTimestampFromTime(testTimestamp))
				summaryDP.SetStartTimestamp(pcommon.NewTimestampFromTime(startTestTimestamp))
				summaryDP.Attributes().PutDouble("double", 1.2)

				// .5 quantile is NaN
				summaryDPQuantile := summaryDP.QuantileValues().AppendEmpty()
				summaryDPQuantile.SetQuantile(0.5)
				summaryDPQuantile.SetValue(math.NaN())

				// .75 quantile is Inf
				summaryDPQuantile = summaryDP.QuantileValues().AppendEmpty()
				summaryDPQuantile.SetQuantile(0.75)
				summaryDPQuantile.SetValue(math.Inf(1))

				// .9 quantile is -Inf
				summaryDPQuantile = summaryDP.QuantileValues().AppendEmpty()
				summaryDPQuantile.SetQuantile(0.9)
				summaryDPQuantile.SetValue(math.Inf(-1))

				// .95 quantile is 19.19
				summaryDPQuantile = summaryDP.QuantileValues().AppendEmpty()
				summaryDPQuantile.SetQuantile(0.95)
				summaryDPQuantile.SetValue(19.19)

				return metrics
			},
			expected: []*Metric{
				{
					Name:           "summary metric",
					Timestamp:      testTimestamp,
					StartTimestamp: startTestTimestamp,
					Value: map[string]any{
						"0.95": float64(19.19),
					},
					Unit: "requests",
					Type: pmetric.MetricTypeSummary.String(),
					Attributes: map[string]any{
						"double": float64(1.2),
					},
					Resource: map[string]any{
						"resource": "nineteen",
					},
				},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			actual := ConvertMetrics(context.Background(), tc.input())
			assert.Equal(t, tc.expected, actual)
		})
	}
}

func TestConvertLogs(t *testing.T) {
	testCases := []struct {
		name     string
		input    func() plog.Logs
		expected []*Log
	}{
		{
			name: "no logs",
			input: func() plog.Logs {
				return plog.NewLogs()
			},
			expected: []*Log{},
		},
		{
			name: "single log with string body",
			input: func() plog.Logs {
				l := plog.NewLogs()

				resource := l.ResourceLogs().AppendEmpty()
				resource.Resource().Attributes().PutStr("resource_id", "unique")

				logRecord := resource.ScopeLogs().AppendEmpty().LogRecords().AppendEmpty()
				logRecord.Attributes().PutStr("custom_field", "custom_value")
				logRecord.Attributes().PutInt("db_id", 22)

				logRecord.SetSeverityText("ERROR")
				logRecord.SetSeverityNumber(plog.SeverityNumberError)
				logRecord.SetTimestamp(pcommon.NewTimestampFromTime(time.Date(2022, time.September, 15, 1, 1, 1, 1, time.UTC)))
				pcommon.NewValueStr("log message").CopyTo(logRecord.Body())
				return l
			},
			expected: []*Log{
				{
					Attributes: map[string]interface{}{
						"custom_field": "custom_value",
						"db_id":        int64(22),
					},
					Body: "log message",
					Resource: map[string]interface{}{
						"resource_id": "unique",
					},
					Timestamp: time.Date(2022, time.September, 15, 1, 1, 1, 1, time.UTC),
					Severity:  "error",
				},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.expected, ConvertLogs(tc.input()))
		})
	}
}

func TestGetLogMessage(t *testing.T) {

	testCases := []struct {
		name     string
		input    func() pcommon.Value
		expected string
	}{
		{
			name: "string value",
			input: func() pcommon.Value {
				return pcommon.NewValueStr("plain string")
			},
			expected: "plain string",
		},
		{
			name: "double value",
			input: func() pcommon.Value {
				return pcommon.NewValueDouble(1248.16)
			},
			expected: "1248.16",
		},
		{
			name: "int value",
			input: func() pcommon.Value {
				return pcommon.NewValueInt(4096)
			},
			expected: "4096",
		},
		{
			name: "bool value",
			input: func() pcommon.Value {
				return pcommon.NewValueBool(true)
			},
			expected: "true",
		},
		{
			name: "bytes value",
			input: func() pcommon.Value {
				// slice := pcommon.NewByteSlice()
				// slice.Append([]byte("log message in bytes")...)
				value := pcommon.NewValueBytes()
				value.SetEmptyBytes().Append([]byte("log message in bytes")...)
				return value
			},
			expected: base64.StdEncoding.EncodeToString([]byte("log message in bytes")),
		},
		{
			name: "empty value",
			input: func() pcommon.Value {
				return pcommon.NewValueEmpty()
			},
			expected: "",
		},
		{
			name: "slice value",
			input: func() pcommon.Value {
				body := pcommon.NewValueSlice()
				body.Slice().AppendEmpty().SetInt(30)
				body.Slice().AppendEmpty().SetInt(60)
				body.Slice().AppendEmpty().SetBool(false)
				body.Slice().AppendEmpty().SetStr("single string")
				return body
			},
			expected: `[30,60,false,"single string"]`,
		},
		{
			name: "map value",
			input: func() pcommon.Value {
				body := pcommon.NewValueMap()
				body.Map().PutStr("key1", "value1")
				body.Map().PutStr("message", "log message")
				body.Map().PutInt("pid", 333)
				return body
			},
			expected: `{
	-> key1: STRING(value1)
	-> message: STRING(log message)
	-> pid: INT(333)
}`,
		},
	}

	t.Parallel()
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.expected, getLogMessage(tc.input()))
		})
	}
}
