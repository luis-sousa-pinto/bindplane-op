// Copyright  observIQ, Inc
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

package graphql

import (
	"context"
	"fmt"
	"testing"
	"time"

	model1 "github.com/observiq/bindplane-op/graphql/model"
	"github.com/observiq/bindplane-op/model"
	"github.com/observiq/bindplane-op/otlp/record"
	"github.com/observiq/bindplane-op/server/mocks"
	storeMocks "github.com/observiq/bindplane-op/store/mocks"
	"github.com/observiq/bindplane-op/store/stats"
	measurementsMocks "github.com/observiq/bindplane-op/store/stats/mocks"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func testDestinationMetricWithValue(config string, dest string, destIndex, value any) *record.Metric {
	return &record.Metric{
		Name:           "otelcol_processor_throughputmeasurement_log_data_size",
		Timestamp:      time.Now(),
		StartTimestamp: time.Now(),
		Value:          value,
		Unit:           "B/s",
		Type:           "Rate",
		Attributes: map[string]any{
			"configuration": config,
			"processor":     fmt.Sprintf("throughputmeasurement/_d1_logs_%s-%d", dest, destIndex),
		},
	}
}
func testSourceMetricWithValue(config string, source int, value any) *record.Metric {
	return &record.Metric{
		Name:           "otelcol_processor_throughputmeasurement_log_data_size",
		Timestamp:      time.Now(),
		StartTimestamp: time.Now(),
		Value:          value,
		Unit:           "B/s",
		Type:           "Rate",
		Attributes: map[string]any{
			"configuration": config,
			"processor":     fmt.Sprintf("throughputmeasurement/_d1_logs_source%d", source),
		},
	}
}
func TestOverviewMetrics(t *testing.T) {
	type args struct {
		measurements   stats.MetricData
		destinationIDs []string
		destinations   []*model.Destination
		configIDs      []string
		configurations []*model.Configuration
		agentIDs       map[string][]string
		period         string
		expected       *model1.GraphMetrics
		expectedErr    string
	}
	tests := []struct {
		name string
		args args
	}{
		{
			"Excludes destination and config without agents, but with measurements. c-4 has no agents, d-4 only in c-4",
			args{
				[]*record.Metric{
					testDestinationMetricWithValue("c-1", "d-1", 0, 10),
					testDestinationMetricWithValue("c-2", "d-2", 0, 200),
					testDestinationMetricWithValue("c-3", "d-3", 0, 3000),
					testDestinationMetricWithValue("c-4", "d-4", 0, 40000),
				},
				[]string{"d-1", "d-2", "d-3", "d-4"},
				[]*model.Destination{
					testDestination("d-1"),
					testDestination("d-2"),
					testDestination("d-3"),
					testDestination("d-4"),
				},
				[]string{"c-1", "c-2", "c-3"},
				[]*model.Configuration{
					testConfiguration("c-1", false, []string{"d-1"}),
					testConfiguration("c-2", false, []string{"d-2"}),
					testConfiguration("c-3", false, []string{"d-3"}),
					testConfiguration("c-4", false, []string{"d-4"}),
				},
				map[string][]string{
					"c-1": {"agent1"},
					"c-2": {"agent2"},
					"c-3": {"agent3"},
				},
				"10s",
				&model1.GraphMetrics{
					Metrics: []*model1.GraphMetric{
						{
							Name:         "log_data_size",
							NodeID:       "destination/d-1",
							PipelineType: "",
							Value:        10,
							Unit:         "B/s",
						},
						{
							Name:         "log_data_size",
							NodeID:       "destination/d-2",
							PipelineType: "",
							Value:        200,
							Unit:         "B/s",
						},
						{
							Name:         "log_data_size",
							NodeID:       "destination/d-3",
							PipelineType: "",
							Value:        3000,
							Unit:         "B/s",
						},

						{
							Name:         "log_data_size",
							NodeID:       "configuration/c-1",
							PipelineType: "",
							Value:        10,
							Unit:         "B/s",
						},
						{
							Name:         "log_data_size",
							NodeID:       "configuration/c-2",
							PipelineType: "",
							Value:        200,
							Unit:         "B/s",
						},
						{
							Name:         "log_data_size",
							NodeID:       "configuration/c-3",
							PipelineType: "",
							Value:        3000,
							Unit:         "B/s",
						},
					},
					EdgeMetrics: []*model1.EdgeMetric{
						{
							Name:         "log_data_size",
							EdgeID:       "configuration/c-1|destination/d-1",
							PipelineType: "",
							Value:        10,
							Unit:         "B/s",
						},
						{
							Name:         "log_data_size",
							EdgeID:       "configuration/c-2|destination/d-2",
							PipelineType: "",
							Value:        200,
							Unit:         "B/s",
						},
						{
							Name:         "log_data_size",
							EdgeID:       "configuration/c-3|destination/d-3",
							PipelineType: "",
							Value:        3000,
							Unit:         "B/s",
						},
					},
					MaxMetricValue: 0,
					MaxLogValue:    3000,
					MaxTraceValue:  0,
				},
				"",
			},
		},
		{
			"Everything metrics don't include c-4 and d-4",
			args{
				[]*record.Metric{
					testDestinationMetricWithValue("c-1", "d-1", 0, 10),
					testDestinationMetricWithValue("c-2", "d-2", 0, 200),
					testDestinationMetricWithValue("c-3", "d-3", 0, 3000),
					testDestinationMetricWithValue("c-4", "d-4", 0, 40000),
				},
				[]string{},
				[]*model.Destination{
					testDestination("d-1"),
					testDestination("d-2"),
					testDestination("d-3"),
					testDestination("d-4"),
				},
				[]string{},
				[]*model.Configuration{
					testConfiguration("c-1", false, []string{"d-1"}),
					testConfiguration("c-2", false, []string{"d-2"}),
					testConfiguration("c-3", false, []string{"d-3"}),
					testConfiguration("c-4", false, []string{"d-4"}),
				},
				map[string][]string{
					"c-1": {"agent1"},
					"c-2": {"agent2"},
					"c-3": {"agent3"},
				},
				"10s",
				&model1.GraphMetrics{
					Metrics: []*model1.GraphMetric{

						{
							Name:         "log_data_size",
							NodeID:       "everything/destination",
							PipelineType: "",
							Value:        3210,
							Unit:         "B/s",
						},

						{
							Name:         "log_data_size",
							NodeID:       "everything/configuration",
							PipelineType: "",
							Value:        3210,
							Unit:         "B/s",
						},
					},
					EdgeMetrics: []*model1.EdgeMetric{
						{
							Name:         "log_data_size",
							EdgeID:       "everything/configuration|everything/destination",
							PipelineType: "",
							Value:        3210,
							Unit:         "B/s",
						},
					},
					MaxMetricValue: 0,
					MaxLogValue:    3210,
					MaxTraceValue:  0,
				},

				"",
			},
		},
		{
			"Zombie agent that shares destination with another agent",
			args{
				[]*record.Metric{
					testDestinationMetricWithValue("c-1", "d-1", 0, 10),
					testDestinationMetricWithValue("c-2", "d-1", 0, 200),
				},
				[]string{},
				[]*model.Destination{
					testDestination("d-1"),
				},
				[]string{},
				[]*model.Configuration{
					testConfiguration("c-1", false, []string{"d-1"}),
					testConfiguration("c-2", false, []string{"d-1"}),
				},
				map[string][]string{
					"c-1": {"agent1"},
				},
				"10s",
				&model1.GraphMetrics{
					Metrics: []*model1.GraphMetric{
						{
							Name:         "log_data_size",
							NodeID:       "everything/destination",
							PipelineType: "",
							Value:        10,
							Unit:         "B/s",
						},
						{
							Name:         "log_data_size",
							NodeID:       "everything/configuration",
							PipelineType: "",
							Value:        10,
							Unit:         "B/s",
						},
					},
					EdgeMetrics: []*model1.EdgeMetric{
						{
							Name:         "log_data_size",
							EdgeID:       "everything/configuration|everything/destination",
							PipelineType: "",
							Value:        10,
							Unit:         "B/s",
						},
					},
					MaxMetricValue: 0,
					MaxLogValue:    10,
					MaxTraceValue:  0,
				},

				"",
			},
		},
		// This case is for getting the metrics to determine the Top 3 configurations
		// or destinations.
		{
			"passing in nil config and source IDs returns metrics for all resources",
			args{
				[]*record.Metric{
					testDestinationMetricWithValue("c-1", "d-1", 0, 10),
				},
				nil,
				[]*model.Destination{
					testDestination("d-1"),
				},
				nil,
				[]*model.Configuration{
					testConfiguration("c-1", false, []string{"d-1"}),
				},
				map[string][]string{
					"c-1": {"agent1"},
				},
				"10s",
				&model1.GraphMetrics{
					Metrics: []*model1.GraphMetric{
						{
							Name:         "log_data_size",
							NodeID:       "configuration/c-1",
							PipelineType: "",
							Value:        10,
							Unit:         "B/s",
						},
						{
							Name:         "log_data_size",
							NodeID:       "destination/d-1",
							PipelineType: "",
							Value:        10,
							Unit:         "B/s",
						},
					},
					EdgeMetrics: []*model1.EdgeMetric{
						{
							Name:         "log_data_size",
							EdgeID:       "configuration/c-1|destination/d-1",
							PipelineType: "",
							Value:        10,
							Unit:         "B/s",
						},
					},
					MaxLogValue:    10,
					MaxMetricValue: 0,
					MaxTraceValue:  0,
				},
				"",
			},
		},
		{
			"one configuration has two destinations with different edge metrics",
			args{
				[]*record.Metric{
					testDestinationMetricWithValue("c-1", "d-1", 0, 10),
					testDestinationMetricWithValue("c-1", "d-2", 1, 200),
					testDestinationMetricWithValue("c-2", "d-2", 0, 3000),
				},
				[]string{"d-1", "d-2"},
				[]*model.Destination{testDestination("d-1"), testDestination("d-2")},
				[]string{"c-1", "c-2"},
				[]*model.Configuration{
					testConfiguration("c-1", false, []string{"d-1", "d-2"}),
					testConfiguration("c-2", false, []string{"d-2"}),
				},
				map[string][]string{
					"c-1": {"agent1"},
					"c-2": {"agent2"},
				},
				"10s",
				&model1.GraphMetrics{
					Metrics: []*model1.GraphMetric{
						{
							Name:         "log_data_size",
							NodeID:       "configuration/c-1",
							PipelineType: "",
							Value:        210,
							Unit:         "B/s",
						},
						{
							Name:         "log_data_size",
							NodeID:       "configuration/c-2",
							PipelineType: "",
							Value:        3000,
							Unit:         "B/s",
						},
						{
							Name:         "log_data_size",
							NodeID:       "destination/d-1",
							PipelineType: "",
							Value:        10,
							Unit:         "B/s",
						},
						{
							Name:         "log_data_size",
							NodeID:       "destination/d-2",
							PipelineType: "",
							Value:        3200,
							Unit:         "B/s",
						},
					},
					EdgeMetrics: []*model1.EdgeMetric{
						{
							Name:         "log_data_size",
							EdgeID:       "configuration/c-1|destination/d-1",
							PipelineType: "",
							Value:        10,
							Unit:         "B/s",
						},
						{
							Name:         "log_data_size",
							EdgeID:       "configuration/c-1|destination/d-2",
							PipelineType: "",
							Value:        200,
							Unit:         "B/s",
						},
						{
							Name:         "log_data_size",
							EdgeID:       "configuration/c-2|destination/d-2",
							PipelineType: "",
							Value:        3000,
							Unit:         "B/s",
						},
					},
					MaxMetricValue: 0,
					MaxLogValue:    3000,
					MaxTraceValue:  0,
				},
				"",
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {

			b := mocks.NewMockBindPlane(t)
			s := storeMocks.NewMockStore(t)
			measurementsMocks := measurementsMocks.NewMockMeasurements(t)
			measurementsMocks.On("OverviewMetrics", mock.Anything, mock.Anything).Return(test.args.measurements, nil)

			b.On("Store").Return(s)
			// mock store should return measurements
			s.On("Measurements", mock.Anything).Return(measurementsMocks, nil)
			s.On("Configuration", mock.Anything, mock.Anything).Maybe().Return(
				func(ctx context.Context, name string) *model.Configuration {
					for _, config := range test.args.configurations {
						if config.Name() == name {
							return config
						}
					}
					return nil
				},
				func(ctx context.Context, id string) error {
					return nil
				},
			)
			s.On("AgentsIDsMatchingConfiguration", mock.Anything, mock.Anything).Maybe().Return(
				func(ctx context.Context, config *model.Configuration) []string {
					return test.args.agentIDs[config.Name()]
				},
				func(ctx context.Context, config *model.Configuration) error {
					return nil
				},
			)
			s.On("Configurations", mock.Anything, mock.Anything).Maybe().Return(test.args.configurations, nil)
			s.On("Destination", mock.Anything, mock.Anything).Maybe().Return(
				func(ctx context.Context, name string) *model.Destination {
					for _, destination := range test.args.destinations {
						if destination.Name() == name {
							return destination
						}
					}
					return nil
				},
				func(ctx context.Context, id string) error {
					return nil
				},
			)

			got, err := OverviewMetrics(context.Background(), b, test.args.period, test.args.configIDs, test.args.destinationIDs)
			if test.args.expectedErr != "" {
				require.EqualError(t, err, test.args.expectedErr)
				return
			}
			require.ElementsMatch(t, test.args.expected.Metrics, got.Metrics)
			require.ElementsMatch(t, test.args.expected.EdgeMetrics, got.EdgeMetrics)
			require.Equal(t, test.args.expected.MaxMetricValue, got.MaxMetricValue)
			require.Equal(t, test.args.expected.MaxLogValue, got.MaxLogValue)
			require.Equal(t, test.args.expected.MaxTraceValue, got.MaxTraceValue)

		})
	}
}
