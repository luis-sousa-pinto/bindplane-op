import { render, screen, waitFor } from "@testing-library/react";
import nock from "nock";
import { OtelConfigEditor } from "./OtelConfigEditor";
import { UpdateStatus } from "../../types/resources";
import { MockedProvider } from "@apollo/client/testing";
import { GetConfigurationDocument } from "../../graphql/generated";

describe("OtelConfigEditor", () => {
  it("displays invalid reason when receiving status invalid after save", async () => {
    var queryCalled = false;
    render(
      <MockedProvider
        mocks={[
          {
            request: {
              query: GetConfigurationDocument,
              variables: {
                name: "raw:1",
              },
            },
            result: () => {
              queryCalled = true;
              return {
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
              };
            },
          },
        ]}
      >
        <OtelConfigEditor configurationName="raw:1" />
      </MockedProvider>
    );

    const invalidReasonText = "REASON_INVALID";

    nock("http://localhost:80")
      .post("/v1/apply", (body) => {
        return true;
      })
      .once()
      .reply(202, {
        updates: [
          {
            resource: {
              metadata: { name: "raw", version: 0, id: "config-id" },
            },
            status: UpdateStatus.INVALID,
            reason: invalidReasonText,
          },
        ],
      });

    await waitFor(() => expect(queryCalled).toBe(true));
    screen.getByTestId("edit-configuration-button").click();
    screen.getByTestId("save-button").click();

    await screen.findByText(invalidReasonText);
  });
});
