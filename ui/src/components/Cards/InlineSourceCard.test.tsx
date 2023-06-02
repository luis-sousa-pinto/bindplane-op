import { MockedProvider } from "@apollo/client/testing";
import { cleanup, render, screen } from "@testing-library/react";
import { SnackbarProvider } from "notistack";
import { InlineSourceCard } from "./InlineSourceCard";
import { PipelineGraphProvider } from "../PipelineGraph/PipelineGraphContext";

import {
  testConfig,
  redisSourceTypeQuery,
  fileSourceTypeQuery,
} from "./__test__/mocks";

describe("InlineSourceCard", () => {
  afterEach(() => {
    cleanup();
  });

  it("indicates whether the source is paused", async () => {
    render(
      <PipelineGraphProvider
        configuration={testConfig}
        refetchConfiguration={jest.fn()}
        selectedTelemetryType={""}
        addSourceOpen={false}
        setAddSourceOpen={jest.fn()}
        addDestinationOpen={false}
        setAddDestinationOpen={jest.fn()}
        maxValues={{
          maxLogValue: 100,
          maxMetricValue: 100,
          maxTraceValue: 100,
        }}
      >
        <SnackbarProvider>
          <MockedProvider mocks={redisSourceTypeQuery} addTypename={false}>
            <InlineSourceCard
              id="source1"
              configuration={testConfig}
              refetchConfiguration={() => {}}
            />
          </MockedProvider>
        </SnackbarProvider>
      </PipelineGraphProvider>
    );

    const sourceBtn = await screen.findByTestId("source-card-source1");
    expect(sourceBtn).toBeEnabled();
    expect(sourceBtn).toHaveTextContent("Paused");
    expect(sourceBtn).toHaveTextContent("redis display name");
  });
  it("displays file source display name", async () => {
    render(
      <PipelineGraphProvider
        configuration={testConfig}
        refetchConfiguration={jest.fn()}
        selectedTelemetryType={""}
        addSourceOpen={false}
        setAddSourceOpen={jest.fn()}
        addDestinationOpen={false}
        setAddDestinationOpen={jest.fn()}
        maxValues={{
          maxLogValue: 100,
          maxMetricValue: 100,
          maxTraceValue: 100,
        }}
      >
        <SnackbarProvider>
          <MockedProvider mocks={fileSourceTypeQuery} addTypename={false}>
            <InlineSourceCard
              id="source0"
              configuration={testConfig}
              refetchConfiguration={() => {}}
            />
          </MockedProvider>
        </SnackbarProvider>
      </PipelineGraphProvider>
    );

    const sourceBtn = await screen.findByTestId("source-card-source0");
    expect(sourceBtn).toBeEnabled();
    expect(sourceBtn).toHaveTextContent("file display name");
  });
});
