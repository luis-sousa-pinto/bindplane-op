import { fireEvent, render, screen } from "@testing-library/react";
import { InstallPageContent } from "./install";
import { MockedProvider, MockedResponse } from "@apollo/client/testing";
import {
  GetConfigurationNamesDocument,
  GetConfigurationNamesQuery,
} from "../../graphql/generated";
import nock from "nock";

const TEST_CONFIGS: GetConfigurationNamesQuery["configurations"]["configurations"] =
  [
    {
      metadata: {
        id: "config-1",
        name: "config-1",
        version: 1,
        labels: {
          platform: "linux",
          env: "test",
          foo: "bar",
        },
      },
    },
    {
      metadata: {
        id: "config-2",
        name: "config-2",
        version: 1,
        labels: {
          platform: "windows",
          env: "test",
          foo: "bar",
        },
      },
    },
  ];

const listConfigsResponse: MockedResponse = {
  request: {
    query: GetConfigurationNamesDocument,
  },
  result: {
    data: {
      configurations: {
        configurations: TEST_CONFIGS,
      },
    },
  },
};

describe("InstallPageContent", () => {
  it("renders", async () => {
    const scope = nock("http://localhost:80")
      .get("/v1/agent-versions/latest/install-command")
      .query(true)
      .reply(200, { command: "the install command" });

    render(
      <MockedProvider mocks={[listConfigsResponse]}>
        <InstallPageContent />
      </MockedProvider>
    );
    expect(await screen.findByTestId("config-select")).toBeInTheDocument();
    expect(await screen.findByText("the install command")).toBeInTheDocument();
    await screen.findByLabelText("Select Config (optional)");
    fireEvent.change(screen.getByTestId("config-select"), {
      target: { value: "config-1" },
    });

    expect(scope.isDone()).toBe(true);
  });
});
