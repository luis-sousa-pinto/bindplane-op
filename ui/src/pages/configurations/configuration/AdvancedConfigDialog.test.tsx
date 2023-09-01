import { fireEvent, render, screen, waitFor } from "@testing-library/react";
import { SnackbarProvider } from "notistack";
import nock from "nock";
import { MockedProvider, MockedResponse } from "@apollo/client/testing";
import { GetConfigurationDocument } from "../../../graphql/generated";
import { AdvancedConfigDialog } from "./AdvancedConfigDialog";
import { ApplyPayload } from "../../../types/rest";
import { UpdateStatus } from "../../../types/resources";

const GET_CONFIGURATION_MOCK: MockedResponse = {
  request: {
    query: GetConfigurationDocument,
    variables: {
      name: "config-name",
    },
  },
  result: {
    data: {
      configuration: {
        metadata: {
          id: "01GXXRKHEEM1RBZKR7R9CBHEW8",
          name: "config-name",
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

describe("AdvancedConfigDialog", () => {
  it("renders without error", () => {
    render(
      <SnackbarProvider>
        <MockedProvider mocks={[GET_CONFIGURATION_MOCK]} addTypename={false}>
          <AdvancedConfigDialog
            open={true}
            configName={"config-name"}
            onSuccess={() => {}}
          />
        </MockedProvider>
      </SnackbarProvider>
    );
  });

  it("disables save button by default", () => {
    render(
      <SnackbarProvider>
        <MockedProvider mocks={[GET_CONFIGURATION_MOCK]} addTypename={false}>
          <AdvancedConfigDialog
            open={true}
            configName={"config-name"}
            onSuccess={() => {}}
          />
        </MockedProvider>
      </SnackbarProvider>
    );

    const saveButton = screen.getByText("Save");
    expect(saveButton).toBeDisabled();
  });

  it("enables save button when interval is selected", () => {
    render(
      <SnackbarProvider>
        <MockedProvider mocks={[GET_CONFIGURATION_MOCK]} addTypename={false}>
          <AdvancedConfigDialog
            open={true}
            configName={"config-name"}
            onSuccess={() => {}}
          />
        </MockedProvider>
      </SnackbarProvider>
    );

    const saveButton = screen.getByText("Save");
    expect(saveButton).toBeDisabled();

    const intervalDropdown = screen.getByRole("combobox");
    fireEvent.change(intervalDropdown, { target: { value: "10s" } });

    expect(saveButton).not.toBeDisabled();
  });

  it("calls onSuccess when updateMeasurementInterval is successful", async () => {
    let onSuccessCalled = false;

    nock("http://localhost")
      .post("/v1/apply")
      .once()
      .reply(202, (_url, body) => {
        const payload = JSON.parse(body.toString()) as ApplyPayload;
        expect(payload.resources.length).toBe(1);

        return {
          updates: [
            {
              resource: { metadata: { name: "config-name" } },
              status: UpdateStatus.CONFIGURED,
            },
          ],
        };
      });

    render(
      <SnackbarProvider>
        <MockedProvider
          mocks={[
            GET_CONFIGURATION_MOCK,
            GET_CONFIGURATION_MOCK,
            GET_CONFIGURATION_MOCK,
          ]}
          addTypename={false}
        >
          <AdvancedConfigDialog
            open={true}
            configName={"config-name"}
            onSuccess={() => {
              onSuccessCalled = true;
            }}
          />
        </MockedProvider>
      </SnackbarProvider>
    );

    const saveButton = await screen.findByText("Save");
    const intervalDropdown = await screen.findByRole("combobox");

    fireEvent.change(intervalDropdown, { target: { value: "10s" } });
    fireEvent.click(saveButton);

    // Simulate the async save process here
    await waitFor(() => expect(onSuccessCalled).toBe(true));
  });
});
