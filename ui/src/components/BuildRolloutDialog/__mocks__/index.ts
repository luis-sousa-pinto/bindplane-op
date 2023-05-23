import { MockedResponse } from "@apollo/client/testing";
import { GetRenderedConfigDocument } from "../../../graphql/generated";

export const CURRENT_CONFIG_MOCK: MockedResponse = {
  request: {
    query: GetRenderedConfigDocument,
    variables: {
      name: "linux-metrics:current",
    },
  },
  result: {
    data: {
      configuration: {
        metadata: {
          name: "linux-metrics",
          id: "linux-metrics",
          version: 8,
        },
        rendered:
          "receivers:\n    hostmetrics/source0:\n        collection_interval: 60s\n        scrapers:\n            filesystem:\n                metrics:\n                    system.filesystem.utilization:\n                        enabled: true\n            load:\n                metrics: null\n            memory:\n                metrics:\n                    system.memory.usage:\n                        enabled: false\n                    system.memory.utilization:\n                        enabled: true\n            network:\n                metrics:\n                    system.network.conntrack.count:\n                        enabled: true\n                    system.network.conntrack.max:\n                        enabled: false\n                    system.network.errors:\n                        enabled: false\n            paging:\n                metrics:\n                    system.paging.utilization:\n                        enabled: true\n            process:\n                metrics:\n                    process.context_switches:\n                        enabled: false\n                    process.cpu.utilization:\n                        enabled: false\n                    process.disk.operations:\n                        enabled: false\n                    process.memory.utilization:\n                        enabled: false\n                    process.open_file_descriptors:\n                        enabled: false\n                    process.paging.faults:\n                        enabled: false\n                    process.signals_pending:\n                        enabled: false\n                    process.threads:\n                        enabled: false\n                mute_process_name_error: true\nprocessors:\n    batch/source0__processor0:\n        send_batch_max_size: 0\n        send_batch_size: 8192\n        timeout: 200ms\n    resourcedetection/source0:\n        detectors:\n            - system\n        system:\n            hostname_sources:\n                - os\nexporters:\n    logging/logging: null\nservice:\n    pipelines:\n        metrics/source0__logging:\n            receivers:\n                - hostmetrics/source0\n            processors:\n                - resourcedetection/source0\n                - batch/source0__processor0\n            exporters:\n                - logging/logging\n",
      },
    },
  },
};

export const NEW_CONFIG_MOCK: MockedResponse = {
  request: {
    query: GetRenderedConfigDocument,
    variables: {
      name: "linux-metrics:latest",
    },
  },
  result: {
    data: {
      configuration: {
        metadata: {
          name: "linux-metrics",
          id: "linux-metrics",
          version: 9,
        },
        rendered:
          "receivers:\n    hostmetrics/source0:\n        collection_interval: 60s\n        scrapers:\n            filesystem:\n                metrics:\n                    system.filesystem.utilization:\n                        enabled: true\n            load:\n                metrics: null\n            network:\n                metrics:\n                    system.network.conntrack.count:\n                        enabled: true\n                    system.network.conntrack.max:\n                        enabled: false\n                    system.network.errors:\n                        enabled: false\n            paging:\n                metrics:\n                    system.paging.utilization:\n                        enabled: true\n            process:\n                metrics:\n                    process.context_switches:\n                        enabled: false\n                    process.cpu.utilization:\n                        enabled: false\n                    process.disk.operations:\n                        enabled: false\n                    process.memory.utilization:\n                        enabled: false\n                    process.open_file_descriptors:\n                        enabled: false\n                    process.paging.faults:\n                        enabled: false\n                    process.signals_pending:\n                        enabled: false\n                    process.threads:\n                        enabled: false\n                mute_process_name_error: true\nprocessors:\n    batch/source0__processor0:\n        send_batch_max_size: 0\n        send_batch_size: 8192\n        timeout: 200ms\n    resourcedetection/source0:\n        detectors:\n            - system\n        system:\n            hostname_sources:\n                - os\nexporters:\n    logging/logging: null\nservice:\n    pipelines:\n        metrics/source0__logging:\n            receivers:\n                - hostmetrics/source0\n            processors:\n                - resourcedetection/source0\n                - batch/source0__processor0\n            exporters:\n                - logging/logging\n",
      },
    },
  },
};
