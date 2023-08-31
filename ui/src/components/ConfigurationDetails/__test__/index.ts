import { MockedResponse } from "@apollo/client/testing";
import {
  ConfigurationChangesDocument,
  GetConfigurationDocument,
  GetCurrentConfigVersionDocument,
  GetLatestConfigDescriptionDocument,
} from "../../../graphql/generated";

export const CURRENT_VERSION = 10000;
export const LATEST_VERSION = 10001;
export const LATEST_DESCRIPTION_BODY = "This is the latest config.";
export const NEW_DESCRIPTION_BODY = "This is a new description.";

const currentConfigMock: MockedResponse = {
  request: {
    query: GetCurrentConfigVersionDocument,
    variables: {
      configurationName: "linux-metrics:current",
    },
  },
  result: {
    data: {
      configuration: {
        metadata: {
          name: "linux-metrics",
          id: "linux-metrics-id",
          version: CURRENT_VERSION,
          labels: {
            platform: "linux",
          },
        },
        agentCount: 100,
      },
    },
  },
};

const latestConfigMock: MockedResponse = {
  request: {
    query: GetLatestConfigDescriptionDocument,
    variables: {
      configurationName: "linux-metrics:latest",
    },
  },
  result: {
    data: {
      configuration: {
        metadata: {
          name: "linux-metrics",
          id: "linux-metrics-id",
          version: LATEST_VERSION,
          description: LATEST_DESCRIPTION_BODY,
        },
      },
    },
  },
};

const subscriptionMock: MockedResponse = {
  request: {
    query: ConfigurationChangesDocument,
    variables: {
      query: "name:linux-metrics",
    },
  },
};

const GET_CONFIGURATION_MOCK: MockedResponse = {
  request: {
    query: GetConfigurationDocument,
    variables: {
      name: "linux-metrics",
    },
  },
  result: {
    data: {
      configuration: {
        metadata: {
          id: "01GXXRKHEEM1RBZKR7R9CBHEW8",
          name: "raw",
          description: "",
          labels: {
            platform: "linux",
          },
          version: 1,
        },
        agentCount: 4,
        spec: {
          measurementInterval: "60s",
          raw: "receivers:\n  hostmetrics:\n    collection_interval: 1m\n    scrapers:\n      load:\n      filesystem:\n      memory:\n      network:\n\nprocessors:\n  batch:\n\nexporters:\n  logging:\n    loglevel: error\n\nservice:\n  pipelines:\n    metrics:\n      receivers: [hostmetrics]\n      processors: [batch]\n      exporters: [logging]\n",
          sources: null,
          destinations: null,
          selector: {
            matchLabels: {
              configuration: "raw",
            },
          },
        },
        graph: {
          attributes: {
            activeTypeFlags: 0,
          },
          sources: [],
          intermediates: [],
          targets: [],
          edges: [],
        },
      },
    },
  },
};

export const DETAILS_MOCKS = [
  GET_CONFIGURATION_MOCK,
  currentConfigMock,
  latestConfigMock,
  subscriptionMock,
];
