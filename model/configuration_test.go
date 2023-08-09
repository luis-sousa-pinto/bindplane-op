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
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/observiq/bindplane-op/model/version"
	mock "github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

func validateResource[T Resource](t *testing.T, name string) T {
	return fileResource[T](t, filepath.Join("testfiles", "validate", name))
}
func testResource[T Resource](t *testing.T, name string) T {
	return fileResource[T](t, filepath.Join("testfiles", name))
}

func newTestResourceConfiguration(t *testing.T) *ResourceConfiguration {
	t.Helper()
	value := 10
	return &ResourceConfiguration{
		ID:          "one",
		Name:        "RC",
		DisplayName: "Resource Configuration",
		ParameterizedSpec: ParameterizedSpec{
			Type: "macOS:1",
			Parameters: []Parameter{
				{
					Name:      "param1",
					Sensitive: false,
					Value:     &value,
				},
			},
			Disabled: false,
		},
	}
}

type testConfiguration struct {
	bindplaneURL                string
	bindplaneInsecureSkipVerify bool
}

func newTestConfiguration() *testConfiguration {
	return &testConfiguration{}
}

func (c *testConfiguration) BindPlaneURL() string {
	return c.bindplaneURL
}

func (c *testConfiguration) BindPlaneInsecureSkipVerify() bool {
	return c.bindplaneInsecureSkipVerify
}

type testResourceSet[T Resource] struct {
	resources map[string]T
}

func newTestResourceSet[T Resource]() testResourceSet[T] {
	return testResourceSet[T]{
		resources: map[string]T{},
	}
}

func (s *testResourceSet[T]) item(name string) (item T, err error) {
	n, v := SplitVersion(name)
	if v == VersionLatest {
		name = n
	}
	item = s.resources[name]
	return
}

func (s *testResourceSet[T]) add(item T) {
	s.resources[item.Name()] = item

	// also store with version
	s.resources[JoinVersion(item.Name(), item.Version())] = item
}

// addLatest should be called for the latest version after other versions are added
func (s *testResourceSet[T]) addLatest(item T) {
	item.SetLatest(true)
	s.add(item)
}

func (s *testResourceSet[T]) remove(name string) {
	delete(s.resources, name)
}

type testResourceStore struct {
	sources          testResourceSet[*Source]
	sourceTypes      testResourceSet[*SourceType]
	processors       testResourceSet[*Processor]
	processorTypes   testResourceSet[*ProcessorType]
	destinations     testResourceSet[*Destination]
	destinationTypes testResourceSet[*DestinationType]
}

func newTestResourceStore() *testResourceStore {
	return &testResourceStore{
		sources:          newTestResourceSet[*Source](),
		sourceTypes:      newTestResourceSet[*SourceType](),
		processors:       newTestResourceSet[*Processor](),
		processorTypes:   newTestResourceSet[*ProcessorType](),
		destinations:     newTestResourceSet[*Destination](),
		destinationTypes: newTestResourceSet[*DestinationType](),
	}
}

var _ ResourceStore = (*testResourceStore)(nil)

func (s *testResourceStore) Source(_ context.Context, name string) (*Source, error) {
	return s.sources.item(name)
}
func (s *testResourceStore) SourceType(_ context.Context, name string) (*SourceType, error) {
	return s.sourceTypes.item(name)
}
func (s *testResourceStore) Processor(_ context.Context, name string) (*Processor, error) {
	return s.processors.item(name)
}
func (s *testResourceStore) ProcessorType(_ context.Context, name string) (*ProcessorType, error) {
	return s.processorTypes.item(name)
}
func (s *testResourceStore) Destination(_ context.Context, name string) (*Destination, error) {
	return s.destinations.item(name)
}
func (s *testResourceStore) DestinationType(_ context.Context, name string) (*DestinationType, error) {
	return s.destinationTypes.item(name)
}

func TestParseConfiguration(t *testing.T) {
	path := filepath.Join("testfiles", "configuration-raw.yaml")
	bytes, err := os.ReadFile(path)
	require.NoError(t, err, "failed to read the testfile")
	var configuration Configuration
	err = yaml.Unmarshal(bytes, &configuration)
	require.NoError(t, err)
	require.Equal(t, "cabin-production-configuration", configuration.Metadata.Name)
	require.Equal(t, "receivers:", strings.Split(configuration.Spec.Raw, "\n")[0])
}

func TestEvalConfiguration(t *testing.T) {
	store := newTestResourceStore()
	config := newTestConfiguration()

	macos := testResource[*SourceType](t, "sourcetype-macos.yaml")
	store.sourceTypes.add(macos)

	cabin := testResource[*Destination](t, "destination-cabin.yaml")
	store.destinations.add(cabin)

	cabinType := testResource[*DestinationType](t, "destinationtype-cabin.yaml")
	store.destinationTypes.add(cabinType)

	configuration := testResource[*Configuration](t, "configuration-macos-sources.yaml")
	result, err := configuration.Render(context.TODO(), nil, config.BindPlaneURL(), config.BindPlaneInsecureSkipVerify(), store, GetOssOtelHeaders())
	require.NoError(t, err)

	expect := strings.TrimLeft(`
receivers:
    plugin/source0__journald:
        plugin:
            name: journald
    plugin/source0__macos:
        parameters:
            - name: enable_system_log
              value: false
            - name: system_log_path
              value: /var/log/system.log
            - name: enable_install_log
              value: true
            - name: install_log_path
              value: /var/log/install.log
            - name: start_at
              value: end
        plugin:
            name: macos
    plugin/source1__journald:
        plugin:
            name: journald
    plugin/source1__macos:
        parameters:
            - name: enable_system_log
              value: true
            - name: system_log_path
              value: /var/log/system.log
            - name: enable_install_log
              value: true
            - name: install_log_path
              value: /var/log/install.log
            - name: start_at
              value: end
        plugin:
            name: macos
processors:
    batch/cabin-production-logs: null
exporters:
    observiq/cabin-production-logs:
        endpoint: https://nozzle.app.observiq.com
        secret_key: 2c088c5e-2afc-483b-be52-e2b657fcff08
        timeout: 10s
service:
    pipelines:
        logs/source0__cabin-production-logs-0:
            receivers:
                - plugin/source0__macos
                - plugin/source0__journald
            processors:
                - batch/cabin-production-logs
            exporters:
                - observiq/cabin-production-logs
        logs/source1__cabin-production-logs-0:
            receivers:
                - plugin/source1__macos
                - plugin/source1__journald
            processors:
                - batch/cabin-production-logs
            exporters:
                - observiq/cabin-production-logs
`, "\n")

	require.Equal(t, expect, result)
}

func TestEvalConfiguration2(t *testing.T) {
	store := newTestResourceStore()
	config := newTestConfiguration()

	macos := testResource[*SourceType](t, "sourcetype-macos.yaml")
	store.sourceTypes.add(macos)

	googleCloudType := testResource[*DestinationType](t, "destinationtype-googlecloud.yaml")
	store.destinationTypes.add(googleCloudType)

	configuration := testResource[*Configuration](t, "configuration-macos-googlecloud.yaml")
	result, err := configuration.Render(context.TODO(), nil, config.BindPlaneURL(), config.BindPlaneInsecureSkipVerify(), store, GetOssOtelHeaders())
	require.NoError(t, err)

	expect := strings.TrimLeft(`
receivers:
    hostmetrics/source0:
        collection_interval: 1m
        scrapers:
            load: null
    hostmetrics/source1:
        collection_interval: 1m
        scrapers:
            load: null
    plugin/source0__journald:
        plugin:
            name: journald
    plugin/source0__macos:
        parameters:
            - name: enable_system_log
              value: false
            - name: system_log_path
              value: /var/log/system.log
            - name: enable_install_log
              value: true
            - name: install_log_path
              value: /var/log/install.log
            - name: start_at
              value: end
        plugin:
            name: macos
    plugin/source1__journald:
        plugin:
            name: journald
    plugin/source1__macos:
        parameters:
            - name: enable_system_log
              value: true
            - name: system_log_path
              value: /var/log/system.log
            - name: enable_install_log
              value: true
            - name: install_log_path
              value: /var/log/install.log
            - name: start_at
              value: end
        plugin:
            name: macos
processors:
    batch/destination0: null
exporters:
    googlecloud/destination0: null
service:
    pipelines:
        logs/source0__destination0-0:
            receivers:
                - plugin/source0__macos
                - plugin/source0__journald
            processors:
                - batch/destination0
            exporters:
                - googlecloud/destination0
        logs/source1__destination0-0:
            receivers:
                - plugin/source1__macos
                - plugin/source1__journald
            processors:
                - batch/destination0
            exporters:
                - googlecloud/destination0
        metrics/source0__destination0-0:
            receivers:
                - hostmetrics/source0
            processors:
                - batch/destination0
            exporters:
                - googlecloud/destination0
        metrics/source1__destination0-0:
            receivers:
                - hostmetrics/source1
            processors:
                - batch/destination0
            exporters:
                - googlecloud/destination0
`, "\n")

	require.Equal(t, expect, result)
}

func TestEvalConfiguration3(t *testing.T) {
	store := newTestResourceStore()
	config := newTestConfiguration()

	otlp := testResource[*SourceType](t, "sourcetype-otlp.yaml")
	store.sourceTypes.add(otlp)

	otlpDestinationType := testResource[*DestinationType](t, "destinationtype-otlp.yaml")
	store.destinationTypes.add(otlpDestinationType)

	configuration := testResource[*Configuration](t, "configuration-otlp.yaml")
	result, err := configuration.Render(context.TODO(), nil, config.BindPlaneURL(), config.BindPlaneInsecureSkipVerify(), store, GetOssOtelHeaders())
	require.NoError(t, err)

	expect := strings.TrimLeft(`
receivers:
    otlp/source0:
        protocols:
            grpc: null
            http: null
processors:
    batch/destination0: null
exporters:
    otlp/destination0:
        endpoint: otelcol:4317
service:
    pipelines:
        logs/source0__destination0-0:
            receivers:
                - otlp/source0
            processors:
                - batch/destination0
            exporters:
                - otlp/destination0
        metrics/source0__destination0-0:
            receivers:
                - otlp/source0
            processors:
                - batch/destination0
            exporters:
                - otlp/destination0
        traces/source0__destination0-0:
            receivers:
                - otlp/source0
            processors:
                - batch/destination0
            exporters:
                - otlp/destination0
`, "\n")

	require.Equal(t, expect, result)
}

func TestEvalConfiguration4(t *testing.T) {
	store := newTestResourceStore()
	config := newTestConfiguration()

	postgresql := testResource[*SourceType](t, "sourcetype-postgresql.yaml")
	store.sourceTypes.add(postgresql)

	googleCloudType := testResource[*DestinationType](t, "destinationtype-googlecloud.yaml")
	store.destinationTypes.add(googleCloudType)

	configuration := testResource[*Configuration](t, "configuration-postgresql-googlecloud.yaml")
	result, err := configuration.Render(context.TODO(), nil, config.BindPlaneURL(), config.BindPlaneInsecureSkipVerify(), store, GetOssOtelHeaders())
	require.NoError(t, err)

	expect := strings.TrimLeft(`
receivers:
    plugin/source0__postgresql:
        parameters:
            postgresql_log_path:
                - /var/log/postgresql/postgresql*.log
                - /var/lib/pgsql/data/log/postgresql*.log
                - /var/lib/pgsql/*/data/log/postgresql*.log
            start_at: end
        path: $OIQ_OTEL_COLLECTOR_HOME/plugins/postgresql_logs.yaml
processors:
    batch/destination0: null
exporters:
    googlecloud/destination0: null
service:
    pipelines:
        logs/source0__destination0-0:
            receivers:
                - plugin/source0__postgresql
            processors:
                - batch/destination0
            exporters:
                - googlecloud/destination0
`, "\n")

	require.Equal(t, expect, result)
}

func TestEvalConfiguration5(t *testing.T) {
	store := newTestResourceStore()
	config := newTestConfiguration()

	postgresql := testResource[*SourceType](t, "sourcetype-macos.yaml")
	store.sourceTypes.add(postgresql)

	googleCloudType := testResource[*DestinationType](t, "destinationtype-googlecloud.yaml")
	store.destinationTypes.add(googleCloudType)

	googleCloud := testResource[*Destination](t, "destination-googlecloud.yaml")
	store.destinations.add(googleCloud)

	resourceAttributeTransposerType := testResource[*ProcessorType](t, "processortype-resourceattributetransposer.yaml")
	store.processorTypes.add(resourceAttributeTransposerType)

	configuration := testResource[*Configuration](t, "configuration-macos-processors.yaml")
	result, err := configuration.Render(context.TODO(), nil, config.BindPlaneURL(), config.BindPlaneInsecureSkipVerify(), store, GetOssOtelHeaders())
	require.NoError(t, err)

	expect := strings.TrimLeft(`
receivers:
    hostmetrics/source0:
        collection_interval: 1m
        scrapers:
            load: null
    plugin/source0__journald:
        plugin:
            name: journald
    plugin/source0__macos:
        parameters:
            - name: enable_system_log
              value: false
            - name: system_log_path
              value: /var/log/system.log
            - name: enable_install_log
              value: true
            - name: install_log_path
              value: /var/log/install.log
            - name: start_at
              value: end
        plugin:
            name: macos
processors:
    batch/googlecloud: null
    resourceattributetransposer/source0__processor0:
        operations:
            - from: from.attribute
              to: to.attribute
    resourceattributetransposer/source0__processor1:
        operations:
            - from: from.attribute2
              to: to.attribute2
exporters:
    googlecloud/googlecloud: null
service:
    pipelines:
        logs/source0__googlecloud-0:
            receivers:
                - plugin/source0__macos
                - plugin/source0__journald
            processors:
                - resourceattributetransposer/source0__processor0
                - resourceattributetransposer/source0__processor1
                - batch/googlecloud
            exporters:
                - googlecloud/googlecloud
        metrics/source0__googlecloud-0:
            receivers:
                - hostmetrics/source0
            processors:
                - resourceattributetransposer/source0__processor0
                - resourceattributetransposer/source0__processor1
                - batch/googlecloud
            exporters:
                - googlecloud/googlecloud
`, "\n")

	require.Equal(t, expect, result)
}

func TestEvalConfigurationDestinationProcessors(t *testing.T) {
	store := newTestResourceStore()
	config := newTestConfiguration()

	postgresql := testResource[*SourceType](t, "sourcetype-macos.yaml")
	store.sourceTypes.add(postgresql)

	googleCloudType := testResource[*DestinationType](t, "destinationtype-googlecloud.yaml")
	store.destinationTypes.add(googleCloudType)

	googleCloud := testResource[*Destination](t, "destination-googlecloud.yaml")
	store.destinations.add(googleCloud)

	resourceAttributeTransposerType := testResource[*ProcessorType](t, "processortype-resourceattributetransposer.yaml")
	store.processorTypes.add(resourceAttributeTransposerType)

	configuration := testResource[*Configuration](t, "configuration-macos-destination-processors.yaml")
	result, err := configuration.Render(context.TODO(), nil, config.BindPlaneURL(), config.BindPlaneInsecureSkipVerify(), store, GetOssOtelHeaders())
	require.NoError(t, err)

	expect := strings.TrimLeft(`
receivers:
    hostmetrics/source0:
        collection_interval: 1m
        scrapers:
            load: null
    plugin/source0__journald:
        plugin:
            name: journald
    plugin/source0__macos:
        parameters:
            - name: enable_system_log
              value: false
            - name: system_log_path
              value: /var/log/system.log
            - name: enable_install_log
              value: true
            - name: install_log_path
              value: /var/log/install.log
            - name: start_at
              value: end
        plugin:
            name: macos
processors:
    batch/googlecloud: null
    resourceattributetransposer/googlecloud-0__processor0:
        operations:
            - from: from.attribute3
              to: to.attribute3
    resourceattributetransposer/googlecloud-0__processor1:
        operations:
            - from: from.attribute4
              to: to.attribute4
    resourceattributetransposer/source0__processor0:
        operations:
            - from: from.attribute
              to: to.attribute
    resourceattributetransposer/source0__processor1:
        operations:
            - from: from.attribute2
              to: to.attribute2
exporters:
    googlecloud/googlecloud: null
service:
    pipelines:
        logs/source0__googlecloud-0:
            receivers:
                - plugin/source0__macos
                - plugin/source0__journald
            processors:
                - resourceattributetransposer/source0__processor0
                - resourceattributetransposer/source0__processor1
                - resourceattributetransposer/googlecloud-0__processor0
                - resourceattributetransposer/googlecloud-0__processor1
                - batch/googlecloud
            exporters:
                - googlecloud/googlecloud
        metrics/source0__googlecloud-0:
            receivers:
                - hostmetrics/source0
            processors:
                - resourceattributetransposer/source0__processor0
                - resourceattributetransposer/source0__processor1
                - resourceattributetransposer/googlecloud-0__processor0
                - resourceattributetransposer/googlecloud-0__processor1
                - batch/googlecloud
            exporters:
                - googlecloud/googlecloud
`, "\n")

	require.Equal(t, expect, result)
}

func TestEvalConfigurationDestinationProcessorsWithMeasurements(t *testing.T) {
	store := newTestResourceStore()
	config := newTestConfiguration()

	postgresql := testResource[*SourceType](t, "sourcetype-macos.yaml")
	store.sourceTypes.add(postgresql)

	googleCloudType := testResource[*DestinationType](t, "destinationtype-googlecloud.yaml")
	store.destinationTypes.add(googleCloudType)

	googleCloud := testResource[*Destination](t, "destination-googlecloud.yaml")
	store.destinations.add(googleCloud)

	resourceAttributeTransposerType := testResource[*ProcessorType](t, "processortype-resourceattributetransposer.yaml")
	store.processorTypes.add(resourceAttributeTransposerType)

	agent := &Agent{
		ID:      "01ARZ3NDEKTSV4RRFFQ69G5FAV",
		Version: v1_9_2.String(),
	}

	configuration := testResource[*Configuration](t, "configuration-macos-destination-processors.yaml")
	result, err := configuration.Render(context.TODO(), agent, config.BindPlaneURL(), config.BindPlaneInsecureSkipVerify(), store, GetOssOtelHeaders())
	require.NoError(t, err)

	expect := strings.TrimLeft(`
receivers:
    hostmetrics/source0:
        collection_interval: 1m
        scrapers:
            load: null
    plugin/source0__journald:
        plugin:
            name: journald
    plugin/source0__macos:
        parameters:
            - name: enable_system_log
              value: false
            - name: system_log_path
              value: /var/log/system.log
            - name: enable_install_log
              value: true
            - name: install_log_path
              value: /var/log/install.log
            - name: start_at
              value: end
        plugin:
            name: macos
    prometheus/_agent_metrics:
        config:
            scrape_configs:
                - job_name: observiq-otel-collector
                  metric_relabel_configs:
                    - action: keep
                      regex: otelcol_processor_throughputmeasurement_.*
                      source_labels:
                        - __name__
                  scrape_interval: 10s
                  static_configs:
                    - labels:
                        agent: 01ARZ3NDEKTSV4RRFFQ69G5FAV
                        configuration: macos-xy
                      targets:
                        - 0.0.0.0:8888
processors:
    batch/_agent_metrics: null
    batch/googlecloud: null
    resourceattributetransposer/googlecloud-0__processor0:
        operations:
            - from: from.attribute3
              to: to.attribute3
    resourceattributetransposer/googlecloud-0__processor1:
        operations:
            - from: from.attribute4
              to: to.attribute4
    resourceattributetransposer/source0__processor0:
        operations:
            - from: from.attribute
              to: to.attribute
    resourceattributetransposer/source0__processor1:
        operations:
            - from: from.attribute2
              to: to.attribute2
    snapshotprocessor: null
    snapshotprocessor/_d0_googlecloud-0: null
    snapshotprocessor/_s0_source0: null
    throughputmeasurement/_d0_logs_googlecloud-0:
        enabled: true
        sampling_ratio: 1
    throughputmeasurement/_d0_metrics_googlecloud-0:
        enabled: true
        sampling_ratio: 1
    throughputmeasurement/_d1_logs_googlecloud-0:
        enabled: true
        sampling_ratio: 1
    throughputmeasurement/_d1_metrics_googlecloud-0:
        enabled: true
        sampling_ratio: 1
    throughputmeasurement/_s0_logs_source0:
        enabled: true
        sampling_ratio: 1
    throughputmeasurement/_s0_metrics_source0:
        enabled: true
        sampling_ratio: 1
    throughputmeasurement/_s1_logs_source0:
        enabled: true
        sampling_ratio: 1
    throughputmeasurement/_s1_metrics_source0:
        enabled: true
        sampling_ratio: 1
exporters:
    googlecloud/googlecloud: null
    otlphttp/_agent_metrics:
        endpoint: /v1/otlphttp
        headers: {}
        retry_on_failure:
            enabled: true
            initial_interval: 5s
            max_elapsed_time: 30s
            max_interval: 5s
        sending_queue:
            enabled: true
            num_consumers: 1
            queue_size: 60
service:
    pipelines:
        logs/source0__googlecloud-0:
            receivers:
                - plugin/source0__macos
                - plugin/source0__journald
            processors:
                - throughputmeasurement/_s0_logs_source0
                - snapshotprocessor/_s0_source0
                - resourceattributetransposer/source0__processor0
                - resourceattributetransposer/source0__processor1
                - throughputmeasurement/_s1_logs_source0
                - throughputmeasurement/_d0_logs_googlecloud-0
                - snapshotprocessor/_d0_googlecloud-0
                - resourceattributetransposer/googlecloud-0__processor0
                - resourceattributetransposer/googlecloud-0__processor1
                - throughputmeasurement/_d1_logs_googlecloud-0
                - batch/googlecloud
                - snapshotprocessor
            exporters:
                - googlecloud/googlecloud
        metrics/_agent_metrics:
            receivers:
                - prometheus/_agent_metrics
            processors:
                - batch/_agent_metrics
            exporters:
                - otlphttp/_agent_metrics
        metrics/source0__googlecloud-0:
            receivers:
                - hostmetrics/source0
            processors:
                - throughputmeasurement/_s0_metrics_source0
                - snapshotprocessor/_s0_source0
                - resourceattributetransposer/source0__processor0
                - resourceattributetransposer/source0__processor1
                - throughputmeasurement/_s1_metrics_source0
                - throughputmeasurement/_d0_metrics_googlecloud-0
                - snapshotprocessor/_d0_googlecloud-0
                - resourceattributetransposer/googlecloud-0__processor0
                - resourceattributetransposer/googlecloud-0__processor1
                - throughputmeasurement/_d1_metrics_googlecloud-0
                - batch/googlecloud
                - snapshotprocessor
            exporters:
                - googlecloud/googlecloud
`, "\n")

	require.Equal(t, expect, result)
}

func TestEvalConfigurationDestinationProcessorsWithMeasurementsMTLS(t *testing.T) {
	store := newTestResourceStore()
	config := newTestConfiguration()

	postgresql := testResource[*SourceType](t, "sourcetype-macos.yaml")
	store.sourceTypes.add(postgresql)

	googleCloudType := testResource[*DestinationType](t, "destinationtype-googlecloud.yaml")
	store.destinationTypes.add(googleCloudType)

	googleCloud := testResource[*Destination](t, "destination-googlecloud.yaml")
	store.destinations.add(googleCloud)

	resourceAttributeTransposerType := testResource[*ProcessorType](t, "processortype-resourceattributetransposer.yaml")
	store.processorTypes.add(resourceAttributeTransposerType)

	agent := &Agent{
		ID:      "01ARZ3NDEKTSV4RRFFQ69G5FAV",
		Version: v1_9_2.String(),
		TLS: &ManagerTLS{
			InsecureSkipVerify: true,
			CAFile:             strp("/path/to/ca"),
			CertFile:           strp("/path/to/cert"),
			KeyFile:            strp("/path/to/key"),
		},
	}

	configuration := testResource[*Configuration](t, "configuration-macos-destination-processors.yaml")
	result, err := configuration.Render(context.TODO(), agent, config.bindplaneURL, config.bindplaneInsecureSkipVerify, store, GetOssOtelHeaders())
	require.NoError(t, err)

	expect := strings.TrimLeft(`
receivers:
    hostmetrics/source0:
        collection_interval: 1m
        scrapers:
            load: null
    plugin/source0__journald:
        plugin:
            name: journald
    plugin/source0__macos:
        parameters:
            - name: enable_system_log
              value: false
            - name: system_log_path
              value: /var/log/system.log
            - name: enable_install_log
              value: true
            - name: install_log_path
              value: /var/log/install.log
            - name: start_at
              value: end
        plugin:
            name: macos
    prometheus/_agent_metrics:
        config:
            scrape_configs:
                - job_name: observiq-otel-collector
                  metric_relabel_configs:
                    - action: keep
                      regex: otelcol_processor_throughputmeasurement_.*
                      source_labels:
                        - __name__
                  scrape_interval: 10s
                  static_configs:
                    - labels:
                        agent: 01ARZ3NDEKTSV4RRFFQ69G5FAV
                        configuration: macos-xy
                      targets:
                        - 0.0.0.0:8888
processors:
    batch/_agent_metrics: null
    batch/googlecloud: null
    resourceattributetransposer/googlecloud-0__processor0:
        operations:
            - from: from.attribute3
              to: to.attribute3
    resourceattributetransposer/googlecloud-0__processor1:
        operations:
            - from: from.attribute4
              to: to.attribute4
    resourceattributetransposer/source0__processor0:
        operations:
            - from: from.attribute
              to: to.attribute
    resourceattributetransposer/source0__processor1:
        operations:
            - from: from.attribute2
              to: to.attribute2
    snapshotprocessor: null
    snapshotprocessor/_d0_googlecloud-0: null
    snapshotprocessor/_s0_source0: null
    throughputmeasurement/_d0_logs_googlecloud-0:
        enabled: true
        sampling_ratio: 1
    throughputmeasurement/_d0_metrics_googlecloud-0:
        enabled: true
        sampling_ratio: 1
    throughputmeasurement/_d1_logs_googlecloud-0:
        enabled: true
        sampling_ratio: 1
    throughputmeasurement/_d1_metrics_googlecloud-0:
        enabled: true
        sampling_ratio: 1
    throughputmeasurement/_s0_logs_source0:
        enabled: true
        sampling_ratio: 1
    throughputmeasurement/_s0_metrics_source0:
        enabled: true
        sampling_ratio: 1
    throughputmeasurement/_s1_logs_source0:
        enabled: true
        sampling_ratio: 1
    throughputmeasurement/_s1_metrics_source0:
        enabled: true
        sampling_ratio: 1
exporters:
    googlecloud/googlecloud: null
    otlphttp/_agent_metrics:
        endpoint: /v1/otlphttp
        headers: {}
        retry_on_failure:
            enabled: true
            initial_interval: 5s
            max_elapsed_time: 30s
            max_interval: 5s
        sending_queue:
            enabled: true
            num_consumers: 1
            queue_size: 60
        tls:
            ca_file: /path/to/ca
            cert_file: /path/to/cert
            insecure_skip_verify: true
            key_file: /path/to/key
service:
    pipelines:
        logs/source0__googlecloud-0:
            receivers:
                - plugin/source0__macos
                - plugin/source0__journald
            processors:
                - throughputmeasurement/_s0_logs_source0
                - snapshotprocessor/_s0_source0
                - resourceattributetransposer/source0__processor0
                - resourceattributetransposer/source0__processor1
                - throughputmeasurement/_s1_logs_source0
                - throughputmeasurement/_d0_logs_googlecloud-0
                - snapshotprocessor/_d0_googlecloud-0
                - resourceattributetransposer/googlecloud-0__processor0
                - resourceattributetransposer/googlecloud-0__processor1
                - throughputmeasurement/_d1_logs_googlecloud-0
                - batch/googlecloud
                - snapshotprocessor
            exporters:
                - googlecloud/googlecloud
        metrics/_agent_metrics:
            receivers:
                - prometheus/_agent_metrics
            processors:
                - batch/_agent_metrics
            exporters:
                - otlphttp/_agent_metrics
        metrics/source0__googlecloud-0:
            receivers:
                - hostmetrics/source0
            processors:
                - throughputmeasurement/_s0_metrics_source0
                - snapshotprocessor/_s0_source0
                - resourceattributetransposer/source0__processor0
                - resourceattributetransposer/source0__processor1
                - throughputmeasurement/_s1_metrics_source0
                - throughputmeasurement/_d0_metrics_googlecloud-0
                - snapshotprocessor/_d0_googlecloud-0
                - resourceattributetransposer/googlecloud-0__processor0
                - resourceattributetransposer/googlecloud-0__processor1
                - throughputmeasurement/_d1_metrics_googlecloud-0
                - batch/googlecloud
                - snapshotprocessor
            exporters:
                - googlecloud/googlecloud
`, "\n")

	require.Equal(t, expect, result)
}

func TestEvalConfigurationDestinationProcessorsWithMeasurementsMTLSInsecureOverride(t *testing.T) {
	store := newTestResourceStore()
	config := newTestConfiguration()

	// BindplaneInsecureSkipVerify should override the InsecureSkipVerify in the manager.yaml of the agent
	config.bindplaneInsecureSkipVerify = true

	postgresql := testResource[*SourceType](t, "sourcetype-macos.yaml")
	store.sourceTypes.add(postgresql)

	googleCloudType := testResource[*DestinationType](t, "destinationtype-googlecloud.yaml")
	store.destinationTypes.add(googleCloudType)

	googleCloud := testResource[*Destination](t, "destination-googlecloud.yaml")
	store.destinations.add(googleCloud)

	resourceAttributeTransposerType := testResource[*ProcessorType](t, "processortype-resourceattributetransposer.yaml")
	store.processorTypes.add(resourceAttributeTransposerType)

	agent := &Agent{
		ID:      "01ARZ3NDEKTSV4RRFFQ69G5FAV",
		Version: v1_9_2.String(),
		TLS: &ManagerTLS{
			InsecureSkipVerify: false,
			CAFile:             strp("/path/to/ca"),
			CertFile:           strp("/path/to/cert"),
			KeyFile:            strp("/path/to/key"),
		},
	}

	configuration := testResource[*Configuration](t, "configuration-macos-destination-processors.yaml")
	result, err := configuration.Render(context.TODO(), agent, config.bindplaneURL, config.bindplaneInsecureSkipVerify, store, GetOssOtelHeaders())
	require.NoError(t, err)

	expect := strings.TrimLeft(`
receivers:
    hostmetrics/source0:
        collection_interval: 1m
        scrapers:
            load: null
    plugin/source0__journald:
        plugin:
            name: journald
    plugin/source0__macos:
        parameters:
            - name: enable_system_log
              value: false
            - name: system_log_path
              value: /var/log/system.log
            - name: enable_install_log
              value: true
            - name: install_log_path
              value: /var/log/install.log
            - name: start_at
              value: end
        plugin:
            name: macos
    prometheus/_agent_metrics:
        config:
            scrape_configs:
                - job_name: observiq-otel-collector
                  metric_relabel_configs:
                    - action: keep
                      regex: otelcol_processor_throughputmeasurement_.*
                      source_labels:
                        - __name__
                  scrape_interval: 10s
                  static_configs:
                    - labels:
                        agent: 01ARZ3NDEKTSV4RRFFQ69G5FAV
                        configuration: macos-xy
                      targets:
                        - 0.0.0.0:8888
processors:
    batch/_agent_metrics: null
    batch/googlecloud: null
    resourceattributetransposer/googlecloud-0__processor0:
        operations:
            - from: from.attribute3
              to: to.attribute3
    resourceattributetransposer/googlecloud-0__processor1:
        operations:
            - from: from.attribute4
              to: to.attribute4
    resourceattributetransposer/source0__processor0:
        operations:
            - from: from.attribute
              to: to.attribute
    resourceattributetransposer/source0__processor1:
        operations:
            - from: from.attribute2
              to: to.attribute2
    snapshotprocessor: null
    snapshotprocessor/_d0_googlecloud-0: null
    snapshotprocessor/_s0_source0: null
    throughputmeasurement/_d0_logs_googlecloud-0:
        enabled: true
        sampling_ratio: 1
    throughputmeasurement/_d0_metrics_googlecloud-0:
        enabled: true
        sampling_ratio: 1
    throughputmeasurement/_d1_logs_googlecloud-0:
        enabled: true
        sampling_ratio: 1
    throughputmeasurement/_d1_metrics_googlecloud-0:
        enabled: true
        sampling_ratio: 1
    throughputmeasurement/_s0_logs_source0:
        enabled: true
        sampling_ratio: 1
    throughputmeasurement/_s0_metrics_source0:
        enabled: true
        sampling_ratio: 1
    throughputmeasurement/_s1_logs_source0:
        enabled: true
        sampling_ratio: 1
    throughputmeasurement/_s1_metrics_source0:
        enabled: true
        sampling_ratio: 1
exporters:
    googlecloud/googlecloud: null
    otlphttp/_agent_metrics:
        endpoint: /v1/otlphttp
        headers: {}
        retry_on_failure:
            enabled: true
            initial_interval: 5s
            max_elapsed_time: 30s
            max_interval: 5s
        sending_queue:
            enabled: true
            num_consumers: 1
            queue_size: 60
        tls:
            ca_file: /path/to/ca
            cert_file: /path/to/cert
            insecure_skip_verify: true
            key_file: /path/to/key
service:
    pipelines:
        logs/source0__googlecloud-0:
            receivers:
                - plugin/source0__macos
                - plugin/source0__journald
            processors:
                - throughputmeasurement/_s0_logs_source0
                - snapshotprocessor/_s0_source0
                - resourceattributetransposer/source0__processor0
                - resourceattributetransposer/source0__processor1
                - throughputmeasurement/_s1_logs_source0
                - throughputmeasurement/_d0_logs_googlecloud-0
                - snapshotprocessor/_d0_googlecloud-0
                - resourceattributetransposer/googlecloud-0__processor0
                - resourceattributetransposer/googlecloud-0__processor1
                - throughputmeasurement/_d1_logs_googlecloud-0
                - batch/googlecloud
                - snapshotprocessor
            exporters:
                - googlecloud/googlecloud
        metrics/_agent_metrics:
            receivers:
                - prometheus/_agent_metrics
            processors:
                - batch/_agent_metrics
            exporters:
                - otlphttp/_agent_metrics
        metrics/source0__googlecloud-0:
            receivers:
                - hostmetrics/source0
            processors:
                - throughputmeasurement/_s0_metrics_source0
                - snapshotprocessor/_s0_source0
                - resourceattributetransposer/source0__processor0
                - resourceattributetransposer/source0__processor1
                - throughputmeasurement/_s1_metrics_source0
                - throughputmeasurement/_d0_metrics_googlecloud-0
                - snapshotprocessor/_d0_googlecloud-0
                - resourceattributetransposer/googlecloud-0__processor0
                - resourceattributetransposer/googlecloud-0__processor1
                - throughputmeasurement/_d1_metrics_googlecloud-0
                - batch/googlecloud
                - snapshotprocessor
            exporters:
                - googlecloud/googlecloud
`, "\n")

	require.Equal(t, expect, result)
}

func TestEvalConfigurationDestinationProcessorsWithMeasurementsTLSSkipVerify(t *testing.T) {
	store := newTestResourceStore()
	config := newTestConfiguration()

	// BindplaneInsecureSkipVerify should propagate to measurements configuration in agent yaml
	config.bindplaneInsecureSkipVerify = true

	postgresql := testResource[*SourceType](t, "sourcetype-macos.yaml")
	store.sourceTypes.add(postgresql)

	googleCloudType := testResource[*DestinationType](t, "destinationtype-googlecloud.yaml")
	store.destinationTypes.add(googleCloudType)

	googleCloud := testResource[*Destination](t, "destination-googlecloud.yaml")
	store.destinations.add(googleCloud)

	resourceAttributeTransposerType := testResource[*ProcessorType](t, "processortype-resourceattributetransposer.yaml")
	store.processorTypes.add(resourceAttributeTransposerType)

	agent := &Agent{
		ID:      "01ARZ3NDEKTSV4RRFFQ69G5FAV",
		Version: v1_9_2.String(),
	}

	configuration := testResource[*Configuration](t, "configuration-macos-destination-processors.yaml")
	result, err := configuration.Render(context.TODO(), agent, config.bindplaneURL, config.bindplaneInsecureSkipVerify, store, GetOssOtelHeaders())
	require.NoError(t, err)

	expect := strings.TrimLeft(`
receivers:
    hostmetrics/source0:
        collection_interval: 1m
        scrapers:
            load: null
    plugin/source0__journald:
        plugin:
            name: journald
    plugin/source0__macos:
        parameters:
            - name: enable_system_log
              value: false
            - name: system_log_path
              value: /var/log/system.log
            - name: enable_install_log
              value: true
            - name: install_log_path
              value: /var/log/install.log
            - name: start_at
              value: end
        plugin:
            name: macos
    prometheus/_agent_metrics:
        config:
            scrape_configs:
                - job_name: observiq-otel-collector
                  metric_relabel_configs:
                    - action: keep
                      regex: otelcol_processor_throughputmeasurement_.*
                      source_labels:
                        - __name__
                  scrape_interval: 10s
                  static_configs:
                    - labels:
                        agent: 01ARZ3NDEKTSV4RRFFQ69G5FAV
                        configuration: macos-xy
                      targets:
                        - 0.0.0.0:8888
processors:
    batch/_agent_metrics: null
    batch/googlecloud: null
    resourceattributetransposer/googlecloud-0__processor0:
        operations:
            - from: from.attribute3
              to: to.attribute3
    resourceattributetransposer/googlecloud-0__processor1:
        operations:
            - from: from.attribute4
              to: to.attribute4
    resourceattributetransposer/source0__processor0:
        operations:
            - from: from.attribute
              to: to.attribute
    resourceattributetransposer/source0__processor1:
        operations:
            - from: from.attribute2
              to: to.attribute2
    snapshotprocessor: null
    snapshotprocessor/_d0_googlecloud-0: null
    snapshotprocessor/_s0_source0: null
    throughputmeasurement/_d0_logs_googlecloud-0:
        enabled: true
        sampling_ratio: 1
    throughputmeasurement/_d0_metrics_googlecloud-0:
        enabled: true
        sampling_ratio: 1
    throughputmeasurement/_d1_logs_googlecloud-0:
        enabled: true
        sampling_ratio: 1
    throughputmeasurement/_d1_metrics_googlecloud-0:
        enabled: true
        sampling_ratio: 1
    throughputmeasurement/_s0_logs_source0:
        enabled: true
        sampling_ratio: 1
    throughputmeasurement/_s0_metrics_source0:
        enabled: true
        sampling_ratio: 1
    throughputmeasurement/_s1_logs_source0:
        enabled: true
        sampling_ratio: 1
    throughputmeasurement/_s1_metrics_source0:
        enabled: true
        sampling_ratio: 1
exporters:
    googlecloud/googlecloud: null
    otlphttp/_agent_metrics:
        endpoint: /v1/otlphttp
        headers: {}
        retry_on_failure:
            enabled: true
            initial_interval: 5s
            max_elapsed_time: 30s
            max_interval: 5s
        sending_queue:
            enabled: true
            num_consumers: 1
            queue_size: 60
        tls:
            insecure_skip_verify: true
service:
    pipelines:
        logs/source0__googlecloud-0:
            receivers:
                - plugin/source0__macos
                - plugin/source0__journald
            processors:
                - throughputmeasurement/_s0_logs_source0
                - snapshotprocessor/_s0_source0
                - resourceattributetransposer/source0__processor0
                - resourceattributetransposer/source0__processor1
                - throughputmeasurement/_s1_logs_source0
                - throughputmeasurement/_d0_logs_googlecloud-0
                - snapshotprocessor/_d0_googlecloud-0
                - resourceattributetransposer/googlecloud-0__processor0
                - resourceattributetransposer/googlecloud-0__processor1
                - throughputmeasurement/_d1_logs_googlecloud-0
                - batch/googlecloud
                - snapshotprocessor
            exporters:
                - googlecloud/googlecloud
        metrics/_agent_metrics:
            receivers:
                - prometheus/_agent_metrics
            processors:
                - batch/_agent_metrics
            exporters:
                - otlphttp/_agent_metrics
        metrics/source0__googlecloud-0:
            receivers:
                - hostmetrics/source0
            processors:
                - throughputmeasurement/_s0_metrics_source0
                - snapshotprocessor/_s0_source0
                - resourceattributetransposer/source0__processor0
                - resourceattributetransposer/source0__processor1
                - throughputmeasurement/_s1_metrics_source0
                - throughputmeasurement/_d0_metrics_googlecloud-0
                - snapshotprocessor/_d0_googlecloud-0
                - resourceattributetransposer/googlecloud-0__processor0
                - resourceattributetransposer/googlecloud-0__processor1
                - throughputmeasurement/_d1_metrics_googlecloud-0
                - batch/googlecloud
                - snapshotprocessor
            exporters:
                - googlecloud/googlecloud
`, "\n")

	require.Equal(t, expect, result)
}

func TestEvalConfigurationMultiDestination(t *testing.T) {
	store := newTestResourceStore()
	config := newTestConfiguration()

	postgresql := testResource[*SourceType](t, "sourcetype-macos.yaml")
	store.sourceTypes.add(postgresql)

	googleCloudType := testResource[*DestinationType](t, "destinationtype-googlecloud.yaml")
	store.destinationTypes.add(googleCloudType)

	cabinType := testResource[*DestinationType](t, "destinationtype-cabin.yaml")
	store.destinationTypes.add(cabinType)

	googleCloud := testResource[*Destination](t, "destination-googlecloud.yaml")
	store.destinations.add(googleCloud)

	cabin := testResource[*Destination](t, "destination-cabin.yaml")
	store.destinations.add(cabin)

	resourceAttributeTransposerType := testResource[*ProcessorType](t, "processortype-resourceattributetransposer.yaml")
	store.processorTypes.add(resourceAttributeTransposerType)
	agent := &Agent{
		ID:      "01ARZ3NDEKTSV4RRFFQ69G5FAV",
		Version: v1_9_2.String(),
	}

	configuration := testResource[*Configuration](t, "configuration-macos-multi-destination.yaml")
	result, err := configuration.Render(context.TODO(), agent, config.BindPlaneURL(), config.BindPlaneInsecureSkipVerify(), store, GetOssOtelHeaders())
	require.NoError(t, err)

	expect := strings.TrimLeft(`
receivers:
    hostmetrics/source0:
        collection_interval: 1m
        scrapers:
            load: null
    plugin/source0__journald:
        plugin:
            name: journald
    plugin/source0__macos:
        parameters:
            - name: enable_system_log
              value: false
            - name: system_log_path
              value: /var/log/system.log
            - name: enable_install_log
              value: true
            - name: install_log_path
              value: /var/log/install.log
            - name: start_at
              value: end
        plugin:
            name: macos
    prometheus/_agent_metrics:
        config:
            scrape_configs:
                - job_name: observiq-otel-collector
                  metric_relabel_configs:
                    - action: keep
                      regex: otelcol_processor_throughputmeasurement_.*
                      source_labels:
                        - __name__
                  scrape_interval: 10s
                  static_configs:
                    - labels:
                        agent: 01ARZ3NDEKTSV4RRFFQ69G5FAV
                        configuration: macos-xy
                      targets:
                        - 0.0.0.0:8888
processors:
    batch/_agent_metrics: null
    batch/cabin-production-logs: null
    batch/googlecloud: null
    resourceattributetransposer/source0__processor0:
        operations:
            - from: from.attribute
              to: to.attribute
    resourceattributetransposer/source0__processor1:
        operations:
            - from: from.attribute2
              to: to.attribute2
    snapshotprocessor: null
    snapshotprocessor/_d0_cabin-production-logs-1: null
    snapshotprocessor/_d0_googlecloud-0: null
    snapshotprocessor/_s0_source0: null
    throughputmeasurement/_d0_logs_cabin-production-logs-1:
        enabled: true
        sampling_ratio: 1
    throughputmeasurement/_d0_logs_googlecloud-0:
        enabled: true
        sampling_ratio: 1
    throughputmeasurement/_d0_metrics_googlecloud-0:
        enabled: true
        sampling_ratio: 1
    throughputmeasurement/_d1_logs_cabin-production-logs-1:
        enabled: true
        sampling_ratio: 1
    throughputmeasurement/_d1_logs_googlecloud-0:
        enabled: true
        sampling_ratio: 1
    throughputmeasurement/_d1_metrics_googlecloud-0:
        enabled: true
        sampling_ratio: 1
    throughputmeasurement/_s0_logs_source0:
        enabled: true
        sampling_ratio: 1
    throughputmeasurement/_s0_metrics_source0:
        enabled: true
        sampling_ratio: 1
    throughputmeasurement/_s1_logs_source0:
        enabled: true
        sampling_ratio: 1
    throughputmeasurement/_s1_metrics_source0:
        enabled: true
        sampling_ratio: 1
exporters:
    googlecloud/googlecloud: null
    observiq/cabin-production-logs:
        endpoint: https://nozzle.app.observiq.com
        secret_key: 2c088c5e-2afc-483b-be52-e2b657fcff08
        timeout: 10s
    otlphttp/_agent_metrics:
        endpoint: /v1/otlphttp
        headers: {}
        retry_on_failure:
            enabled: true
            initial_interval: 5s
            max_elapsed_time: 30s
            max_interval: 5s
        sending_queue:
            enabled: true
            num_consumers: 1
            queue_size: 60
service:
    pipelines:
        logs/source0__cabin-production-logs-1:
            receivers:
                - plugin/source0__macos
                - plugin/source0__journald
            processors:
                - throughputmeasurement/_s0_logs_source0
                - snapshotprocessor/_s0_source0
                - resourceattributetransposer/source0__processor0
                - resourceattributetransposer/source0__processor1
                - throughputmeasurement/_s1_logs_source0
                - throughputmeasurement/_d0_logs_cabin-production-logs-1
                - snapshotprocessor/_d0_cabin-production-logs-1
                - throughputmeasurement/_d1_logs_cabin-production-logs-1
                - batch/cabin-production-logs
                - snapshotprocessor
            exporters:
                - observiq/cabin-production-logs
        logs/source0__googlecloud-0:
            receivers:
                - plugin/source0__macos
                - plugin/source0__journald
            processors:
                - snapshotprocessor/_s0_source0
                - resourceattributetransposer/source0__processor0
                - resourceattributetransposer/source0__processor1
                - throughputmeasurement/_d0_logs_googlecloud-0
                - snapshotprocessor/_d0_googlecloud-0
                - throughputmeasurement/_d1_logs_googlecloud-0
                - batch/googlecloud
                - snapshotprocessor
            exporters:
                - googlecloud/googlecloud
        metrics/_agent_metrics:
            receivers:
                - prometheus/_agent_metrics
            processors:
                - batch/_agent_metrics
            exporters:
                - otlphttp/_agent_metrics
        metrics/source0__googlecloud-0:
            receivers:
                - hostmetrics/source0
            processors:
                - throughputmeasurement/_s0_metrics_source0
                - snapshotprocessor/_s0_source0
                - resourceattributetransposer/source0__processor0
                - resourceattributetransposer/source0__processor1
                - throughputmeasurement/_s1_metrics_source0
                - throughputmeasurement/_d0_metrics_googlecloud-0
                - snapshotprocessor/_d0_googlecloud-0
                - throughputmeasurement/_d1_metrics_googlecloud-0
                - batch/googlecloud
                - snapshotprocessor
            exporters:
                - googlecloud/googlecloud
`, "\n")

	require.Equal(t, expect, result)
}

func TestEvalConfigurationSameDestination(t *testing.T) {
	store := newTestResourceStore()
	config := newTestConfiguration()

	postgresql := testResource[*SourceType](t, "sourcetype-macos.yaml")
	store.sourceTypes.add(postgresql)

	googleCloudType := testResource[*DestinationType](t, "destinationtype-googlecloud.yaml")
	store.destinationTypes.add(googleCloudType)

	googleCloud := testResource[*Destination](t, "destination-googlecloud.yaml")
	store.destinations.add(googleCloud)

	resourceAttributeTransposerType := testResource[*ProcessorType](t, "processortype-resourceattributetransposer.yaml")
	store.processorTypes.add(resourceAttributeTransposerType)
	agent := &Agent{
		ID:      "01ARZ3NDEKTSV4RRFFQ69G5FAV",
		Version: v1_9_2.String(),
	}

	configuration := testResource[*Configuration](t, "configuration-macos-same-destination.yaml")
	result, err := configuration.Render(context.TODO(), agent, config.BindPlaneURL(), config.BindPlaneInsecureSkipVerify(), store, GetOssOtelHeaders())
	require.NoError(t, err)

	expect := strings.TrimLeft(`receivers:
    hostmetrics/source0:
        collection_interval: 1m
        scrapers:
            load: null
    plugin/source0__journald:
        plugin:
            name: journald
    plugin/source0__macos:
        parameters:
            - name: enable_system_log
              value: false
            - name: system_log_path
              value: /var/log/system.log
            - name: enable_install_log
              value: true
            - name: install_log_path
              value: /var/log/install.log
            - name: start_at
              value: end
        plugin:
            name: macos
    prometheus/_agent_metrics:
        config:
            scrape_configs:
                - job_name: observiq-otel-collector
                  metric_relabel_configs:
                    - action: keep
                      regex: otelcol_processor_throughputmeasurement_.*
                      source_labels:
                        - __name__
                  scrape_interval: 10s
                  static_configs:
                    - labels:
                        agent: 01ARZ3NDEKTSV4RRFFQ69G5FAV
                        configuration: macos-xy
                      targets:
                        - 0.0.0.0:8888
processors:
    batch/_agent_metrics: null
    batch/googlecloud: null
    resourceattributetransposer/source0__processor0:
        operations:
            - from: from.attribute
              to: to.attribute
    resourceattributetransposer/source0__processor1:
        operations:
            - from: from.attribute2
              to: to.attribute2
    snapshotprocessor: null
    snapshotprocessor/_d0_googlecloud-0: null
    snapshotprocessor/_d0_googlecloud-1: null
    snapshotprocessor/_s0_source0: null
    throughputmeasurement/_d0_logs_googlecloud-0:
        enabled: true
        sampling_ratio: 1
    throughputmeasurement/_d0_logs_googlecloud-1:
        enabled: true
        sampling_ratio: 1
    throughputmeasurement/_d0_metrics_googlecloud-0:
        enabled: true
        sampling_ratio: 1
    throughputmeasurement/_d0_metrics_googlecloud-1:
        enabled: true
        sampling_ratio: 1
    throughputmeasurement/_d1_logs_googlecloud-0:
        enabled: true
        sampling_ratio: 1
    throughputmeasurement/_d1_logs_googlecloud-1:
        enabled: true
        sampling_ratio: 1
    throughputmeasurement/_d1_metrics_googlecloud-0:
        enabled: true
        sampling_ratio: 1
    throughputmeasurement/_d1_metrics_googlecloud-1:
        enabled: true
        sampling_ratio: 1
    throughputmeasurement/_s0_logs_source0:
        enabled: true
        sampling_ratio: 1
    throughputmeasurement/_s0_metrics_source0:
        enabled: true
        sampling_ratio: 1
    throughputmeasurement/_s1_logs_source0:
        enabled: true
        sampling_ratio: 1
    throughputmeasurement/_s1_metrics_source0:
        enabled: true
        sampling_ratio: 1
exporters:
    googlecloud/googlecloud: null
    otlphttp/_agent_metrics:
        endpoint: /v1/otlphttp
        headers: {}
        retry_on_failure:
            enabled: true
            initial_interval: 5s
            max_elapsed_time: 30s
            max_interval: 5s
        sending_queue:
            enabled: true
            num_consumers: 1
            queue_size: 60
service:
    pipelines:
        logs/source0__googlecloud-0:
            receivers:
                - plugin/source0__macos
                - plugin/source0__journald
            processors:
                - throughputmeasurement/_s0_logs_source0
                - snapshotprocessor/_s0_source0
                - resourceattributetransposer/source0__processor0
                - resourceattributetransposer/source0__processor1
                - throughputmeasurement/_s1_logs_source0
                - throughputmeasurement/_d0_logs_googlecloud-0
                - snapshotprocessor/_d0_googlecloud-0
                - throughputmeasurement/_d1_logs_googlecloud-0
                - batch/googlecloud
                - snapshotprocessor
            exporters:
                - googlecloud/googlecloud
        logs/source0__googlecloud-1:
            receivers:
                - plugin/source0__macos
                - plugin/source0__journald
            processors:
                - snapshotprocessor/_s0_source0
                - resourceattributetransposer/source0__processor0
                - resourceattributetransposer/source0__processor1
                - throughputmeasurement/_d0_logs_googlecloud-1
                - snapshotprocessor/_d0_googlecloud-1
                - throughputmeasurement/_d1_logs_googlecloud-1
                - batch/googlecloud
                - snapshotprocessor
            exporters:
                - googlecloud/googlecloud
        metrics/_agent_metrics:
            receivers:
                - prometheus/_agent_metrics
            processors:
                - batch/_agent_metrics
            exporters:
                - otlphttp/_agent_metrics
        metrics/source0__googlecloud-0:
            receivers:
                - hostmetrics/source0
            processors:
                - throughputmeasurement/_s0_metrics_source0
                - snapshotprocessor/_s0_source0
                - resourceattributetransposer/source0__processor0
                - resourceattributetransposer/source0__processor1
                - throughputmeasurement/_s1_metrics_source0
                - throughputmeasurement/_d0_metrics_googlecloud-0
                - snapshotprocessor/_d0_googlecloud-0
                - throughputmeasurement/_d1_metrics_googlecloud-0
                - batch/googlecloud
                - snapshotprocessor
            exporters:
                - googlecloud/googlecloud
        metrics/source0__googlecloud-1:
            receivers:
                - hostmetrics/source0
            processors:
                - snapshotprocessor/_s0_source0
                - resourceattributetransposer/source0__processor0
                - resourceattributetransposer/source0__processor1
                - throughputmeasurement/_d0_metrics_googlecloud-1
                - snapshotprocessor/_d0_googlecloud-1
                - throughputmeasurement/_d1_metrics_googlecloud-1
                - batch/googlecloud
                - snapshotprocessor
            exporters:
                - googlecloud/googlecloud
`, "\n")

	require.Equal(t, expect, result)
}

func TestEvalConfigurationFailsMissingResource(t *testing.T) {
	store := newTestResourceStore()
	config := newTestConfiguration()

	postgresql := testResource[*SourceType](t, "sourcetype-macos.yaml")
	store.sourceTypes.add(postgresql)

	googleCloudType := testResource[*DestinationType](t, "destinationtype-googlecloud.yaml")
	store.destinationTypes.add(googleCloudType)

	googleCloud := testResource[*Destination](t, "destination-googlecloud.yaml")
	store.destinations.add(googleCloud)

	resourceAttributeTransposerType := testResource[*ProcessorType](t, "processortype-resourceattributetransposer.yaml")
	store.processorTypes.add(resourceAttributeTransposerType)

	configuration := testResource[*Configuration](t, "configuration-macos-processors.yaml")

	tests := []struct {
		name            string
		deleteResources func()
		expectError     error
		expect          string
	}{
		{
			name:            "deletes sourceType",
			deleteResources: func() { store.sourceTypes.remove(postgresql.Name()) },
			expectError:     errors.New("unknown SourceType: MacOS"),
		},
		{
			name:            "deletes googleCloudType",
			deleteResources: func() { store.destinationTypes.remove(googleCloudType.Name()) },
			expectError:     errors.New("unknown DestinationType: googlecloud"),
		},
		{
			name:            "deletes destination",
			deleteResources: func() { store.destinations.remove(googleCloud.Name()) },
			expectError:     errors.New("unknown Destination: googlecloud"),
		},
		{
			name:            "deletes processorType",
			deleteResources: func() { store.processorTypes.remove(resourceAttributeTransposerType.Name()) },
			expectError: errors.Join(
				errors.New("unknown ProcessorType: resource-attribute-transposer"),
				errors.New("unknown ProcessorType: resource-attribute-transposer"),
			),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			// before rendering, delete resources that we reference
			test.deleteResources()

			_, err := configuration.Render(context.TODO(), nil, config.BindPlaneURL(), config.BindPlaneInsecureSkipVerify(), store, GetOssOtelHeaders())
			require.Error(t, err)
			require.EqualError(t, test.expectError, err.Error())

			// reset for next iteration
			store.sourceTypes.add(postgresql)
			store.destinationTypes.add(googleCloudType)
			store.destinations.add(googleCloud)
			store.processorTypes.add(resourceAttributeTransposerType)
		})
	}
}

func TestConfigurationRender_DisabledDestination(t *testing.T) {
	store := newTestResourceStore()
	config := newTestConfiguration()

	macos := testResource[*SourceType](t, "sourcetype-macos.yaml")
	store.sourceTypes.add(macos)

	googleCloudType := testResource[*DestinationType](t, "destinationtype-googlecloud.yaml")
	store.destinationTypes.add(googleCloudType)

	googleCloud := testResource[*Destination](t, "destination-googlecloud.yaml")
	store.destinations.add(googleCloud)

	cabinType := testResource[*DestinationType](t, "destinationtype-cabin.yaml")
	store.destinationTypes.add(cabinType)

	cabin := testResource[*Destination](t, "destination-cabin.yaml")
	store.destinations.add(cabin)

	configuration := testResource[*Configuration](t, "configuration-macos-googlecloud-disabled.yaml")
	result, err := configuration.Render(context.TODO(), nil, config.BindPlaneURL(), config.BindPlaneInsecureSkipVerify(), store, GetOssOtelHeaders())
	require.NoError(t, err)

	// We expect the full pipeline, omitting the disabled googlecloud destination
	expect := strings.TrimLeft(`
receivers:
    plugin/source0__journald:
        plugin:
            name: journald
    plugin/source0__macos:
        parameters:
            - name: enable_system_log
              value: false
            - name: system_log_path
              value: /var/log/system.log
            - name: enable_install_log
              value: true
            - name: install_log_path
              value: /var/log/install.log
            - name: start_at
              value: end
        plugin:
            name: macos
    plugin/source1__journald:
        plugin:
            name: journald
    plugin/source1__macos:
        parameters:
            - name: enable_system_log
              value: true
            - name: system_log_path
              value: /var/log/system.log
            - name: enable_install_log
              value: true
            - name: install_log_path
              value: /var/log/install.log
            - name: start_at
              value: end
        plugin:
            name: macos
processors:
    batch/cabin-production-logs: null
exporters:
    observiq/cabin-production-logs:
        endpoint: https://nozzle.app.observiq.com
        secret_key: 2c088c5e-2afc-483b-be52-e2b657fcff08
        timeout: 10s
service:
    pipelines:
        logs/source0__cabin-production-logs-1:
            receivers:
                - plugin/source0__macos
                - plugin/source0__journald
            processors:
                - batch/cabin-production-logs
            exporters:
                - observiq/cabin-production-logs
        logs/source1__cabin-production-logs-1:
            receivers:
                - plugin/source1__macos
                - plugin/source1__journald
            processors:
                - batch/cabin-production-logs
            exporters:
                - observiq/cabin-production-logs
`, "\n")
	require.Equal(t, expect, result)
}
func TestConfigurationRender_DisabledSource(t *testing.T) {
	store := newTestResourceStore()
	config := newTestConfiguration()

	macos := testResource[*SourceType](t, "sourcetype-macos.yaml")
	store.sourceTypes.add(macos)

	fileLog := testResource[*SourceType](t, "sourcetype-filelog.yaml")
	store.sourceTypes.add(fileLog)

	googleCloudType := testResource[*DestinationType](t, "destinationtype-googlecloud.yaml")
	store.destinationTypes.add(googleCloudType)

	googleCloud := testResource[*Destination](t, "destination-googlecloud.yaml")
	store.destinations.add(googleCloud)

	configuration := testResource[*Configuration](t, "configuration-macos-source-disabled.yaml")
	result, err := configuration.Render(context.TODO(), nil, config.BindPlaneURL(), config.BindPlaneInsecureSkipVerify(), store, GetOssOtelHeaders())
	require.NoError(t, err)

	// We expect the full pipeline, omitting the disabled macOS source
	expect := strings.TrimLeft(`
receivers:
    plugin/source1:
        parameters:
            encoding: utf-8
            file_path:
                - /foo/bar/baz.log
            log_type: file
            multiline_line_start_pattern: ""
            parse_format: none
            start_at: end
        path: $OIQ_OTEL_COLLECTOR_HOME/plugins/file_logs.yaml
processors:
    batch/googlecloud: null
    resourcedetection/source1:
        detectors:
            - system
        system:
            hostname_sources:
                - os
exporters:
    googlecloud/googlecloud: null
service:
    pipelines:
        logs/source1__googlecloud-0:
            receivers:
                - plugin/source1
            processors:
                - resourcedetection/source1
                - batch/googlecloud
            exporters:
                - googlecloud/googlecloud
`, "\n")
	require.Equal(t, expect, result)
}

func TestConfigurationRender_DisabledProcessor(t *testing.T) {
	store := newTestResourceStore()
	config := newTestConfiguration()

	postgresql := testResource[*SourceType](t, "sourcetype-macos.yaml")
	store.sourceTypes.add(postgresql)

	googleCloudType := testResource[*DestinationType](t, "destinationtype-googlecloud.yaml")
	store.destinationTypes.add(googleCloudType)

	googleCloud := testResource[*Destination](t, "destination-googlecloud.yaml")
	store.destinations.add(googleCloud)

	resourceAttributeTransposerType := testResource[*ProcessorType](t, "processortype-resourceattributetransposer.yaml")
	store.processorTypes.add(resourceAttributeTransposerType)

	configuration := testResource[*Configuration](t, "configuration-macos-processors-disabled.yaml")
	result, err := configuration.Render(context.TODO(), nil, config.BindPlaneURL(), config.BindPlaneInsecureSkipVerify(), store, GetOssOtelHeaders())
	require.NoError(t, err)

	expect := strings.TrimLeft(`
receivers:
    hostmetrics/source0:
        collection_interval: 1m
        scrapers:
            load: null
    plugin/source0__journald:
        plugin:
            name: journald
    plugin/source0__macos:
        parameters:
            - name: enable_system_log
              value: false
            - name: system_log_path
              value: /var/log/system.log
            - name: enable_install_log
              value: true
            - name: install_log_path
              value: /var/log/install.log
            - name: start_at
              value: end
        plugin:
            name: macos
processors:
    batch/googlecloud: null
exporters:
    googlecloud/googlecloud: null
service:
    pipelines:
        logs/source0__googlecloud-0:
            receivers:
                - plugin/source0__macos
                - plugin/source0__journald
            processors:
                - batch/googlecloud
            exporters:
                - googlecloud/googlecloud
        metrics/source0__googlecloud-0:
            receivers:
                - hostmetrics/source0
            processors:
                - batch/googlecloud
            exporters:
                - googlecloud/googlecloud
`, "\n")

	require.Equal(t, expect, result)
}

func TestEvalConfiguration_FileLogStorage(t *testing.T) {
	t.Parallel()
	store := newTestResourceStore()
	config := newTestConfiguration()

	macos := testResource[*SourceType](t, "sourcetype-macos.yaml")
	store.sourceTypes.add(macos)

	filelog := testResource[*SourceType](t, "sourcetype-filelog-storage.yaml")
	store.sourceTypes.add(filelog)

	googleCloudType := testResource[*DestinationType](t, "destinationtype-googlecloud.yaml")
	store.destinationTypes.add(googleCloudType)

	googleCloud := testResource[*Destination](t, "destination-googlecloud.yaml")
	store.destinations.add(googleCloud)

	configuration := testResource[*Configuration](t, "configuration-filelog-storage.yaml")
	result, err := configuration.Render(context.TODO(), nil, config.BindPlaneURL(), config.BindPlaneInsecureSkipVerify(), store, GetOssOtelHeaders())
	require.NoError(t, err)

	expect := strings.TrimLeft(`
receivers:
    plugin/source0:
        parameters:
            encoding: utf-8
            file_path:
                - /foo/bar/baz.log
            log_type: file
            multiline_line_start_pattern: ""
            parse_format: none
            start_at: end
            storage: file_storage/source0
        path: $OIQ_OTEL_COLLECTOR_HOME/plugins/file_logs.yaml
    plugin/source1:
        parameters:
            encoding: utf-8
            file_path:
                - /foo/bar/baz2.log
            log_type: file
            multiline_line_start_pattern: ""
            parse_format: none
            start_at: end
            storage: file_storage/source1
        path: $OIQ_OTEL_COLLECTOR_HOME/plugins/file_logs.yaml
processors:
    batch/googlecloud: null
    resourcedetection/source0:
        detectors:
            - system
        system:
            hostname_sources:
                - os
    resourcedetection/source1:
        detectors:
            - system
        system:
            hostname_sources:
                - os
exporters:
    googlecloud/googlecloud: null
extensions:
    file_storage/source0:
        directory: /tmp/offset_storage_dir
    file_storage/source1:
        directory: /tmp/offset_storage_dir
service:
    extensions:
        - file_storage/source0
        - file_storage/source1
    pipelines:
        logs/source0__googlecloud-0:
            receivers:
                - plugin/source0
            processors:
                - resourcedetection/source0
                - batch/googlecloud
            exporters:
                - googlecloud/googlecloud
        logs/source1__googlecloud-0:
            receivers:
                - plugin/source1
            processors:
                - resourcedetection/source1
                - batch/googlecloud
            exporters:
                - googlecloud/googlecloud
`, "\n")

	require.Equal(t, expect, result)
}

func TestEvalConfiguration_TestAgentMetricsTLS(t *testing.T) {
	t.Parallel()
	store := newTestResourceStore()
	config := &testConfiguration{
		bindplaneURL:                "https://127.0.0.1:8443",
		bindplaneInsecureSkipVerify: false,
	}

	agent := Agent{
		Version: "v1.13.22",
	}
	otlp := testResource[*SourceType](t, "sourcetype-otlp.yaml")
	store.sourceTypes.add(otlp)

	otlpDestinationType := testResource[*DestinationType](t, "destinationtype-otlp.yaml")
	store.destinationTypes.add(otlpDestinationType)

	configuration := testResource[*Configuration](t, "configuration-otlp.yaml")

	result, err := configuration.Render(context.TODO(), &agent, config.BindPlaneURL(), config.BindPlaneInsecureSkipVerify(), store, GetOssOtelHeaders())
	require.NoError(t, err)

	expect := strings.TrimLeft(`
receivers:
    otlp/source0:
        protocols:
            grpc: null
            http: null
    prometheus/_agent_metrics:
        config:
            scrape_configs:
                - job_name: observiq-otel-collector
                  metric_relabel_configs:
                    - action: keep
                      regex: otelcol_processor_throughputmeasurement_.*
                      source_labels:
                        - __name__
                  scrape_interval: 10s
                  static_configs:
                    - labels:
                        agent: ""
                        configuration: otlp
                      targets:
                        - 0.0.0.0:8888
processors:
    batch/_agent_metrics: null
    batch/destination0: null
    snapshotprocessor: null
    snapshotprocessor/_d0_destination0-0: null
    snapshotprocessor/_s0_source0: null
    throughputmeasurement/_d0_logs_destination0-0:
        enabled: true
        sampling_ratio: 1
    throughputmeasurement/_d0_metrics_destination0-0:
        enabled: true
        sampling_ratio: 1
    throughputmeasurement/_d0_traces_destination0-0:
        enabled: true
        sampling_ratio: 1
    throughputmeasurement/_d1_logs_destination0-0:
        enabled: true
        sampling_ratio: 1
    throughputmeasurement/_d1_metrics_destination0-0:
        enabled: true
        sampling_ratio: 1
    throughputmeasurement/_d1_traces_destination0-0:
        enabled: true
        sampling_ratio: 1
    throughputmeasurement/_s0_logs_source0:
        enabled: true
        sampling_ratio: 1
    throughputmeasurement/_s0_metrics_source0:
        enabled: true
        sampling_ratio: 1
    throughputmeasurement/_s0_traces_source0:
        enabled: true
        sampling_ratio: 1
    throughputmeasurement/_s1_logs_source0:
        enabled: true
        sampling_ratio: 1
    throughputmeasurement/_s1_metrics_source0:
        enabled: true
        sampling_ratio: 1
    throughputmeasurement/_s1_traces_source0:
        enabled: true
        sampling_ratio: 1
exporters:
    otlp/destination0:
        endpoint: otelcol:4317
    otlphttp/_agent_metrics:
        endpoint: https://127.0.0.1:8443/v1/otlphttp
        headers: {}
        retry_on_failure:
            enabled: true
            initial_interval: 5s
            max_elapsed_time: 30s
            max_interval: 5s
        sending_queue:
            enabled: true
            num_consumers: 1
            queue_size: 60
service:
    pipelines:
        logs/source0__destination0-0:
            receivers:
                - otlp/source0
            processors:
                - throughputmeasurement/_s0_logs_source0
                - snapshotprocessor/_s0_source0
                - throughputmeasurement/_s1_logs_source0
                - throughputmeasurement/_d0_logs_destination0-0
                - snapshotprocessor/_d0_destination0-0
                - throughputmeasurement/_d1_logs_destination0-0
                - batch/destination0
                - snapshotprocessor
            exporters:
                - otlp/destination0
        metrics/_agent_metrics:
            receivers:
                - prometheus/_agent_metrics
            processors:
                - batch/_agent_metrics
            exporters:
                - otlphttp/_agent_metrics
        metrics/source0__destination0-0:
            receivers:
                - otlp/source0
            processors:
                - throughputmeasurement/_s0_metrics_source0
                - snapshotprocessor/_s0_source0
                - throughputmeasurement/_s1_metrics_source0
                - throughputmeasurement/_d0_metrics_destination0-0
                - snapshotprocessor/_d0_destination0-0
                - throughputmeasurement/_d1_metrics_destination0-0
                - batch/destination0
                - snapshotprocessor
            exporters:
                - otlp/destination0
        traces/source0__destination0-0:
            receivers:
                - otlp/source0
            processors:
                - throughputmeasurement/_s0_traces_source0
                - snapshotprocessor/_s0_source0
                - throughputmeasurement/_s1_traces_source0
                - throughputmeasurement/_d0_traces_destination0-0
                - snapshotprocessor/_d0_destination0-0
                - throughputmeasurement/_d1_traces_destination0-0
                - batch/destination0
                - snapshotprocessor
            exporters:
                - otlp/destination0
`, "\n")

	require.Equal(t, expect, result)
}

func TestEvalConfiguration_TestAgentMetricsTLSInsecure(t *testing.T) {
	t.Parallel()
	store := newTestResourceStore()
	config := &testConfiguration{
		bindplaneURL:                "https://127.0.0.1:8443",
		bindplaneInsecureSkipVerify: true,
	}

	agent := Agent{
		Version: "v1.13.22",
	}
	otlp := testResource[*SourceType](t, "sourcetype-otlp.yaml")
	store.sourceTypes.add(otlp)

	otlpDestinationType := testResource[*DestinationType](t, "destinationtype-otlp.yaml")
	store.destinationTypes.add(otlpDestinationType)

	configuration := testResource[*Configuration](t, "configuration-otlp.yaml")

	result, err := configuration.Render(context.TODO(), &agent, config.BindPlaneURL(), config.BindPlaneInsecureSkipVerify(), store, GetOssOtelHeaders())
	require.NoError(t, err)

	expect := strings.TrimLeft(`
receivers:
    otlp/source0:
        protocols:
            grpc: null
            http: null
    prometheus/_agent_metrics:
        config:
            scrape_configs:
                - job_name: observiq-otel-collector
                  metric_relabel_configs:
                    - action: keep
                      regex: otelcol_processor_throughputmeasurement_.*
                      source_labels:
                        - __name__
                  scrape_interval: 10s
                  static_configs:
                    - labels:
                        agent: ""
                        configuration: otlp
                      targets:
                        - 0.0.0.0:8888
processors:
    batch/_agent_metrics: null
    batch/destination0: null
    snapshotprocessor: null
    snapshotprocessor/_d0_destination0-0: null
    snapshotprocessor/_s0_source0: null
    throughputmeasurement/_d0_logs_destination0-0:
        enabled: true
        sampling_ratio: 1
    throughputmeasurement/_d0_metrics_destination0-0:
        enabled: true
        sampling_ratio: 1
    throughputmeasurement/_d0_traces_destination0-0:
        enabled: true
        sampling_ratio: 1
    throughputmeasurement/_d1_logs_destination0-0:
        enabled: true
        sampling_ratio: 1
    throughputmeasurement/_d1_metrics_destination0-0:
        enabled: true
        sampling_ratio: 1
    throughputmeasurement/_d1_traces_destination0-0:
        enabled: true
        sampling_ratio: 1
    throughputmeasurement/_s0_logs_source0:
        enabled: true
        sampling_ratio: 1
    throughputmeasurement/_s0_metrics_source0:
        enabled: true
        sampling_ratio: 1
    throughputmeasurement/_s0_traces_source0:
        enabled: true
        sampling_ratio: 1
    throughputmeasurement/_s1_logs_source0:
        enabled: true
        sampling_ratio: 1
    throughputmeasurement/_s1_metrics_source0:
        enabled: true
        sampling_ratio: 1
    throughputmeasurement/_s1_traces_source0:
        enabled: true
        sampling_ratio: 1
exporters:
    otlp/destination0:
        endpoint: otelcol:4317
    otlphttp/_agent_metrics:
        endpoint: https://127.0.0.1:8443/v1/otlphttp
        headers: {}
        retry_on_failure:
            enabled: true
            initial_interval: 5s
            max_elapsed_time: 30s
            max_interval: 5s
        sending_queue:
            enabled: true
            num_consumers: 1
            queue_size: 60
        tls:
            insecure_skip_verify: true
service:
    pipelines:
        logs/source0__destination0-0:
            receivers:
                - otlp/source0
            processors:
                - throughputmeasurement/_s0_logs_source0
                - snapshotprocessor/_s0_source0
                - throughputmeasurement/_s1_logs_source0
                - throughputmeasurement/_d0_logs_destination0-0
                - snapshotprocessor/_d0_destination0-0
                - throughputmeasurement/_d1_logs_destination0-0
                - batch/destination0
                - snapshotprocessor
            exporters:
                - otlp/destination0
        metrics/_agent_metrics:
            receivers:
                - prometheus/_agent_metrics
            processors:
                - batch/_agent_metrics
            exporters:
                - otlphttp/_agent_metrics
        metrics/source0__destination0-0:
            receivers:
                - otlp/source0
            processors:
                - throughputmeasurement/_s0_metrics_source0
                - snapshotprocessor/_s0_source0
                - throughputmeasurement/_s1_metrics_source0
                - throughputmeasurement/_d0_metrics_destination0-0
                - snapshotprocessor/_d0_destination0-0
                - throughputmeasurement/_d1_metrics_destination0-0
                - batch/destination0
                - snapshotprocessor
            exporters:
                - otlp/destination0
        traces/source0__destination0-0:
            receivers:
                - otlp/source0
            processors:
                - throughputmeasurement/_s0_traces_source0
                - snapshotprocessor/_s0_source0
                - throughputmeasurement/_s1_traces_source0
                - throughputmeasurement/_d0_traces_destination0-0
                - snapshotprocessor/_d0_destination0-0
                - throughputmeasurement/_d1_traces_destination0-0
                - batch/destination0
                - snapshotprocessor
            exporters:
                - otlp/destination0
`, "\n")

	require.Equal(t, expect, result)
}

func strp(s string) *string {
	return &s
}

func TestUpdateStatus(t *testing.T) {
	rolloutOptions := RolloutOptions{
		MaxErrors: 0,
		PhaseAgentCount: PhaseAgentCount{
			Initial:    3,
			Multiplier: 2,
			Maximum:    30,
		},
	}

	tests := []struct {
		name                   string
		initialRollout         Rollout
		progress               RolloutProgress
		expectNewAgentsPending int
		expectRollout          Rollout
	}{
		{
			name: "not started, no progress",
			initialRollout: Rollout{
				Status: RolloutStatusPending,
			},
			progress:               RolloutProgress{},
			expectNewAgentsPending: 0,
			expectRollout: Rollout{
				Status: RolloutStatusPending,
			},
		},
		{
			name: "too many errors",
			initialRollout: Rollout{
				Status: RolloutStatusStarted,
			},
			progress: RolloutProgress{
				Errors: 10,
			},
			expectNewAgentsPending: 0,
			expectRollout: Rollout{
				Status: RolloutStatusError,
				Progress: RolloutProgress{
					Errors: 10,
				},
			},
		},
		{
			name: "progress, still waiting",
			initialRollout: Rollout{
				Status: RolloutStatusStarted,
			},
			progress: RolloutProgress{
				Completed: 10,
				Pending:   9,
				Waiting:   1,
			},
			expectNewAgentsPending: 0,
			expectRollout: Rollout{
				Status: RolloutStatusStarted,
				Progress: RolloutProgress{
					Completed: 10,
					Pending:   9,
					Waiting:   1,
				},
			},
		},
		{
			name: "progress, new phase",
			initialRollout: Rollout{
				Status:  RolloutStatusStarted,
				Options: rolloutOptions,
				Phase:   2,
			},
			progress: RolloutProgress{
				Completed: 10,
				Pending:   0,
				Waiting:   20,
			},
			expectNewAgentsPending: 12,
			expectRollout: Rollout{
				Status:  RolloutStatusStarted,
				Options: rolloutOptions,
				Phase:   3,
				Progress: RolloutProgress{
					Completed: 10,
					Pending:   12,
					Waiting:   8,
				},
			},
		},
		{
			name: "progress, max phase size",
			initialRollout: Rollout{
				Status:  RolloutStatusStarted,
				Options: DefaultRolloutOptions[RolloutLarge],
				Phase:   20,
			},
			progress: RolloutProgress{
				Completed: 10,
				Pending:   0,
				Waiting:   200,
			},
			expectNewAgentsPending: DefaultRolloutOptions[RolloutLarge].PhaseAgentCount.Maximum,
			expectRollout: Rollout{
				Status:  RolloutStatusStarted,
				Options: DefaultRolloutOptions[RolloutLarge],
				Phase:   21,
				Progress: RolloutProgress{
					Completed: 10,
					Pending:   DefaultRolloutOptions[RolloutLarge].PhaseAgentCount.Maximum,
					Waiting:   200 - DefaultRolloutOptions[RolloutLarge].PhaseAgentCount.Maximum,
				},
			},
		},
		{
			name: "progress, last phase",
			initialRollout: Rollout{
				Status:  RolloutStatusStarted,
				Options: DefaultRolloutOptions[RolloutLarge],
				Phase:   2,
			},
			progress: RolloutProgress{
				Completed: 10,
				Pending:   0,
				Waiting:   2,
			},
			expectNewAgentsPending: 2,
			expectRollout: Rollout{
				Status:  RolloutStatusStarted,
				Options: DefaultRolloutOptions[RolloutLarge],
				Phase:   3,
				Progress: RolloutProgress{
					Completed: 10,
					Pending:   2,
					Waiting:   0,
				},
			},
		},
		{
			name: "progress, complete",
			initialRollout: Rollout{
				Status:  RolloutStatusStarted,
				Options: DefaultRolloutOptions[RolloutLarge],
				Phase:   2,
			},
			progress: RolloutProgress{
				Completed: 10,
				Pending:   0,
				Waiting:   0,
			},
			expectNewAgentsPending: 0,
			expectRollout: Rollout{
				Status:  RolloutStatusStable,
				Options: DefaultRolloutOptions[RolloutLarge],
				Phase:   2,
				Progress: RolloutProgress{
					Completed: 10,
					Pending:   0,
					Waiting:   0,
				},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			rollout := test.initialRollout
			newAgentsPending := rollout.UpdateStatus(test.progress)
			require.Equal(t, test.expectNewAgentsPending, newAgentsPending)
			require.Equal(t, test.expectRollout, rollout)
		})
	}
}

func TestResourceConfigurationShallowEqual(t *testing.T) {
	t.Run("Equal", func(t *testing.T) {
		rc1 := newTestResourceConfiguration(t)
		rc2 := newTestResourceConfiguration(t)
		require.True(t, rc1.ShallowEqual(rc2))
	})
	t.Run("ID Not equal", func(t *testing.T) {
		rc1 := newTestResourceConfiguration(t)
		rc2 := newTestResourceConfiguration(t)
		rc1.ID = "different"

		require.False(t, rc1.ShallowEqual(rc2))
	})
	t.Run("Name Not equal", func(t *testing.T) {
		rc1 := newTestResourceConfiguration(t)
		rc2 := newTestResourceConfiguration(t)
		rc1.Name = "different"

		require.False(t, rc1.ShallowEqual(rc2))
	})
	t.Run("DisplayName Not equal", func(t *testing.T) {
		rc1 := newTestResourceConfiguration(t)
		rc2 := newTestResourceConfiguration(t)
		rc1.DisplayName = "different"

		require.False(t, rc1.ShallowEqual(rc2))
	})
	t.Run("Type Not equal", func(t *testing.T) {
		rc1 := newTestResourceConfiguration(t)
		rc2 := newTestResourceConfiguration(t)
		rc1.Type = "different"

		require.False(t, rc1.ShallowEqual(rc2))
	})
	t.Run("Disabled Not equal", func(t *testing.T) {
		rc1 := newTestResourceConfiguration(t)
		rc2 := newTestResourceConfiguration(t)
		rc1.Disabled = !rc2.Disabled

		require.False(t, rc1.ShallowEqual(rc2))
	})
	t.Run("Different Number of Parameters", func(t *testing.T) {
		rc1 := newTestResourceConfiguration(t)
		rc2 := newTestResourceConfiguration(t)
		rc1.Parameters = append(rc1.Parameters,
			Parameter{
				Name:      "param1",
				Sensitive: false,
				Value:     new(string),
			},
		)

		require.False(t, rc1.ShallowEqual(rc2))
	})
	t.Run("Parameters Differ by Name", func(t *testing.T) {
		rc1 := newTestResourceConfiguration(t)
		rc2 := newTestResourceConfiguration(t)
		rc1.Parameters = []Parameter{
			rc2.Parameters[0],
		}

		rc1.Parameters[0].Name = "Different"

		require.False(t, rc1.ShallowEqual(rc2))
	})
}
func TestConfigurationPrintableFieldTitles(t *testing.T) {
	conf := Configuration{}
	expected := []string{"Name", "Version", "Match"}
	require.Equal(t, expected, conf.PrintableFieldTitles())
}

func TestConfigurationPrintableVersionFieldTitles(t *testing.T) {
	conf := Configuration{}
	expected := []string{"Name", "Version", "Date", "Match", "Current", "Pending", "Rollout"}
	require.Equal(t, expected, conf.PrintableVersionFieldTitles())
}

func TestConfigurationPrintableFieldValue(t *testing.T) {
	meta := ResourceMeta{
		Metadata: Metadata{
			Name: "Test",
		},
	}
	rollout := RolloutStatusPending
	status := ConfigurationStatus{Current: true, Pending: false, Rollout: Rollout{}}
	conf := Configuration{ResourceMeta: meta, Spec: ConfigurationSpec{}, StatusType: NewStatusType(status)}

	require.Equal(t, "Test", conf.PrintableFieldValue("Name"))
	require.Equal(t, "*", conf.PrintableFieldValue("Current"))
	require.Equal(t, "", conf.PrintableFieldValue("Pending"))
	require.Equal(t, rollout.String(), conf.PrintableFieldValue("Rollout"))
	require.Equal(t, "-", conf.PrintableFieldValue("Nonexistent Title"))
}

func TestConfigurationGraph(t *testing.T) {
	ctx := context.Background()
	conf := Configuration{
		Spec: ConfigurationSpec{
			Sources: []ResourceConfiguration{
				{
					Name: "s1",
				},
				{
					Name: "s2",
				},
				{
					Name: "s3",
				},
			},
			Destinations: []ResourceConfiguration{
				{
					Name: "d1",
				},
				{
					Name: "d2",
				},
			},
		},
	}

	src := &Source{}
	dest := &Destination{}
	srcType := &SourceType{}
	destType := &DestinationType{}

	mockResStore := NewMockResourceStore(t)

	// We don't need the specific information for the source and destination types so we can just return generic types
	mockResStore.On("SourceType", mock.Anything, mock.Anything).Return(srcType, nil)
	mockResStore.On("DestinationType", mock.Anything, mock.Anything).Return(destType, nil)
	mockResStore.On("Source", mock.Anything, mock.Anything).Return(src, nil)
	mockResStore.On("Destination", mock.Anything, mock.Anything).Return(dest, nil)

	// Call Graph method
	g, err := conf.Graph(ctx, mockResStore)
	require.NoError(t, err)
	require.NotNil(t, g)

	// Check nodes in the graph
	require.Len(t, g.Sources, len(conf.Spec.Sources), "The number of source nodes should match the number of sources")
	require.Len(t, g.Targets, len(conf.Spec.Destinations), "The number of destination nodes should match the number of destinations")
	require.Len(t, g.Intermediates, len(conf.Spec.Sources)+len(conf.Spec.Destinations), "The number of intermediate nodes should match the total number of sources and destinations")

	// Check connections in the graph
	for _, srcNode := range g.Sources {
		outgoingEdges := 0
		for _, edge := range g.Edges {
			if edge.Source == srcNode.ID {
				outgoingEdges++
			}
		}
		require.Equal(t, 1, outgoingEdges, fmt.Sprintf("Each source node should be connected to one intermediate node, source node ID: %s", srcNode.ID))
	}
	for _, tgtNode := range g.Targets {
		incomingEdges := 0
		for _, edge := range g.Edges {
			if edge.Target == tgtNode.ID {
				incomingEdges++
			}
		}
		require.Equal(t, 1, incomingEdges, fmt.Sprintf("Each destination node should be connected from one intermediate node, destination node ID: %s", tgtNode.ID))
	}
}

func TestCreateRouteReceiver(t *testing.T) {
	t.Run("it should return a route receiver with a unique name", func(t *testing.T) {
		name := "route_receiver"
		errorHandler := func(error) {} // No-op errorHandler for this test

		receiverName, partials := createRouteReceiver(context.Background(), name, errorHandler)

		// Check the receiver name is correct
		require.Equal(t, name, receiverName)

		// Check that the partial configuration is not nil and contains a receiver
		require.NotNil(t, partials)

		// Check that the receiver is a 'route' receiver
		for _, partial := range partials {
			for _, receivers := range partial.Receivers {
				for receiverID := range receivers {
					require.Contains(t, string(receiverID), "route")
				}
			}
		}
	})
}

// TestMaskSensitiveParameters tests the maskSensitiveParameters method of ResourceConfiguration
func TestMaskSensitiveParametersRC(t *testing.T) {
	ctx := context.Background()
	// Creating an instance of ResourceConfiguration with parameters
	rc := &ResourceConfiguration{
		ParameterizedSpec: ParameterizedSpec{
			Parameters: []Parameter{
				{
					Name:      "TestParameter1",
					Value:     "TestValue1",
					Sensitive: true,
				},
				{
					Name:      "TestParameter2",
					Value:     "TestValue2",
					Sensitive: false,
				},
			},
		},
	}

	// Calling the method
	rc.maskSensitiveParameters(ctx)

	// Asserting that sensitive parameter values are replaced with the placeholder
	for _, param := range rc.Parameters {
		if param.Sensitive {
			require.Equal(t, SensitiveParameterPlaceholder, param.Value)
		} else {
			require.NotEqual(t, SensitiveParameterPlaceholder, param.Value)
		}
	}
}

// TestPreserveSensitiveParametersRC tests the preserveSensitiveParameters method of ResourceConfiguration
func TestPreserveSensitiveParametersRC(t *testing.T) {
	ctx := context.Background()

	existing := &ResourceConfiguration{
		ParameterizedSpec: ParameterizedSpec{
			Parameters: []Parameter{
				{
					Name:      "TestParameter1",
					Value:     "ExistingValue1",
					Sensitive: true,
				},
				{
					Name:      "TestParameter2",
					Value:     "ExistingValue2",
					Sensitive: false,
				},
			},
			Processors: []ResourceConfiguration{
				{
					ParameterizedSpec: ParameterizedSpec{
						Parameters: []Parameter{
							{
								Name:      "ProcessorParameter1",
								Value:     "ExistingProcessorValue1",
								Sensitive: true,
							},
							{
								Name:      "ProcessorParameter2",
								Value:     "ExistingProcessorValue2",
								Sensitive: false,
							},
						},
					},
				},
			},
		},
	}

	rc := &ResourceConfiguration{
		ParameterizedSpec: ParameterizedSpec{
			Parameters: []Parameter{
				{
					Name:  "TestParameter1",
					Value: SensitiveParameterPlaceholder,
				},
				{
					Name:  "TestParameter2",
					Value: SensitiveParameterPlaceholder,
				},
			},
			Processors: []ResourceConfiguration{
				{
					ParameterizedSpec: ParameterizedSpec{
						Parameters: []Parameter{
							{
								Name:  "ProcessorParameter1",
								Value: SensitiveParameterPlaceholder,
							},
							{
								Name:  "ProcessorParameter2",
								Value: SensitiveParameterPlaceholder,
							},
						},
					},
				},
			},
		},
	}

	// Calling the method
	rc.preserveSensitiveParameters(ctx, existing)

	// Asserting that sensitive parameter values are replaced with the value from the existing resource
	for i, param := range rc.Parameters {
		require.Equal(t, existing.Parameters[i].Value, param.Value)
	}

	// Now also assert for the processors
	for i, processor := range rc.Processors {
		for j, param := range processor.Parameters {
			require.Equal(t, existing.Processors[i].Parameters[j].Value, param.Value)
		}
	}
}

func TestRolloutPrintableFieldValue(t *testing.T) {
	// Create an instance of Rollout
	r := &Rollout{
		Name:   "TestRollout",
		Status: 1,
		Phase:  2,
		Progress: RolloutProgress{
			Completed: 3,
			Errors:    4,
			Pending:   5,
			Waiting:   6,
		},
	}

	// Create a map of expected values for each field title
	expectedValues := map[string]string{
		"Name":      "TestRollout",
		"Status":    "Started",
		"Phase":     "2",
		"Completed": "3",
		"Errors":    "4",
		"Pending":   "5",
		"Waiting":   "6",
	}

	titles := r.PrintableFieldTitles()

	for _, title := range titles {
		val := r.PrintableFieldValue(title)
		// Check if the returned value for each title matches the expected value
		require.Equal(t, expectedValues[title], val)
	}

	// Check for a title that does not exist
	val := r.PrintableFieldValue("NonExistentTitle")
	require.Equal(t, "-", val)
}

func TestConfigurationDuplicate(t *testing.T) {
	original := &Configuration{
		ResourceMeta: ResourceMeta{
			Metadata: Metadata{
				ID:   uuid.New().String(),
				Name: "original_configuration",
			},
		},
		Spec: ConfigurationSpec{
			Selector: AgentSelector{
				MatchLabels: map[string]string{
					"configuration": "original_configuration",
				},
			},
		},
	}

	newName := "test_name"
	duplicate := original.Duplicate(newName)

	// Check if the Name, ID and matchLabel have been correctly updated
	require.Equal(t, newName, duplicate.Metadata.Name, "The Name was not updated correctly")
	require.NotEqual(t, original.Metadata.ID, duplicate.Metadata.ID, "The ID should be different")
	require.Equal(t, newName, duplicate.Spec.Selector.MatchLabels["configuration"], "The matchLabel was not updated correctly")

	// Set these fields to the original's fields to compare the remaining data
	duplicate.Metadata.Name = original.Metadata.Name
	duplicate.Metadata.ID = original.Metadata.ID
	duplicate.Spec.Selector.MatchLabels["configuration"] = original.Spec.Selector.MatchLabels["configuration"]

	require.Equal(t, original, duplicate, "The duplicated configuration should be identical except for the Name, ID and matchLabel fields")
}

func TestDestinationPrintableFields(t *testing.T) {
	// Initialize a Destination
	d := &Destination{
		ResourceMeta: ResourceMeta{
			APIVersion: "v1",
			Kind:       "destination",
			Metadata: Metadata{
				Name:        "test_destination",
				Description: "This is a test destination",
				ID:          "123",
			},
		},
	}

	expectedTitles := []string{"Name", "Type", "Description"}
	require.Equal(t, expectedTitles, d.PrintableFieldTitles(), "PrintableFieldTitles didn't return the expected titles")

	// Check if PrintableFieldValue correctly returns the values for each title
	for _, title := range expectedTitles {
		var expectedValue string

		switch title {
		case "Name":
			expectedValue = d.Metadata.Name
		case "Type":
			expectedValue = d.ResourceTypeName() // ResourceTypeName() should be defined in the Destination struct
		case "Description":
			expectedValue = d.Metadata.Description
		}

		require.Equal(t, expectedValue, d.PrintableFieldValue(title), "PrintableFieldValue didn't return the expected value for the title: "+title)
	}
}

type MockIndexer struct {
	fields map[string]string
}

func (m *MockIndexer) Add(key string, value string) {
	m.fields[key] = value
}

func TestConfigurationIndexFields(t *testing.T) {
	mockIndexer := &MockIndexer{fields: make(map[string]string)}

	configuration := &Configuration{
		ResourceMeta: ResourceMeta{
			Metadata: Metadata{
				Name: "TestConfiguration",
			},
		},
		Spec: ConfigurationSpec{
			Sources: []ResourceConfiguration{
				{
					Name: "TestSource",
					ParameterizedSpec: ParameterizedSpec{
						Type: "TestSourceType",
					},
				},
			},
			Destinations: []ResourceConfiguration{
				{
					Name: "TestDestination",

					ParameterizedSpec: ParameterizedSpec{
						Type: "TestDestinationType",
					},
				},
			},
		},
		StatusType: StatusType[ConfigurationStatus]{},
	}

	configuration.IndexFields(mockIndexer.Add)

	// Assertions
	require.Equal(t, "TestConfiguration", mockIndexer.fields["name"])
	require.Equal(t, "modular", mockIndexer.fields["type"])
	require.Equal(t, "TestSource", mockIndexer.fields["source"])
	require.Equal(t, "TestDestination", mockIndexer.fields["destination"])
	require.Equal(t, "Pending", mockIndexer.fields["rollout-status"])
}

func TestNewConfiguration(t *testing.T) {
	name := "TestConfiguration"
	raw := "TestRawConfiguration"
	spec := ConfigurationSpec{
		Raw: raw,
	}

	// Test NewConfiguration
	config := NewConfiguration(name)
	require.Equal(t, name, config.Metadata.Name)
	require.Equal(t, version.V1, config.APIVersion)
	require.Equal(t, KindConfiguration, config.Kind)
	require.NotNil(t, config.Metadata.Labels)
	require.Equal(t, ConfigurationSpec{}, config.Spec)

	// Test NewRawConfiguration
	rawConfig := NewRawConfiguration(name, raw)
	require.Equal(t, name, rawConfig.Metadata.Name)
	require.Equal(t, version.V1, rawConfig.APIVersion)
	require.Equal(t, KindConfiguration, rawConfig.Kind)
	require.NotNil(t, rawConfig.Metadata.Labels)
	require.Equal(t, spec, rawConfig.Spec)

	// Test NewConfigurationWithSpec
	specConfig := NewConfigurationWithSpec(name, spec)
	require.Equal(t, name, specConfig.Metadata.Name)
	require.Equal(t, version.V1, specConfig.APIVersion)
	require.Equal(t, KindConfiguration, specConfig.Kind)
	require.NotNil(t, specConfig.Metadata.Labels)
	require.Equal(t, spec, specConfig.Spec)
}
