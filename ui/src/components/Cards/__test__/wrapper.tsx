import { SnackbarProvider } from "notistack";
import { MockedProvider, MockedResponse } from "@apollo/client/testing";
import { PipelineContext } from "../../PipelineGraph/PipelineGraphContext";
import { testConfig } from "./test-resources";

export const Wrapper: React.FC<{
  mocks?: MockedResponse[];
}> = ({ children, mocks }) => {
  return (
    <PipelineContext.Provider
      value={{
        configuration: testConfig,
        refetchConfiguration: () => {},
        editProcessorsInfo: null,
        hoveredSet: [],
        setHoveredNodeAndEdgeSet: () => {},
        selectedTelemetryType: "logging",
        editProcessors: () => {},
        editProcessorsOpen: false,
        closeProcessorDialog: () => {},
        maxValues: {
          maxLogValue: 100,
          maxMetricValue: 100,
          maxTraceValue: 100,
        },
        addDestinationOpen: false,
        addSourceOpen: false,
        setAddDestinationOpen: () => {},
        setAddSourceOpen: () => {},
      }}
    >
      <MockedProvider mocks={mocks}>
        <SnackbarProvider>{children}</SnackbarProvider>
      </MockedProvider>
    </PipelineContext.Provider>
  );
};
