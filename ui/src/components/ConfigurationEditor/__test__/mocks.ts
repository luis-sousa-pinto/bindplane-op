import { MockedResponse } from "@apollo/client/testing";
import {
  NO_VERSION_HISTORY,
  HISTORY_LATEST_IS_NEW,
  HISTORY_LATEST_IS_CURRENT,
  NO_VERSION_HISTORY_WITH_PENDING,
  HISTORY_WITH_PENDING_AND_NEW,
} from ".";
import {
  GetConfigurationVersionsDocument,
  SourceTypeDocument,
  GetDestinationWithTypeDocument,
  GetConfigurationQuery,
  GetConfigurationDocument,
} from "../../../graphql/generated";

export const VERSION_MOCK_NO_HISTORY: MockedResponse = {
  request: {
    query: GetConfigurationVersionsDocument,
    variables: {
      name: "linux-metrics",
    },
  },
  result: {
    data: NO_VERSION_HISTORY,
  },
};

export const VERSION_MOCK_NO_HISTORY_WITH_PENDING: MockedResponse = {
  request: {
    query: GetConfigurationVersionsDocument,
    variables: {
      name: "linux-metrics",
    },
  },
  result: {
    data: NO_VERSION_HISTORY_WITH_PENDING,
  },
};

export const VERSION_MOCK_WITH_HISTORY: MockedResponse = {
  request: {
    query: GetConfigurationVersionsDocument,
    variables: {
      name: "linux-metrics",
    },
  },
  result: {
    data: HISTORY_LATEST_IS_CURRENT,
  },
};

export const VERSION_MOCK_WITH_HISTORY_AND_NEW: MockedResponse = {
  request: {
    query: GetConfigurationVersionsDocument,
    variables: {
      name: "linux-metrics",
    },
  },
  result: {
    data: HISTORY_LATEST_IS_NEW,
  },
};

export const VERSION_MOCK_WITH_NEW_AND_PENDING: MockedResponse = {
  request: {
    query: GetConfigurationVersionsDocument,
    variables: {
      name: "linux-metrics",
    },
  },
  result: {
    data: HISTORY_WITH_PENDING_AND_NEW,
  },
};

export const SOURCE_TYPE_HOST_MOCK: MockedResponse = {
  request: {
    query: SourceTypeDocument,
    variables: {
      name: "host",
    },
  },
  result: {
    data: {
      sourceType: {
        metadata: {
          id: "fc444e19-2bee-4f6d-b103-227e2ef78a70",
          displayName: "Host",
          name: "host",
          version: 2,
          icon: "/icons/sources/host.svg",
          description: "Collect metrics from the collector's host.",
        },
        spec: {
          parameters: [
            {
              label: "",
              name: "metric_filtering",
              description: "",
              required: false,
              type: "metrics",
              default: [
                "system.disk.io",
                "system.disk.io_time",
                "system.disk.merged",
                "system.disk.operation_time",
                "system.disk.operations",
                "system.disk.pending_operations",
                "system.disk.weighted_io_time",
                "system.processes.count",
                "system.processes.created",
                "system.cpu.time",
                "system.cpu.utilization",
              ],
              documentation: null,
              relevantIf: null,
              advancedConfig: false,
              validValues: null,
              options: {
                creatable: false,
                trackUnchecked: false,
                sectionHeader: null,
                gridColumns: null,
                labels: {},
                metricCategories: [
                  {
                    label: "Filesystem Metrics",
                    column: 0,
                    metrics: [
                      {
                        name: "system.filesystem.inodes.usage",
                        description: null,
                        kpi: null,
                      },
                      {
                        name: "system.filesystem.usage",
                        description: null,
                        kpi: null,
                      },
                      {
                        name: "system.filesystem.utilization",
                        description: null,
                        kpi: null,
                      },
                    ],
                  },
                  {
                    label: "Memory Metrics",
                    column: 0,
                    metrics: [
                      {
                        name: "system.memory.usage",
                        description: null,
                        kpi: null,
                      },
                      {
                        name: "system.memory.utilization",
                        description: null,
                        kpi: null,
                      },
                    ],
                  },
                  {
                    label: "Network Metrics",
                    column: 0,
                    metrics: [
                      {
                        name: "system.network.connections",
                        description: null,
                        kpi: null,
                      },
                      {
                        name: "system.network.conntrack.count",
                        description: null,
                        kpi: null,
                      },
                      {
                        name: "system.network.conntrack.max",
                        description: null,
                        kpi: null,
                      },
                      {
                        name: "system.network.dropped",
                        description: null,
                        kpi: null,
                      },
                      {
                        name: "system.network.errors",
                        description: null,
                        kpi: null,
                      },
                      {
                        name: "system.network.io",
                        description: null,
                        kpi: null,
                      },
                      {
                        name: "system.network.packets",
                        description: null,
                        kpi: null,
                      },
                    ],
                  },
                  {
                    label: "Paging Metrics",
                    column: 0,
                    metrics: [
                      {
                        name: "system.paging.faults",
                        description: null,
                        kpi: null,
                      },
                      {
                        name: "system.paging.operations",
                        description: null,
                        kpi: null,
                      },
                      {
                        name: "system.paging.usage",
                        description: null,
                        kpi: null,
                      },
                      {
                        name: "system.paging.utilization",
                        description: null,
                        kpi: null,
                      },
                    ],
                  },
                  {
                    label: "Load Metrics",
                    column: 1,
                    metrics: [
                      {
                        name: "system.cpu.load_average.15m",
                        description: null,
                        kpi: null,
                      },
                      {
                        name: "system.cpu.load_average.1m",
                        description: null,
                        kpi: null,
                      },
                      {
                        name: "system.cpu.load_average.5m",
                        description: null,
                        kpi: null,
                      },
                    ],
                  },
                  {
                    label: "CPU Metrics",
                    column: 1,
                    metrics: [
                      {
                        name: "system.cpu.time",
                        description: null,
                        kpi: null,
                      },
                      {
                        name: "system.cpu.utilization",
                        description: null,
                        kpi: null,
                      },
                    ],
                  },
                  {
                    label: "Disk Metrics",
                    column: 1,
                    metrics: [
                      {
                        name: "system.disk.io",
                        description: null,
                        kpi: null,
                      },
                      {
                        name: "system.disk.io_time",
                        description: null,
                        kpi: null,
                      },
                      {
                        name: "system.disk.merged",
                        description: null,
                        kpi: null,
                      },
                      {
                        name: "system.disk.operation_time",
                        description: null,
                        kpi: null,
                      },
                      {
                        name: "system.disk.operations",
                        description: null,
                        kpi: null,
                      },
                      {
                        name: "system.disk.pending_operations",
                        description: null,
                        kpi: null,
                      },
                      {
                        name: "system.disk.weighted_io_time",
                        description: null,
                        kpi: null,
                      },
                    ],
                  },
                  {
                    label: "Processes Metrics",
                    column: 1,
                    metrics: [
                      {
                        name: "system.processes.count",
                        description: null,
                        kpi: null,
                      },
                      {
                        name: "system.processes.created",
                        description: null,
                        kpi: null,
                      },
                    ],
                  },
                ],
              },
            },
            {
              label: "Process Metrics",
              name: "enable_process",
              description:
                "Enable to collect process metrics. Compatible with Linux and Windows. The collector must be running as root (Linux) and Administrator (Windows).",
              required: false,
              type: "bool",
              default: true,
              documentation: null,
              relevantIf: null,
              advancedConfig: false,
              validValues: null,
              options: {
                creatable: false,
                trackUnchecked: false,
                sectionHeader: true,
                gridColumns: null,
                labels: {},
                metricCategories: null,
              },
            },
            {
              label: "",
              name: "process_metrics_filtering",
              description: "",
              required: false,
              type: "metrics",
              default: [],
              documentation: null,
              relevantIf: [
                {
                  name: "enable_process",
                  operator: "equals",
                  value: true,
                },
              ],
              advancedConfig: false,
              validValues: null,
              options: {
                creatable: false,
                trackUnchecked: false,
                sectionHeader: null,
                gridColumns: null,
                labels: {},
                metricCategories: [
                  {
                    label: "Process Metrics",
                    column: 0,
                    metrics: [
                      {
                        name: "process.cpu.time",
                        description: null,
                        kpi: null,
                      },
                      {
                        name: "process.disk.io",
                        description: null,
                        kpi: null,
                      },
                      {
                        name: "process.memory.physical_usage",
                        description: null,
                        kpi: null,
                      },
                      {
                        name: "process.memory.virtual_usage",
                        description: null,
                        kpi: null,
                      },
                    ],
                  },
                ],
              },
            },
            {
              label: "Process Name Filtering",
              name: "enable_process_filter",
              description: "Enable to configure filtering for process metrics.",
              required: false,
              type: "bool",
              default: false,
              documentation: null,
              relevantIf: [
                {
                  name: "enable_process",
                  operator: "equals",
                  value: true,
                },
              ],
              advancedConfig: false,
              validValues: null,
              options: {
                creatable: false,
                trackUnchecked: false,
                sectionHeader: true,
                gridColumns: null,
                labels: {},
                metricCategories: null,
              },
            },
            {
              label: "Process Include Filter",
              name: "process_include",
              description:
                "List of processes to include for metric collection. Defaults to all processes.",
              required: false,
              type: "strings",
              default: [],
              documentation: null,
              relevantIf: [
                {
                  name: "enable_process_filter",
                  operator: "equals",
                  value: true,
                },
              ],
              advancedConfig: false,
              validValues: null,
              options: {
                creatable: false,
                trackUnchecked: false,
                sectionHeader: null,
                gridColumns: null,
                labels: {},
                metricCategories: null,
              },
            },
            {
              label: "Process Exclude Filter",
              name: "process_exclude",
              description:
                "List of processes to exclude from metric collection.",
              required: false,
              type: "strings",
              default: [],
              documentation: null,
              relevantIf: [
                {
                  name: "enable_process_filter",
                  operator: "equals",
                  value: true,
                },
              ],
              advancedConfig: false,
              validValues: null,
              options: {
                creatable: false,
                trackUnchecked: false,
                sectionHeader: null,
                gridColumns: null,
                labels: {},
                metricCategories: null,
              },
            },
            {
              label: "Process Filter Match Type",
              name: "process_filter_match_strategy",
              description: "Strategy for matching process names.",
              required: false,
              type: "enum",
              default: "regexp",
              documentation: null,
              relevantIf: [
                {
                  name: "enable_process_filter",
                  operator: "equals",
                  value: true,
                },
              ],
              advancedConfig: false,
              validValues: ["regexp", "strict"],
              options: {
                creatable: false,
                trackUnchecked: false,
                sectionHeader: null,
                gridColumns: null,
                labels: {},
                metricCategories: null,
              },
            },
            {
              label: "Collection Interval",
              name: "collection_interval",
              description: "How often (seconds) to scrape for metrics.",
              required: false,
              type: "int",
              default: 60,
              documentation: null,
              relevantIf: null,
              advancedConfig: true,
              validValues: null,
              options: {
                creatable: false,
                trackUnchecked: false,
                sectionHeader: null,
                gridColumns: null,
                labels: {},
                metricCategories: null,
              },
            },
          ],
        },
      },
    },
  },
};

export const CUSTOM_DESTINATION_TYPE_MOCK: MockedResponse = {
  request: {
    query: GetDestinationWithTypeDocument,
    variables: {
      name: "logging",
    },
  },
  result: {
    data: {
      destinationWithType: {
        destination: {
          metadata: {
            name: "logging",
            version: 0,
            id: "logging",
            labels: {},
          },
          spec: {
            type: "custom",
            parameters: [
              {
                name: "telemetry_types",
                value: ["Metrics", "Logs"],
              },
              {
                name: "configuration",
                value: "logging:",
              },
            ],
            disabled: false,
          },
        },
        destinationType: {
          metadata: {
            id: "ca735962-d5d8-46d1-a225-6759dfa7d427",
            name: "custom",
            version: 2,
            icon: "/icons/destinations/custom.svg",
            description:
              "Insert a custom OpenTelemetry exporter configuration.",
          },
          spec: {
            parameters: [
              {
                label: "Telemetry Types",
                name: "telemetry_types",
                description:
                  "Select which types of telemetry the exporter supports.",
                required: false,
                type: "enums",
                default: [],
                relevantIf: null,
                documentation: null,
                advancedConfig: false,
                validValues: ["Metrics", "Logs", "Traces"],
                options: {
                  multiline: false,
                  creatable: false,
                  trackUnchecked: false,
                  sectionHeader: null,
                  gridColumns: null,
                  labels: {},
                  metricCategories: null,
                },
              },
              {
                label: "Configuration",
                name: "configuration",
                description:
                  "Enter any supported OpenTelemetry exporter and the YAML will be inserted into the configuration.",
                required: true,
                type: "yaml",
                default: null,
                relevantIf: null,
                documentation: [
                  {
                    text: "Exporter Syntax",
                    url: "https://github.com/observIQ/bindplane-agent/blob/main/docs/exporters.md",
                  },
                ],
                advancedConfig: false,
                validValues: null,
                options: {
                  multiline: false,
                  creatable: false,
                  trackUnchecked: false,
                  sectionHeader: null,
                  gridColumns: 12,
                  labels: {},
                  metricCategories: null,
                },
              },
            ],
          },
        },
      },
    },
  },
};

export const LINUX_METRICS_V1_MOCK_RESULT: GetConfigurationQuery["configuration"] =
  {
    metadata: {
      id: "linux-metrics",
      name: "linux-metrics",
      description: "Kubernetes container logs.",
      labels: {
        platform: "linux",
      },
      version: 1,
    },
    agentCount: 0,
    spec: {
      raw: "",
      sources: [],
      destinations: [],
      selector: {
        matchLabels: {
          configuration: "linux-metrics",
        },
      },
    },
    graph: {
      attributes: {
        activeTypeFlags: 2,
      },
      sources: [],
      intermediates: [],
      targets: [],
      edges: [],
    },
  };

export const LINUX_METRICS_V2_MOCK_RESULT: GetConfigurationQuery["configuration"] =
  {
    metadata: {
      id: "linux-metrics",
      name: "linux-metrics",
      description: "Kubernetes container logs.",
      labels: {
        platform: "linux",
      },
      version: 2,
    },
    agentCount: 0,
    spec: {
      raw: "",
      sources: [
        {
          type: "host",
          name: "",
          parameters: [
            {
              name: "metric_filtering",
              value: [],
            },
            {
              name: "enable_process",
              value: true,
            },
            {
              name: "process_metrics_filtering",
              value: [],
            },
            {
              name: "enable_process_filter",
              value: false,
            },
            {
              name: "process_include",
              value: [],
            },
            {
              name: "process_exclude",
              value: [],
            },
            {
              name: "process_filter_match_strategy",
              value: "regexp",
            },
            {
              name: "collection_interval",
              value: 60,
            },
          ],
          processors: null,
          disabled: false,
        },
      ],
      destinations: [
        {
          type: "",
          name: "logging",
          parameters: null,
          processors: null,
          disabled: false,
        },
      ],
      selector: {
        matchLabels: {
          configuration: "linux-metrics",
        },
      },
    },
    graph: {
      attributes: {
        activeTypeFlags: 2,
      },
      sources: [
        {
          id: "source/source0",
          type: "sourceNode",
          label: "host",
          attributes: {
            activeTypeFlags: 2,
            kind: "Source",
            resourceId: "source0",
            supportedTypeFlags: 2,
          },
        },
      ],
      intermediates: [
        {
          id: "source/source0/processors",
          type: "processorNode",
          label: "Processors",
          attributes: {
            activeTypeFlags: 2,
            supportedTypeFlags: 2,
          },
        },
        {
          id: "destination/logging/processors",
          type: "processorNode",
          label: "Processors",
          attributes: {
            activeTypeFlags: 2,
            supportedTypeFlags: 0,
          },
        },
      ],
      targets: [
        {
          id: "destination/logging",
          type: "destinationNode",
          label: "logging",
          attributes: {
            activeTypeFlags: 2,
            isInline: false,
            kind: "Destination",
            resourceId: "logging",
            supportedTypeFlags: 0,
          },
        },
      ],
      edges: [
        {
          id: "source/source0|source/source0/processors",
          source: "source/source0",
          target: "source/source0/processors",
        },
        {
          id: "source/source0/processors|destination/logging/processors",
          source: "source/source0/processors",
          target: "destination/logging/processors",
        },
        {
          id: "destination/logging/processors|destination/logging",
          source: "destination/logging/processors",
          target: "destination/logging",
        },
      ],
    },
  };

export const LINUX_METRICS_V3_MOCK_RESULT: GetConfigurationQuery["configuration"] =
  {
    metadata: {
      id: "linux-metrics",
      name: "linux-metrics",
      description: "Kubernetes container logs.",
      labels: {
        platform: "linux",
      },
      version: 3,
    },
    agentCount: 0,
    spec: {
      raw: "",
      sources: [
        {
          type: "host",
          name: "",
          parameters: [
            {
              name: "metric_filtering",
              value: [],
            },
            {
              name: "enable_process",
              value: false,
            },
            {
              name: "process_metrics_filtering",
              value: [],
            },
            {
              name: "enable_process_filter",
              value: false,
            },
            {
              name: "process_include",
              value: [],
            },
            {
              name: "process_exclude",
              value: [],
            },
            {
              name: "process_filter_match_strategy",
              value: "regexp",
            },
            {
              name: "collection_interval",
              value: 60,
            },
          ],
          processors: null,
          disabled: false,
        },
      ],
      destinations: [
        {
          type: "",
          name: "logging",
          parameters: null,
          processors: null,
          disabled: false,
        },
      ],
      selector: {
        matchLabels: {
          configuration: "linux-metrics",
        },
      },
    },
    graph: {
      attributes: {
        activeTypeFlags: 2,
      },
      sources: [
        {
          id: "source/source0",
          type: "sourceNode",
          label: "host",
          attributes: {
            activeTypeFlags: 2,
            kind: "Source",
            resourceId: "source0",
            supportedTypeFlags: 2,
          },
        },
      ],
      intermediates: [
        {
          id: "source/source0/processors",
          type: "processorNode",
          label: "Processors",
          attributes: {
            activeTypeFlags: 2,
            supportedTypeFlags: 2,
          },
        },
        {
          id: "destination/logging/processors",
          type: "processorNode",
          label: "Processors",
          attributes: {
            activeTypeFlags: 2,
            supportedTypeFlags: 0,
          },
        },
      ],
      targets: [
        {
          id: "destination/logging",
          type: "destinationNode",
          label: "logging",
          attributes: {
            activeTypeFlags: 2,
            isInline: false,
            kind: "Destination",
            resourceId: "logging",
            supportedTypeFlags: 0,
          },
        },
      ],
      edges: [
        {
          id: "source/source0|source/source0/processors",
          source: "source/source0",
          target: "source/source0/processors",
        },
        {
          id: "source/source0/processors|destination/logging/processors",
          source: "source/source0/processors",
          target: "destination/logging/processors",
        },
        {
          id: "destination/logging/processors|destination/logging",
          source: "destination/logging/processors",
          target: "destination/logging",
        },
      ],
    },
  };

export const LINUX_METRICS_V4_MOCK_RESULT: GetConfigurationQuery["configuration"] =
  {
    metadata: {
      id: "linux-metrics",
      name: "linux-metrics",
      description: "Kubernetes container logs.",
      labels: {
        platform: "linux",
      },
      version: 4,
    },
    agentCount: 0,
    spec: {
      raw: "",
      sources: [
        {
          type: "host",
          name: "",
          parameters: [
            {
              name: "metric_filtering",
              value: [
                "system.disk.io",
                "system.disk.io_time",
                "system.disk.merged",
                "system.disk.operation_time",
                "system.disk.operations",
                "system.disk.pending_operations",
                "system.disk.weighted_io_time",
                "system.processes.count",
                "system.processes.created",
                "system.cpu.time",
                "system.cpu.utilization",
                "system.network.conntrack.max",
                "system.filesystem.utilization",
              ],
            },
            {
              name: "enable_process",
              value: false,
            },
            {
              name: "process_metrics_filtering",
              value: [],
            },
            {
              name: "enable_process_filter",
              value: false,
            },
            {
              name: "process_include",
              value: [],
            },
            {
              name: "process_exclude",
              value: [],
            },
            {
              name: "process_filter_match_strategy",
              value: "regexp",
            },
            {
              name: "collection_interval",
              value: 60,
            },
          ],
          processors: null,
          disabled: false,
        },
      ],
      destinations: [
        {
          type: "",
          name: "logging",
          parameters: null,
          processors: null,
          disabled: false,
        },
      ],
      selector: {
        matchLabels: {
          configuration: "linux-metrics",
        },
      },
    },
    graph: {
      attributes: {
        activeTypeFlags: 2,
      },
      sources: [
        {
          id: "source/source0",
          type: "sourceNode",
          label: "host",
          attributes: {
            activeTypeFlags: 2,
            kind: "Source",
            resourceId: "source0",
            supportedTypeFlags: 2,
          },
        },
      ],
      intermediates: [
        {
          id: "source/source0/processors",
          type: "processorNode",
          label: "Processors",
          attributes: {
            activeTypeFlags: 2,
            supportedTypeFlags: 2,
          },
        },
        {
          id: "destination/logging/processors",
          type: "processorNode",
          label: "Processors",
          attributes: {
            activeTypeFlags: 2,
            supportedTypeFlags: 0,
          },
        },
      ],
      targets: [
        {
          id: "destination/logging",
          type: "destinationNode",
          label: "logging",
          attributes: {
            activeTypeFlags: 2,
            isInline: false,
            kind: "Destination",
            resourceId: "logging",
            supportedTypeFlags: 0,
          },
        },
      ],
      edges: [
        {
          id: "source/source0|source/source0/processors",
          source: "source/source0",
          target: "source/source0/processors",
        },
        {
          id: "source/source0/processors|destination/logging/processors",
          source: "source/source0/processors",
          target: "destination/logging/processors",
        },
        {
          id: "destination/logging/processors|destination/logging",
          source: "destination/logging/processors",
          target: "destination/logging",
        },
      ],
    },
  };

export const ALL_CONFIG_MOCKS: MockedResponse[] = [
  {
    request: {
      query: GetConfigurationDocument,
      variables: {
        name: "linux-metrics:1",
      },
    },
    result: {
      data: {
        configuration: LINUX_METRICS_V1_MOCK_RESULT,
      },
    },
  },
  {
    request: {
      query: GetConfigurationDocument,
      variables: {
        name: "linux-metrics:2",
      },
    },
    result: {
      data: {
        configuration: LINUX_METRICS_V2_MOCK_RESULT,
      },
    },
  },
  {
    request: {
      query: GetConfigurationDocument,
      variables: {
        name: "linux-metrics:3",
      },
    },
    result: {
      data: {
        configuration: LINUX_METRICS_V3_MOCK_RESULT,
      },
    },
  },
  {
    request: {
      query: GetConfigurationDocument,
      variables: {
        name: "linux-metrics:4",
      },
    },
    result: {
      data: {
        configuration: LINUX_METRICS_V4_MOCK_RESULT,
      },
    },
  },
];
