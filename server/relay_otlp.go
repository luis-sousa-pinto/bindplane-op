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

package server

import (
	"go.opentelemetry.io/collector/pdata/plog"
	"go.opentelemetry.io/collector/pdata/pmetric"
	"go.opentelemetry.io/collector/pdata/ptrace"
)

// initialize pdata marshalers/unmarshalers to reuse
var (
	metricMarshaler   = pmetric.JSONMarshaler{}
	metricUnmarshaler = pmetric.JSONUnmarshaler{}
	logMarshaler      = plog.JSONMarshaler{}
	logUnmarshaler    = plog.JSONUnmarshaler{}
	traceMarshaler    = ptrace.JSONMarshaler{}
	traceUnmarshaler  = ptrace.JSONUnmarshaler{}
)

// RelayMetrics is a wrapper around pmetric.Metrics to allow easier JSON marshalling
type RelayMetrics pmetric.Metrics

// NewRelayMetrics creates a new RelayMetrics from the pmetric.Metrics
func NewRelayMetrics(t pmetric.Metrics) RelayMetrics {
	return RelayMetrics(t)
}

// MarshalJSON marshals to json using pmetric marshaller
func (r RelayMetrics) MarshalJSON() ([]byte, error) {
	return metricMarshaler.MarshalMetrics(pmetric.Metrics(r))
}

// UnmarshalJSON unmarshals from json using pmetric unmarshaller
func (r *RelayMetrics) UnmarshalJSON(data []byte) error {
	newMetrics, err := metricUnmarshaler.UnmarshalMetrics(data)
	if err != nil {
		return err
	}

	*r = RelayMetrics(newMetrics)
	return nil
}

// OTLP returns the pmetric.Metrics type
func (r RelayMetrics) OTLP() pmetric.Metrics {
	return pmetric.Metrics(r)
}

// RelayLogs is a wrapper around plog.Logs to allow easier JSON marshalling
type RelayLogs plog.Logs

// NewRelayLogs creates a new RelayLogs from the plog.Logs
func NewRelayLogs(t plog.Logs) RelayLogs {
	return RelayLogs(t)
}

// MarshalJSON marshals to json using plog marshaller
func (r RelayLogs) MarshalJSON() ([]byte, error) {
	return logMarshaler.MarshalLogs(plog.Logs(r))
}

// UnmarshalJSON unmarshals from json using plog unmarshaller
func (r *RelayLogs) UnmarshalJSON(data []byte) error {
	newMetrics, err := logUnmarshaler.UnmarshalLogs(data)
	if err != nil {
		return err
	}

	*r = RelayLogs(newMetrics)
	return nil
}

// OTLP returns the plog.Logs type
func (r RelayLogs) OTLP() plog.Logs {
	return plog.Logs(r)
}

// RelayTraces is a wrapper around ptrace.Traces to allow easier JSON marshalling
type RelayTraces ptrace.Traces

// NewRelayTraces creates a new RelayTraces from the ptrace.Traces
func NewRelayTraces(t ptrace.Traces) RelayTraces {
	return RelayTraces(t)
}

// MarshalJSON marshals to json using plog marshaller
func (r RelayTraces) MarshalJSON() ([]byte, error) {
	return traceMarshaler.MarshalTraces(ptrace.Traces(r))
}

// UnmarshalJSON unmarshals from json using ptrace unmarshaller
func (r *RelayTraces) UnmarshalJSON(data []byte) error {
	newMetrics, err := traceUnmarshaler.UnmarshalTraces(data)
	if err != nil {
		return err
	}

	*r = RelayTraces(newMetrics)
	return nil
}

// OTLP returns the ptrace.Traces type
func (r RelayTraces) OTLP() ptrace.Traces {
	return ptrace.Traces(r)
}
