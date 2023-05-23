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

// Package stats provides the store for measurements associated with agents and configurations.
package stats

import (
	"context"
	"fmt"
	"math"
	"regexp"
	"time"

	"github.com/observiq/bindplane-op/otlp/record"
)

// QueryOptions represents the set of options available for a measurements query
type QueryOptions struct {
	Period time.Duration
}

// MakeQueryOptions constructs a QueryOptions struct from the requested options
func MakeQueryOptions(options []QueryOption) QueryOptions {
	opts := QueryOptions{
		Period: 1 * time.Hour,
	}
	for _, opt := range options {
		opt(&opts)
	}
	return opts
}

// QueryOption is an option used in Store queries
type QueryOption func(*QueryOptions)

// WithPeriod specifies the period for which the metrics should be returned
func WithPeriod(period time.Duration) QueryOption {
	return func(opts *QueryOptions) {
		opts.Period = period
	}
}

// GetDurationFromPeriod returns the rollup duration for a given query period
func GetDurationFromPeriod(opts QueryOptions) time.Duration {
	switch opts.Period {
	case 1 * time.Minute:
		return 10 * time.Second
	case 5 * time.Minute:
		return 1 * time.Minute
	case 1 * time.Hour:
		return 5 * time.Minute
	case 24 * time.Hour:
		return 1 * time.Hour
	default:
		return 1 * time.Hour
	}
}

// MetricData is returned by Measurements when metrics are requested for agents and configurations
type MetricData []*record.Metric

// ----------------------------------------------------------------------

// Measurements provides query and storage of time-series metrics associated with agents and configurations.
//
//go:generate mockery --name Measurements --filename mock_measurements.go --structname MockMeasurements
type Measurements interface {
	// Clear clears the store and is mostly used for testing.
	Clear()

	// MeasurementsSize returns the count of keys in the store, and is used only for testing
	MeasurementsSize(context.Context) (int, error)

	// AgentMetrics provides metrics for an individual agents. They are essentially configuration metrics filtered to a
	// list of agents.
	//
	// Note: While the same record.Metric struct is used to return the metrics, these are not the same metrics provided to
	// Store. They will be aggregated and counter metrics will be converted into rates.
	AgentMetrics(ctx context.Context, id []string, options ...QueryOption) (MetricData, error)

	// ConfigurationMetrics provides all metrics associated with a configuration aggregated from all agents using the
	// configuration.
	//
	// Note: While the same record.Metric struct is used to return the metrics, these are not the same metrics provided to
	// Store. They will be aggregated and counter metrics will be converted into rates.
	ConfigurationMetrics(ctx context.Context, name string, options ...QueryOption) (MetricData, error)

	// OverviewMetrics provides all metrics needed for the overview page. This page shows configurations and destinations.
	// The metrics required are the MeasurementPositionDestinationAfterProcessors metric of each configuration which has a
	// processor with the prefix "throughputmeasurement/_d1_"
	OverviewMetrics(ctx context.Context, options ...QueryOption) (MetricData, error)

	// SaveAgentMetrics saves new metrics. These metrics will be aggregated to determine metrics associated with agents and configurations.
	SaveAgentMetrics(ctx context.Context, metrics []*record.Metric) error

	// ProcessMetrics is called in the background at regular intervals and performs metric roll-up and removes old data
	ProcessMetrics(ctx context.Context) error
}

// RateMetric is a metric that can be used to calculate a rate of change, as the Rate
// function can be called from locations using pmetrics or pgMeasurements
type RateMetric struct {
	Timestamp      time.Time
	StartTimestamp time.Time
	Value          float64
}

// Rate returns the rate of change of the metric over the time period between the two metrics
func Rate(first, last RateMetric) (*RateMetric, error) {
	// If we ended up with start & end being the same moment, or somehow
	// start is later than end, we don't want to provide a 0 or negative rate
	if first.Timestamp.After(last.Timestamp) {
		return nil, fmt.Errorf("first metric timestamp must be before last metric timestamp")
	}

	var duration time.Duration
	var start time.Time
	var delta float64
	// If we're looking at the same "reset" of the counter, we can do meaningful math
	// with the "first" metric's value, otherwise we can only act on the information we
	// have about the latest reset stored in the "last" metric
	if last.StartTimestamp.Equal(first.StartTimestamp) {
		delta = last.Value - first.Value
		// This kind of rollover shouldn't happen any more as it should be caught by the
		// StartTimestamp comparison
		if delta < 0 {
			delta = last.Value
		}

		duration = last.Timestamp.Sub(first.Timestamp)
		start = first.StartTimestamp
		if duration <= 0*time.Second {
			return nil, fmt.Errorf("duration must be greater than 0")
		}
	} else {
		delta = last.Value
		duration = last.Timestamp.Sub(last.StartTimestamp)
		start = last.StartTimestamp
		if duration <= 0*time.Second {
			return nil, fmt.Errorf("duration must be greater than 0")
		}
	}

	rate := delta / duration.Seconds()
	// Reduce to 2 decimal places
	rate = math.Round(rate*100) / 100

	return &RateMetric{
		Timestamp:      last.Timestamp,
		StartTimestamp: start,
		Value:          rate,
	}, nil
}

// ----------------------------------------------------------------------
// Metric methods specific to measurements

// Metric Attribute names used for measurements and Metric Names as they come exported with prometheus
const (
	AgentAttributeName         = "agent"
	ConfigurationAttributeName = "configuration"
	ProcessorAttributeName     = "processor"
	LogDataSizeMetricName      = "otelcol_processor_throughputmeasurement_log_data_size"
	MetricDataSizeMetricName   = "otelcol_processor_throughputmeasurement_metric_data_size"
	TraceDataSizeMetricName    = "otelcol_processor_throughputmeasurement_trace_data_size"
)

// SupportedMetricNames is the list of metrics we care about coming from the self-monitoring of the agent
var SupportedMetricNames = []string{
	LogDataSizeMetricName,
	MetricDataSizeMetricName,
	TraceDataSizeMetricName,
}

// Processor returns the value of the metric attribute that corresponds to the processor name or "" if the attribute is
// missing or not a string.
func Processor(m *record.Metric) string {
	return m.AttributeString(ProcessorAttributeName, "")
}

// ProcessorParsed returns the individual parts of the processor name parsed out of the processor attribute
func ProcessorParsed(m *record.Metric) (position string, pipelineType string, name string) {
	return ParseProcessorName(Processor(m))
}

// Configuration returns the value of the metric attribute that corresponds to the configuration name or "" if the attribute is
// missing or not a string.
func Configuration(m *record.Metric) string {
	return m.AttributeString(ConfigurationAttributeName, "")
}

// Agent returns the value of the metric attribute that corresponds to the processor name or "" if the attribute is
// missing or not a string.
func Agent(m *record.Metric) string {
	return m.AttributeString(AgentAttributeName, "")
}

// Value returns the value of the metric as a float64
func Value(m *record.Metric) (value float64, ok bool) {
	if m.Value == nil {
		return 0, false
	}
	switch value := m.Value.(type) {
	case float64:
		return value, true
	case float32:
		return float64(value), true
	case int64:
		return float64(value), true
	case int32:
		return float64(value), true
	case int:
		return float64(value), true
	}
	return 0, false
}

// ----------------------------------------------------------------------
// measurements utils

var processorNameRegex = regexp.MustCompile("throughputmeasurement/_(.*?)_(.*?)_(.*)")

// ParseProcessorName parses a processor name like throughputmeasurement/_s0_metrics_source0, returning the position,
// type, and name of the processor.
func ParseProcessorName(processorName string) (position string, pipelineType string, name string) {
	matches := processorNameRegex.FindStringSubmatch(processorName)
	if len(matches) == 4 {
		return matches[1], matches[2], matches[3]
	}
	return "", "", ""
}
