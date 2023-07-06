import { GetConfigurationVersionsQuery } from "../../../graphql/generated";

export const NO_VERSION_HISTORY: GetConfigurationVersionsQuery = {
  configurationHistory: [
    {
      metadata: {
        name: "linux-metrics",
        id: "linux-metrics",
        version: 1,
      },
      activeTypes: ["logs", "metrics", "traces"],
      status: {
        latest: true,
        pending: false,
        current: false,
      },
    },
  ],
};

export const NO_VERSION_HISTORY_WITH_PENDING: GetConfigurationVersionsQuery = {
  configurationHistory: [
    {
      metadata: {
        name: "linux-metrics",
        id: "linux-metrics",
        version: 1,
      },
      activeTypes: ["logs", "metrics", "traces"],

      status: {
        latest: true,
        pending: true,
        current: false,
      },
    },
  ],
};

export const NO_VERSION_HISTORY_CURRENT_IS_STABLE: GetConfigurationVersionsQuery =
  {
    configurationHistory: [
      {
        metadata: {
          name: "linux-metrics",
          id: "linux-metrics",
          version: 1,
        },
        activeTypes: ["logs", "metrics", "traces"],
        status: {
          latest: true,
          pending: true,
          current: true,
        },
      },
    ],
  };

export const HISTORY_CURRENT_AND_NEW: GetConfigurationVersionsQuery = {
  configurationHistory: [
    {
      metadata: {
        name: "linux-metrics",
        id: "linux-metrics",
        version: 2,
      },
      activeTypes: ["logs", "metrics", "traces"],
      status: {
        latest: true,
        pending: false,
        current: false,
      },
    },
    {
      metadata: {
        name: "linux-metrics",
        id: "linux-metrics",
        version: 1,
      },
      activeTypes: ["logs", "metrics", "traces"],
      status: {
        latest: false,
        pending: true,
        current: true,
      },
    },
  ],
};

export const HISTORY_LATEST_IS_PENDING: GetConfigurationVersionsQuery = {
  configurationHistory: [
    {
      metadata: {
        name: "linux-metrics",
        id: "linux-metrics",
        version: 3,
      },
      activeTypes: ["logs", "metrics", "traces"],
      status: {
        current: false,
        pending: true,
        latest: true,
      },
    },
    {
      metadata: {
        name: "linux-metrics",
        id: "linux-metrics",
        version: 2,
      },
      activeTypes: ["logs", "metrics", "traces"],
      status: {
        current: true,
        pending: false,
        latest: false,
      },
    },
    {
      metadata: {
        name: "linux-metrics",
        id: "linux-metrics",
        version: 1,
      },
      activeTypes: ["logs", "metrics", "traces"],
      status: {
        current: false,
        pending: false,
        latest: false,
      },
    },
  ],
};

export const HISTORY_LATEST_IS_CURRENT: GetConfigurationVersionsQuery = {
  configurationHistory: [
    {
      metadata: {
        name: "linux-metrics",
        id: "linux-metrics",
        version: 3,
      },
      activeTypes: ["logs", "metrics", "traces"],
      status: {
        current: true,
        pending: true,
        latest: true,
      },
    },
    {
      metadata: {
        name: "linux-metrics",
        id: "linux-metrics",
        version: 2,
      },
      activeTypes: ["logs", "metrics", "traces"],
      status: {
        current: false,
        pending: false,
        latest: false,
      },
    },
    {
      metadata: {
        name: "linux-metrics",
        id: "linux-metrics",
        version: 1,
      },
      activeTypes: ["logs", "metrics", "traces"],
      status: {
        current: false,
        pending: false,
        latest: false,
      },
    },
  ],
};

export const HISTORY_LATEST_IS_NEW: GetConfigurationVersionsQuery = {
  configurationHistory: [
    {
      metadata: {
        name: "linux-metrics",
        id: "linux-metrics",
        version: 3,
      },
      activeTypes: ["logs", "metrics", "traces"],
      status: {
        current: false,
        pending: false,
        latest: true,
      },
    },
    {
      metadata: {
        name: "linux-metrics",
        id: "linux-metrics",
        version: 2,
      },
      activeTypes: ["logs", "metrics", "traces"],
      status: {
        current: true,
        pending: false,
        latest: false,
      },
    },
    {
      metadata: {
        name: "linux-metrics",
        id: "linux-metrics",
        version: 1,
      },
      activeTypes: ["logs", "metrics", "traces"],
      status: {
        current: false,
        pending: false,
        latest: false,
      },
    },
  ],
};

export const HISTORY_LATEST_IS_NEW_WITH_PENDING: GetConfigurationVersionsQuery =
  {
    configurationHistory: [
      {
        metadata: {
          name: "linux-metrics",
          id: "linux-metrics",
          version: 3,
        },
        activeTypes: ["logs", "metrics", "traces"],
        status: {
          current: false,
          pending: false,
          latest: true,
        },
      },
      {
        metadata: {
          name: "linux-metrics",
          id: "linux-metrics",
          version: 2,
        },
        activeTypes: ["logs", "metrics", "traces"],
        status: {
          current: false,
          pending: true,
          latest: false,
        },
      },
      {
        metadata: {
          name: "linux-metrics",
          id: "linux-metrics",
          version: 1,
        },
        activeTypes: ["logs", "metrics", "traces"],
        status: {
          current: true,
          pending: false,
          latest: false,
        },
      },
    ],
  };

export const HISTORY_WITH_PENDING_AND_NEW: GetConfigurationVersionsQuery = {
  configurationHistory: [
    {
      metadata: {
        name: "linux-metrics",
        id: "linux-metrics",
        version: 4,
      },
      activeTypes: ["logs", "metrics", "traces"],
      status: {
        current: false,
        pending: false,
        latest: true,
      },
    },
    {
      metadata: {
        name: "linux-metrics",
        id: "linux-metrics",
        version: 3,
      },
      activeTypes: ["logs", "metrics", "traces"],
      status: {
        current: false,
        pending: true,
        latest: false,
      },
    },
    {
      metadata: {
        name: "linux-metrics",
        id: "linux-metrics",
        version: 2,
      },
      activeTypes: ["logs", "metrics", "traces"],
      status: {
        current: true,
        pending: false,
        latest: false,
      },
    },
    {
      metadata: {
        name: "linux-metrics",
        id: "linux-metrics",
        version: 1,
      },
      activeTypes: ["logs", "metrics", "traces"],
      status: {
        current: false,
        pending: false,
        latest: false,
      },
    },
  ],
};
