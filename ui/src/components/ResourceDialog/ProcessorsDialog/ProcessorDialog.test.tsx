import { MockedProvider, MockedResponse } from "@apollo/client/testing";
import {
  fireEvent,
  render,
  Screen,
  screen,
  waitFor,
} from "@testing-library/react";
import { SnackbarProvider } from "notistack";
import {
  GetProcessorTypeDocument,
  GetProcessorTypesDocument,
  ProcessorDialogDestinationTypeDocument,
  ProcessorDialogSourceTypeDocument,
  UpdateProcessorsDocument,
} from "../../../graphql/generated";
import {
  PipelineContext,
  PipelineGraphContextValue,
} from "../../PipelineGraph/PipelineGraphContext";
import { ProcessorDialogComponent } from "./ProcessorDialog";

const DEFAULT_PARAMETER_OPTIONS = {
  creatable: false,
  trackUnchecked: false,
  gridColumns: 6,
  sectionHeader: false,
  multiline: false,
  metricCategories: null,
  labels: null,
  password: null,
};

const CONFIG_NO_PROCESSORS = {
  metadata: {
    id: "test",
    name: "test",
    labels: {
      platform: "macos",
    },
  },
  spec: {
    contentType: "",
    sources: [
      {
        type: "file",
        parameters: [
          {
            name: "file_path",
            value: ["/tmp/log.log"],
          },
          {
            name: "exclude_file_path",
            value: [],
          },
          {
            name: "log_type",
            value: "file",
          },
          {
            name: "parse_format",
            value: "none",
          },
          {
            name: "regex_pattern",
            value: "",
          },
          {
            name: "multiline_line_start_pattern",
            value: "",
          },
          {
            name: "encoding",
            value: "utf-8",
          },
          {
            name: "start_at",
            value: "end",
          },
        ],
        disabled: false,
      },
    ],
    destinations: [
      {
        name: "google-cloud-dest",
        type: "",
        parameters: null,
        disabled: true,
      },
    ],
    selector: {
      matchLabels: {
        configuration: "test",
      },
    },
  },
};

const CONFIG_WITH_PROCESSORS = {
  metadata: {
    id: "test",
    name: "test",
    labels: {
      platform: "macos",
    },
  },
  spec: {
    contentType: "",
    sources: [
      {
        type: "file",
        processors: [
          {
            type: "custom",
            parameters: [
              { name: "telemetry_types", value: [] },
              { name: "configuration", value: "blah" },
            ],
            disabled: false,
          },
        ],
        parameters: [
          {
            name: "file_path",
            value: ["/tmp/log.log"],
          },
          {
            name: "exclude_file_path",
            value: [],
          },
          {
            name: "log_type",
            value: "file",
          },
          {
            name: "parse_format",
            value: "none",
          },
          {
            name: "regex_pattern",
            value: "",
          },
          {
            name: "multiline_line_start_pattern",
            value: "",
          },
          {
            name: "encoding",
            value: "utf-8",
          },
          {
            name: "start_at",
            value: "end",
          },
        ],
        disabled: false,
      },
    ],
    destinations: [
      {
        name: "google-cloud-dest",
        type: "",
        parameters: null,
        disabled: false,
        processors: [
          {
            type: "custom",
            disabled: false,
            parameters: [
              { name: "telemetry_types", value: [] },
              { name: "configuration", value: "blah" },
            ],
          },
        ],
      },
    ],
    selector: {
      matchLabels: {
        configuration: "test",
      },
    },
  },
};

const CUSTOM_PROCESSOR = {
  metadata: {
    id: "7f98849b-e71b-45c6-8cde-31bdfe2b2fc7",
    name: "custom",
    displayName: "Custom",
    description: "Insert a custom OpenTelemetry processor configuration.",
  },
  spec: {
    telemetryTypes: ["metrics", "logs", "traces"],
    version: "0.0.1",
    parameters: [
      {
        name: "telemetry_types",
        label: "Telemetry Types",
        description: "Select which types of telemetry the processor supports.",
        type: "enums",
        validValues: ["Metrics", "Logs", "Traces"],
        relevantIf: null,
        documentation: null,
        advancedConfig: false,
        default: [],
        required: true,
        options: DEFAULT_PARAMETER_OPTIONS,
      },
      {
        name: "configuration",
        default: null,
        relevantIf: null,
        advancedConfig: null,
        validValues: null,
        label: "Configuration",
        description:
          "Enter any supported Processor and the YAML will be inserted into the configuration.",
        required: true,
        type: "yaml",
        options: DEFAULT_PARAMETER_OPTIONS,
        documentation: [
          {
            text: "Processor Syntax",
            url: "https://github.com/observIQ/observiq-otel-collector/blob/main/docs/processors.md",
          },
        ],
      },
    ],
  },
};

const SOURCE_TYPE_MOCK: MockedResponse = {
  request: {
    query: ProcessorDialogSourceTypeDocument,
    variables: {
      name: "file",
    },
  },
  result: {
    data: {
      sourceType: {
        __typename: "SourceType",
        metadata: {
          id: "7f98849b-e71b-45c6-8cde-31bdfe2b2fc4",
          name: "file",
          displayName: "File",
          description: "Reads logs from a file",
        },
        spec: {
          telemetryTypes: ["logs"],
        },
      },
    },
  },
};

const DESTINATION_TYPE_MOCK: MockedResponse = {
  request: {
    query: ProcessorDialogDestinationTypeDocument,
    variables: {
      name: "google-cloud-dest",
    },
  },
  result: {
    data: {
      destinationWithType: {
        destinationType: {
          metadata: {
            id: "7f98849b-e71b-45c6-8cde-31bdfe2b2fc8",
            name: "google",
            displayName: "Google Cloud",
            description: "Google cloud destination",
          },
          spec: {
            telemetryTypes: ["logs", "metrics", "traces"],
          },
        },
      },
    },
  },
};

const PROCESSOR_TYPES_MOCK: MockedResponse = {
  request: {
    query: GetProcessorTypesDocument,
  },
  result: () => {
    return {
      data: {
        processorTypes: [CUSTOM_PROCESSOR],
      },
    };
  },
};

const GET_PROCESSOR_TYPE_MOCK: MockedResponse = {
  request: {
    query: GetProcessorTypeDocument,
    variables: {
      type: "custom",
    },
  },
  result: () => {
    return {
      data: {
        processorType: CUSTOM_PROCESSOR,
      },
    };
  },
};

const DEFAULT_PROVIDER_VALUES: PipelineGraphContextValue = {
  selectedTelemetryType: "logs",
  hoveredSet: [],
  setHoveredNodeAndEdgeSet: () => {},
  refetchConfiguration: () => {},
  configuration: CONFIG_NO_PROCESSORS,
  editProcessors: () => {},
  closeProcessorDialog: () => {},
  editProcessorsInfo: { resourceType: "source", index: 0 },
  editProcessorsOpen: true,
  maxValues: {
    maxMetricValue: 0,
    maxLogValue: 0,
    maxTraceValue: 0,
  },
  setMaxValues: () => {},
};

describe("ProcessorDialogComponent", () => {
  it("renders", async () => {
    render(
      <MockedProvider mocks={[SOURCE_TYPE_MOCK]}>
        <SnackbarProvider>
          <PipelineContext.Provider value={DEFAULT_PROVIDER_VALUES}>
            <ProcessorDialogComponent open={true} processors={[]} />
          </PipelineContext.Provider>
        </SnackbarProvider>
      </MockedProvider>
    );

    await screen.findByText("Source File: Processors");
  });

  it("can add a processor to a source", async () => {
    render(
      <MockedProvider
        mocks={[
          SOURCE_TYPE_MOCK,
          PROCESSOR_TYPES_MOCK,
          GET_PROCESSOR_TYPE_MOCK,
        ]}
      >
        <SnackbarProvider>
          <PipelineContext.Provider value={DEFAULT_PROVIDER_VALUES}>
            <ProcessorDialogComponent open={true} processors={[]} />
          </PipelineContext.Provider>
        </SnackbarProvider>
      </MockedProvider>
    );

    await addCustomProcessorToSource(screen);
  });

  it("can add a processor to a destination", async () => {
    render(
      <MockedProvider
        mocks={[
          DESTINATION_TYPE_MOCK,
          PROCESSOR_TYPES_MOCK,
          GET_PROCESSOR_TYPE_MOCK,
        ]}
      >
        <SnackbarProvider>
          <PipelineContext.Provider
            value={{
              ...DEFAULT_PROVIDER_VALUES,
              editProcessorsInfo: { resourceType: "destination", index: 0 },
            }}
          >
            <ProcessorDialogComponent open={true} processors={[]} />
          </PipelineContext.Provider>
        </SnackbarProvider>
      </MockedProvider>
    );

    await addCustomProcessorToDestination(screen);
  });

  it("Calls the GQL Mutation updateProcessors on Save click", async () => {
    var updateProcessorsCalled = false;

    const mutationMock: MockedResponse = {
      request: {
        query: UpdateProcessorsDocument,
        variables: {
          input: {
            configuration: "test",
            resourceType: "SOURCE",
            resourceIndex: 0,
            processors: [
              {
                type: "custom",
                parameters: [
                  {
                    name: "telemetry_types",
                    value: [],
                  },
                  {
                    name: "configuration",
                    value: "blah",
                  },
                ],
                disabled: false,
              },
            ],
          },
        },
      },
      result: () => {
        updateProcessorsCalled = true;
        return { data: { updateProcessors: null } };
      },
    };

    render(
      <MockedProvider
        mocks={[
          SOURCE_TYPE_MOCK,
          mutationMock,
          PROCESSOR_TYPES_MOCK,
          GET_PROCESSOR_TYPE_MOCK,
        ]}
      >
        <SnackbarProvider>
          <PipelineContext.Provider value={DEFAULT_PROVIDER_VALUES}>
            <ProcessorDialogComponent open={true} processors={[]} />
          </PipelineContext.Provider>
        </SnackbarProvider>
      </MockedProvider>
    );

    await addCustomProcessorToSource(screen);

    screen.getByText("Save").click();
    await waitFor(() => expect(updateProcessorsCalled).toBe(true));
  });

  it("Can edit a source processor", async () => {
    var saveCalled: boolean = false;

    const mutationMock: MockedResponse = {
      request: {
        query: UpdateProcessorsDocument,
        variables: {
          input: {
            configuration: "test",
            resourceType: "SOURCE",
            resourceIndex: 0,
            processors: [
              {
                type: "custom",
                parameters: [
                  {
                    name: "telemetry_types",
                    value: [],
                  },
                  {
                    name: "configuration",
                    value: "edited",
                  },
                ],
                disabled: false,
              },
            ],
          },
        },
      },
      result: () => {
        saveCalled = true;

        return {
          data: {
            updateProcessors: null,
          },
        };
      },
    };
    render(
      <MockedProvider
        mocks={[
          SOURCE_TYPE_MOCK,
          PROCESSOR_TYPES_MOCK,
          GET_PROCESSOR_TYPE_MOCK,
          mutationMock,
        ]}
      >
        <SnackbarProvider>
          <PipelineContext.Provider value={DEFAULT_PROVIDER_VALUES}>
            <ProcessorDialogComponent open={true} processors={[]} />
          </PipelineContext.Provider>
        </SnackbarProvider>
      </MockedProvider>
    );

    await addCustomProcessorToSource(screen);

    const editButton = await screen.findByTestId("edit-processor-0");
    editButton.click();

    await screen.findByText("Edit Processor: Custom");

    // Change the value of the textbox
    fireEvent.change(screen.getByRole("textbox"), {
      target: { value: "edited" },
    });

    // Save it
    screen.getByText("Done").click();

    // Verify we're back on the main view and Custom is present
    await screen.findByText("Source File: Processors");
    screen.getByText("Custom");
    screen.getByText("Save").click();

    await waitFor(() => expect(saveCalled).toBe(true));
  });

  it("can edit a destination processor", async () => {
    var saveCalled: boolean = false;

    const mutationMock: MockedResponse = {
      request: {
        query: UpdateProcessorsDocument,
        variables: {
          input: {
            configuration: "test",
            resourceType: "DESTINATION",
            resourceIndex: 0,
            processors: [
              {
                type: "custom",
                parameters: [
                  {
                    name: "telemetry_types",
                    value: [],
                  },
                  {
                    name: "configuration",
                    value: "edited",
                  },
                ],
                disabled: false,
              },
            ],
          },
        },
      },
      result: () => {
        saveCalled = true;

        return {
          data: {
            updateProcessors: null,
          },
        };
      },
    };
    render(
      <MockedProvider
        mocks={[PROCESSOR_TYPES_MOCK, GET_PROCESSOR_TYPE_MOCK, mutationMock]}
      >
        <SnackbarProvider>
          <PipelineContext.Provider
            value={{
              ...DEFAULT_PROVIDER_VALUES,
              editProcessorsInfo: { resourceType: "destination", index: 0 },
            }}
          >
            <ProcessorDialogComponent open={true} processors={[]} />
          </PipelineContext.Provider>
        </SnackbarProvider>
      </MockedProvider>
    );

    await addCustomProcessorToDestination(screen);

    const editButton = await screen.findByTestId("edit-processor-0");
    editButton.click();

    fireEvent.change(screen.getByRole("textbox"), {
      target: { value: "edited" },
    });

    screen.getByText("Done").click();

    await screen.findByText("Destination google-cloud-dest: Processors");
    screen.getByText("Save").click();

    await waitFor(() => expect(saveCalled).toBe(true));
  });

  it("can delete a source processor", async () => {
    var saveCalled: boolean = false;

    const mutationMock: MockedResponse = {
      request: {
        query: UpdateProcessorsDocument,
        variables: {
          input: {
            configuration: "test",
            resourceType: "SOURCE",
            resourceIndex: 0,
            processors: [],
          },
        },
      },
      result: () => {
        saveCalled = true;

        return {
          data: {
            updateProcessors: null,
          },
        };
      },
    };
    render(
      <MockedProvider
        mocks={[
          SOURCE_TYPE_MOCK,
          PROCESSOR_TYPES_MOCK,
          GET_PROCESSOR_TYPE_MOCK,
          mutationMock,
        ]}
      >
        <SnackbarProvider>
          <PipelineContext.Provider
            value={{
              ...DEFAULT_PROVIDER_VALUES,
              configuration: CONFIG_WITH_PROCESSORS,
            }}
          >
            <ProcessorDialogComponent
              open={true}
              processors={[
                {
                  type: "custom",
                  parameters: [
                    { name: "telemetry_types", value: [] },
                    { name: "configuration", value: "blah" },
                  ],
                  disabled: false,
                },
              ]}
            />
          </PipelineContext.Provider>
        </SnackbarProvider>
      </MockedProvider>
    );

    await screen.findByText("Source File: Processors");
    screen.getByTestId("edit-processor-0").click();

    await screen.findByText("Edit Processor: Custom");
    screen.getByText("Delete").click();

    await screen.findByText("Source File: Processors");
    expect(screen.queryByText("Custom")).toBeNull();

    screen.getByText("Save").click();
    await waitFor(() => expect(saveCalled).toBe(true));
  });
  it("can delete a destination processor", async () => {
    var saveCalled: boolean = false;

    const mutationMock: MockedResponse = {
      request: {
        query: UpdateProcessorsDocument,
        variables: {
          input: {
            configuration: "test",
            resourceType: "DESTINATION",
            resourceIndex: 0,
            processors: [],
          },
        },
      },
      result: () => {
        saveCalled = true;

        return {
          data: {
            updateProcessors: null,
          },
        };
      },
    };
    render(
      <MockedProvider
        mocks={[PROCESSOR_TYPES_MOCK, GET_PROCESSOR_TYPE_MOCK, mutationMock]}
      >
        <SnackbarProvider>
          <PipelineContext.Provider
            value={{
              ...DEFAULT_PROVIDER_VALUES,
              configuration: CONFIG_WITH_PROCESSORS,
              editProcessorsInfo: { resourceType: "destination", index: 0 },
            }}
          >
            <ProcessorDialogComponent
              open={true}
              processors={[
                {
                  type: "custom",
                  disabled: false,
                  parameters: [
                    { name: "telemetry_types", value: [] },
                    { name: "configuration", value: "blah" },
                  ],
                },
              ]}
            />
          </PipelineContext.Provider>
        </SnackbarProvider>
      </MockedProvider>
    );

    await screen.findByText("Destination google-cloud-dest: Processors");
    screen.getByTestId("edit-processor-0").click();

    await screen.findByText("Edit Processor: Custom");
    screen.getByText("Delete").click();

    await screen.findByText("Destination google-cloud-dest: Processors");
    expect(screen.queryByText("Custom")).toBeNull();

    screen.getByText("Save").click();
    await waitFor(() => expect(saveCalled).toBe(true));
  });
});

/* ---------------------------- Helper functions ---------------------------- */

async function addCustomProcessorToSource(screen: Screen) {
  await screen.findByText("Source File: Processors");
  screen.getByText("Add processor").click();

  // Verify we're on select view
  await screen.findByText("File: Add a processor");
  screen.getByText("Custom").click();

  // Go to the configure view
  await screen.findByText("Add Processor: Custom");
  fireEvent.change(screen.getByRole("textbox"), {
    target: { value: "blah" },
  });

  // save it
  screen.getByText("Done").click();

  // Verify we're back on the main view and Custom is present
  await screen.findByText("Source File: Processors");
  screen.getByText("Custom");
}
async function addCustomProcessorToDestination(screen: Screen) {
  await screen.findByText("Destination google-cloud-dest: Processors");
  screen.getByText("Add processor").click();

  // Verify we're on select view
  await screen.findByText("google-cloud-dest: Add a processor");
  screen.getByText("Custom").click();

  // Go to the configure view
  await screen.findByText("Add Processor: Custom");
  fireEvent.change(screen.getByRole("textbox"), {
    target: { value: "blah" },
  });

  // save it
  screen.getByText("Done").click();

  // verify we're back on the main view and Custom is present
  await screen.findByText("Destination google-cloud-dest: Processors");
  screen.getByText("Custom");
}
