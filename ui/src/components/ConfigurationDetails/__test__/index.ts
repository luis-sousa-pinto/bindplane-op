import { MockedResponse } from "@apollo/client/testing";
import {
  ConfigurationChangesDocument,
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

export const DETAILS_MOCKS = [
  currentConfigMock,
  latestConfigMock,
  subscriptionMock,
];
