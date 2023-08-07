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
  GetProcessorTypesQuery,
  GetProcessorWithTypeDocument,
  GetProcessorWithTypeQuery,
  ParameterType,
  PipelineType,
  Processor,
  ProcessorDialogDestinationTypeDocument,
  ProcessorDialogSourceTypeDocument,
  UpdateProcessorsDocument,
} from "../../../graphql/generated";
import { PipelineContext } from "../../PipelineGraph/PipelineGraphContext";
import { ProcessorDialogComponent } from "./ProcessorDialog";
import { MinimumRequiredConfig } from "../../PipelineGraph/PipelineGraph";
import nock from "nock";
import { ApplyPayload } from "../../../types/rest";
import { UpdateStatus } from "../../../types/resources";

const DEFAULT_PARAMETER_OPTIONS = {
  creatable: false,
  trackUnchecked: false,
  gridColumns: 6,
  sectionHeader: false,
  subHeader: null,
  horizontalDivider: false,
  multiline: false,
  metricCategories: null,
  labels: null,
  password: null,
  sensitive: false,
};

const CONFIG_NO_PROCESSORS = {
  metadata: {
    id: "test",
    name: "test",
    labels: {
      platform: "macos",
    },
    version: 0,
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

const CONFIG_WITH_RESOURCE_PROCESSORS: MinimumRequiredConfig = {
  metadata: {
    id: "test",
    name: "test",
    labels: {
      platform: "macos",
    },
    version: 1,
  },
  spec: {
    sources: [
      {
        type: "file",
        processors: [
          {
            name: "my-custom-processor",
            type: "",
            parameters: [],
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
          {
            name: "my-custom-processor",
            type: "",
            parameters: [],
            disabled: false,
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

const CONFIG_WITH_INLINE_PROCESSORS = {
  metadata: {
    id: "test",
    name: "test",
    labels: {
      platform: "macos",
    },
    version: 1,
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

const CUSTOM_PROCESSOR_TYPE: GetProcessorTypesQuery["processorTypes"][0] = {
  metadata: {
    name: "custom",
    id: "custom-id",
    displayName: "Custom",
    description:
      "Enter any supported Processor and the YAML will be inserted into the configuration. OpenTelemetry processor configuration.",
    version: 0,
    labels: {},
    deprecated: false,
    additionalInfo: null,
  },
  spec: {
    telemetryTypes: [
      PipelineType.Metrics,
      PipelineType.Logs,
      PipelineType.Traces,
    ],
    parameters: [
      {
        name: "telemetry_types",
        label: "Telemetry Types",
        description: "Select which types of telemetry the processor supports.",
        type: ParameterType.Enums,
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
        type: ParameterType.Yaml,
        options: DEFAULT_PARAMETER_OPTIONS,
        documentation: [
          {
            text: "Processor Syntax",
            url: "https://github.com/observIQ/bindplane-agent/blob/main/docs/processors.md",
          },
        ],
      },
    ],
  },
};

const CUSTOM_RESOURCE_PROCESSOR: GetProcessorWithTypeQuery["processorWithType"]["processor"] =
  {
    metadata: {
      name: "my-custom-processor",
      id: "my-custom-processor-id",
      version: 1,
      labels: {},
    },
    spec: {
      disabled: false,
      type: "custom",
      parameters: [
        {
          name: "telemetry_types",
          value: ["Metrics", "Logs", "Traces"],
        },
        {
          name: "configuration",
          value: "yaml: value1",
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
          id: "source-type-id",
          name: "file",
          displayName: "File",
          description: "Reads logs from a file",
          version: 0,
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
            id: "destination-type-id",
            name: "google",
            displayName: "Google Cloud",
            description: "Google cloud destination",
            version: 0,
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
        processorTypes: [CUSTOM_PROCESSOR_TYPE],
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
        processorType: CUSTOM_PROCESSOR_TYPE,
      },
    };
  },
};

const GET_PROCESSOR_WITH_TYPE_MOCK: MockedResponse<GetProcessorWithTypeQuery> =
  {
    request: {
      query: GetProcessorWithTypeDocument,
      variables: {
        name: "my-custom-processor",
      },
    },
    result: () => {
      return {
        data: {
          processorWithType: {
            processor: CUSTOM_RESOURCE_PROCESSOR,
            processorType: CUSTOM_PROCESSOR_TYPE,
          },
        },
      };
    },
  };

describe("ProcessorDialogComponent", () => {
  it("renders", async () => {
    render(
      <MockedProvider mocks={[SOURCE_TYPE_MOCK]}>
        <SnackbarProvider>
          <PipelineContext.Provider
            value={{
              refetchConfiguration: () => {},
              configuration: CONFIG_NO_PROCESSORS,
              selectedTelemetryType: "logs",
              hoveredSet: [],
              setHoveredNodeAndEdgeSet: () => {},
              editProcessors: () => {},
              closeProcessorDialog: () => {},
              editProcessorsInfo: { resourceType: "source", index: 0 },
              editProcessorsOpen: true,
              addDestinationOpen: false,
              addSourceOpen: false,
              setAddSourceOpen: () => {},
              setAddDestinationOpen: () => {},
              maxValues: {
                maxMetricValue: 0,
                maxLogValue: 0,
                maxTraceValue: 0,
              },
            }}
          >
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
          <PipelineContext.Provider
            value={{
              refetchConfiguration: () => {},
              configuration: CONFIG_NO_PROCESSORS,
              selectedTelemetryType: "logs",
              hoveredSet: [],
              setHoveredNodeAndEdgeSet: () => {},
              editProcessors: () => {},
              closeProcessorDialog: () => {},
              editProcessorsInfo: { resourceType: "source", index: 0 },
              editProcessorsOpen: true,
              addDestinationOpen: false,
              addSourceOpen: false,
              setAddSourceOpen: () => {},
              setAddDestinationOpen: () => {},
              maxValues: {
                maxMetricValue: 0,
                maxLogValue: 0,
                maxTraceValue: 0,
              },
            }}
          >
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
              refetchConfiguration: () => {},
              configuration: CONFIG_NO_PROCESSORS,
              selectedTelemetryType: "logs",
              hoveredSet: [],
              setHoveredNodeAndEdgeSet: () => {},
              editProcessors: () => {},
              closeProcessorDialog: () => {},
              editProcessorsInfo: { resourceType: "destination", index: 0 },
              editProcessorsOpen: true,
              addDestinationOpen: false,
              addSourceOpen: false,
              setAddSourceOpen: () => {},
              setAddDestinationOpen: () => {},
              maxValues: {
                maxMetricValue: 0,
                maxLogValue: 0,
                maxTraceValue: 0,
              },
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
          <PipelineContext.Provider
            value={{
              selectedTelemetryType: "logs",
              hoveredSet: [],
              setHoveredNodeAndEdgeSet: () => {},
              refetchConfiguration: () => {},
              configuration: CONFIG_NO_PROCESSORS,
              editProcessors: () => {},
              closeProcessorDialog: () => {},
              editProcessorsInfo: { resourceType: "source", index: 0 },
              editProcessorsOpen: true,
              addDestinationOpen: false,
              addSourceOpen: false,
              setAddSourceOpen: () => {},
              setAddDestinationOpen: () => {},
              maxValues: {
                maxMetricValue: 0,
                maxLogValue: 0,
                maxTraceValue: 0,
              },
            }}
          >
            <ProcessorDialogComponent open={true} processors={[]} />
          </PipelineContext.Provider>
        </SnackbarProvider>
      </MockedProvider>
    );

    await addCustomProcessorToSource(screen);

    screen.getByText("Save").click();
    await waitFor(() => expect(updateProcessorsCalled).toBe(true));
  });

  it("Can edit an inline source processor", async () => {
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
                displayName: "Awesome Processor",
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
          GET_PROCESSOR_TYPE_MOCK,
          mutationMock,
        ]}
      >
        <SnackbarProvider>
          <PipelineContext.Provider
            value={{
              selectedTelemetryType: "logs",
              hoveredSet: [],
              setHoveredNodeAndEdgeSet: () => {},
              refetchConfiguration: () => {},
              configuration: CONFIG_NO_PROCESSORS,
              editProcessors: () => {},
              closeProcessorDialog: () => {},
              editProcessorsInfo: { resourceType: "source", index: 0 },
              editProcessorsOpen: true,
              addDestinationOpen: false,
              addSourceOpen: false,
              setAddSourceOpen: () => {},
              setAddDestinationOpen: () => {},
              maxValues: {
                maxMetricValue: 0,
                maxLogValue: 0,
                maxTraceValue: 0,
              },
            }}
          >
            <ProcessorDialogComponent open={true} processors={[]} />
          </PipelineContext.Provider>
        </SnackbarProvider>
      </MockedProvider>
    );

    await addCustomProcessorToSource(screen);

    const editButton = await screen.findByTestId("edit-processor-0");
    editButton.click();

    await screen.findByText("Custom");

    // Change the value of the textbox
    fireEvent.change(screen.getByTestId("yaml-editor"), {
      target: { value: "edited" },
    });

    // Change the Short Description
    fireEvent.change(screen.getByLabelText("Short Description"), {
      target: { value: "Awesome Processor" },
    });

    // Save it
    screen.getByText("Done").click();

    // Verify we're back on the main view and Custom is present
    await screen.findByText("Source File: Processors");
    screen.getByText("Awesome Processor");
    screen.getByText("Custom:");
    screen.getByText("Save").click();

    await waitFor(() => expect(saveCalled).toBe(true));
  });

  it("can edit a resource source processor", async () => {
    nock("http://localhost:80")
      .post("/v1/apply")
      .once()
      .reply(202, (_url, body) => {
        const payload = JSON.parse(body.toString()) as ApplyPayload;
        expect(payload.resources.length).toBe(1);

        const payloadProcessor = payload.resources[0] as Processor;

        const editedField = payloadProcessor.spec.parameters!.find(
          (p) => p.name === "configuration"
        );
        expect(editedField?.value).toBe("edited");

        return {
          updates: [
            {
              resource: {},
              status: UpdateStatus.CONFIGURED,
            },
          ],
        };
      });

    render(
      <MockedProvider
        mocks={[
          PROCESSOR_TYPES_MOCK,
          GET_PROCESSOR_WITH_TYPE_MOCK,
          GET_PROCESSOR_WITH_TYPE_MOCK,
          SOURCE_TYPE_MOCK,
        ]}
      >
        <SnackbarProvider>
          <PipelineContext.Provider
            value={{
              selectedTelemetryType: "logs",
              hoveredSet: [],
              setHoveredNodeAndEdgeSet: () => {},
              refetchConfiguration: () => {},
              configuration: CONFIG_WITH_RESOURCE_PROCESSORS,
              editProcessors: () => {},
              closeProcessorDialog: () => {},
              editProcessorsInfo: { resourceType: "source", index: 0 },
              editProcessorsOpen: true,
              addDestinationOpen: false,
              addSourceOpen: false,
              setAddSourceOpen: () => {},
              setAddDestinationOpen: () => {},
              maxValues: {
                maxMetricValue: 0,
                maxLogValue: 0,
                maxTraceValue: 0,
              },
            }}
          >
            <ProcessorDialogComponent
              open={true}
              processors={[{ name: "my-custom-processor", disabled: false }]}
            />
          </PipelineContext.Provider>
        </SnackbarProvider>
      </MockedProvider>
    );

    await screen.findByText("Custom:");
    await screen.findByText("my-custom-processor");

    screen.getByTestId("edit-processor-0").click();

    // Edit screen
    await screen.findByText("Custom");

    // Change the value of the textbox
    fireEvent.change(screen.getByTestId("yaml-editor"), {
      target: { value: "edited" },
    });

    // Save it
    screen.getByText("Done").click();
    const saveBtn = await screen.findByText("Save");
    saveBtn.click();

    await waitFor(() => expect(nock.isDone()).toBe(true));
  });

  it("can edit a destination inline processor", async () => {
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
                displayName: "Rad Processor",
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
          PROCESSOR_TYPES_MOCK,
          GET_PROCESSOR_TYPE_MOCK,
          GET_PROCESSOR_TYPE_MOCK,
          DESTINATION_TYPE_MOCK,
          mutationMock,
        ]}
      >
        <SnackbarProvider>
          <PipelineContext.Provider
            value={{
              selectedTelemetryType: "logs",
              hoveredSet: [],
              setHoveredNodeAndEdgeSet: () => {},
              refetchConfiguration: () => {},
              configuration: CONFIG_NO_PROCESSORS,
              editProcessors: () => {},
              closeProcessorDialog: () => {},
              editProcessorsInfo: { resourceType: "destination", index: 0 },
              editProcessorsOpen: true,
              addDestinationOpen: false,
              addSourceOpen: false,
              setAddSourceOpen: () => {},
              setAddDestinationOpen: () => {},
              maxValues: {
                maxMetricValue: 0,
                maxLogValue: 0,
                maxTraceValue: 0,
              },
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

    const yamlEditor = await screen.findByTestId("yaml-editor");
    fireEvent.change(yamlEditor, {
      target: { value: "edited" },
    });

    fireEvent.change(screen.getByLabelText("Short Description"), {
      target: { value: "Rad Processor" },
    });

    screen.getByText("Done").click();

    await screen.findByText("Destination google-cloud-dest: Processors");
    screen.getByText("Save").click();

    await waitFor(() => expect(saveCalled).toBe(true));
  });

  it("can edit a resource destination processor", async () => {
    nock("http://localhost:80")
      .post("/v1/apply")
      .once()
      .reply(202, (_url, body) => {
        const payload = JSON.parse(body.toString()) as ApplyPayload;
        expect(payload.resources.length).toBe(1);

        const payloadProcessor = payload.resources[0] as Processor;

        const editedField = payloadProcessor.spec.parameters!.find(
          (p) => p.name === "configuration"
        );
        expect(editedField?.value).toBe("edited");

        return {
          updates: [
            {
              resource: {},
              status: UpdateStatus.CONFIGURED,
            },
          ],
        };
      });

    render(
      <MockedProvider
        mocks={[
          PROCESSOR_TYPES_MOCK,
          GET_PROCESSOR_WITH_TYPE_MOCK,
          GET_PROCESSOR_WITH_TYPE_MOCK,
          DESTINATION_TYPE_MOCK,
        ]}
      >
        <SnackbarProvider>
          <PipelineContext.Provider
            value={{
              selectedTelemetryType: "logs",
              hoveredSet: [],
              setHoveredNodeAndEdgeSet: () => {},
              refetchConfiguration: () => {},
              configuration: CONFIG_WITH_RESOURCE_PROCESSORS,
              editProcessors: () => {},
              closeProcessorDialog: () => {},
              editProcessorsInfo: { resourceType: "destination", index: 0 },
              editProcessorsOpen: true,
              addDestinationOpen: false,
              addSourceOpen: false,
              setAddSourceOpen: () => {},
              setAddDestinationOpen: () => {},
              maxValues: {
                maxMetricValue: 0,
                maxLogValue: 0,
                maxTraceValue: 0,
              },
            }}
          >
            <ProcessorDialogComponent
              open={true}
              processors={[{ name: "my-custom-processor", disabled: false }]}
            />
          </PipelineContext.Provider>
        </SnackbarProvider>
      </MockedProvider>
    );

    await screen.findByText("my-custom-processor");
    screen.getByTestId("edit-processor-0").click();

    // Edit screen
    await screen.findByText("Custom");

    // Change the value of the textbox
    fireEvent.change(screen.getByTestId("yaml-editor"), {
      target: { value: "edited" },
    });

    // Save it
    screen.getByText("Done").click();
    const saveBtn = await screen.findByText("Save");
    saveBtn.click();

    await waitFor(() => expect(nock.isDone()).toBe(true));
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
          GET_PROCESSOR_TYPE_MOCK,
          mutationMock,
        ]}
      >
        <SnackbarProvider>
          <PipelineContext.Provider
            value={{
              refetchConfiguration: () => {},
              selectedTelemetryType: "logs",
              hoveredSet: [],
              setHoveredNodeAndEdgeSet: () => {},
              configuration: CONFIG_WITH_INLINE_PROCESSORS,
              editProcessors: () => {},
              closeProcessorDialog: () => {},
              editProcessorsInfo: { resourceType: "source", index: 0 },
              editProcessorsOpen: true,
              addDestinationOpen: false,
              addSourceOpen: false,
              setAddSourceOpen: () => {},
              setAddDestinationOpen: () => {},
              maxValues: {
                maxMetricValue: 0,
                maxLogValue: 0,
                maxTraceValue: 0,
              },
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

    await screen.findByText("Custom");
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
        mocks={[
          PROCESSOR_TYPES_MOCK,
          GET_PROCESSOR_TYPE_MOCK,
          GET_PROCESSOR_TYPE_MOCK,
          DESTINATION_TYPE_MOCK,
          mutationMock,
        ]}
      >
        <SnackbarProvider>
          <PipelineContext.Provider
            value={{
              refetchConfiguration: () => {},
              selectedTelemetryType: "logs",
              hoveredSet: [],
              setHoveredNodeAndEdgeSet: () => {},
              configuration: CONFIG_WITH_INLINE_PROCESSORS,
              editProcessors: () => {},
              closeProcessorDialog: () => {},
              editProcessorsInfo: { resourceType: "destination", index: 0 },
              editProcessorsOpen: true,
              addDestinationOpen: false,
              addSourceOpen: false,
              setAddSourceOpen: () => {},
              setAddDestinationOpen: () => {},
              maxValues: {
                maxMetricValue: 0,
                maxLogValue: 0,
                maxTraceValue: 0,
              },
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

    await screen.findByText("Custom");
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
  await screen.findByText("Add a processor");
  screen.getByText("Custom").click();

  // Go to the configure view
  await screen.findByText("Custom");
  fireEvent.change(screen.getByTestId("yaml-editor"), {
    target: { value: "blah" },
  });

  // save it
  screen.getByText("Done").click();

  // Verify we're back on the main view and Custom is present
  await screen.findByText("Add processor");
  screen.getByText("Custom");
}
async function addCustomProcessorToDestination(screen: Screen) {
  await screen.findByText("Destination google-cloud-dest: Processors");
  screen.getByText("Add processor").click();

  // Verify we're on select view
  await screen.findByText("Add a processor");
  screen.getByText("Custom").click();

  // Go to the configure view
  await screen.findByText("Custom");
  fireEvent.change(screen.getByTestId("yaml-editor"), {
    target: { value: "blah" },
  });

  // save it
  screen.getByText("Done").click();

  // verify we're back on the main view and Custom is present
  await screen.findByText("Add processor");
  screen.getByText("Custom");
}
