import { GetRolloutHistoryQuery } from "../../../graphql/generated";

export const mockHistory: GetRolloutHistoryQuery = {
  configurationHistory: [
    {
      metadata: {
        name: "linux-metrics",
        id: "linux-metrics",
        version: 6,
        dateModified: "2023-03-30T13:09:29.480725-04:00",
      },
      status: {
        rollout: {
          status: 0,
          errors: 0,
        },
      },
    },
    {
      metadata: {
        name: "linux-metrics",
        id: "linux-metrics",
        version: 5,
        dateModified: "2023-03-30T12:55:59.238498-04:00",
      },
      status: {
        rollout: {
          status: 1,
          errors: 0,
        },
      },
    },
    {
      metadata: {
        name: "linux-metrics",
        id: "linux-metrics",
        version: 4,
        dateModified: "2023-03-30T12:42:02.784739-04:00",
      },
      status: {
        rollout: {
          status: 2,
          errors: 0,
        },
      },
    },
    {
      metadata: {
        name: "linux-metrics",
        id: "linux-metrics",
        version: 3,
        dateModified: "2023-03-30T12:27:45.991532-04:00",
      },
      status: {
        rollout: {
          status: 3,
          errors: 2,
        },
      },
    },
    {
      metadata: {
        name: "linux-metrics",
        id: "linux-metrics",
        version: 2,
        dateModified: "2023-03-30T12:13:37.376920-04:00",
      },
      status: {
        rollout: {
          status: 4,
          errors: 0,
        },
      },
    },
    {
      metadata: {
        name: "linux-metrics",
        id: "linux-metrics",
        version: 1,
        dateModified: "2023-03-30T12:00:12.548763-04:00",
      },
      status: {
        rollout: {
          status: 5,
          errors: 0,
        },
      },
    },
  ],
};
