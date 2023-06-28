import { screen, render, fireEvent, waitFor } from "@testing-library/react";
import nock from "nock";
import { SnackbarProvider } from "notistack";
import { MemoryRouter } from "react-router-dom";
import { DEFAULT_RAW_CONFIG, RawConfigWizard } from ".";
import { RawConfigFormValues } from "../../../../types/forms";
import { UpdateStatus } from "../../../../types/resources";
import { ApplyPayload } from "../../../../types/rest";
import { newConfiguration } from "../../../../utils/resources";
import { MockedProvider } from "@apollo/client/testing";

describe("RawConfigForm", () => {
  const initFormValues: RawConfigFormValues = {
    name: "test",
    description: "test-description",
    rawConfig: "raw:",
    platform: "macos",
    secondaryPlatform: "",
    fileName: "",
  };

  it("populates inputs correctly from initialValues", () => {
    const initFormValues: RawConfigFormValues = {
      name: "test",
      description: "test-description",
      rawConfig: "raw:",
      platform: "macos",
      secondaryPlatform: "",
      fileName: "",
    };

    render(
      <MockedProvider mocks={[]}>
        <SnackbarProvider>
          <MemoryRouter>
            <RawConfigWizard
              initialValues={initFormValues}
              onSuccess={() => {}}
            />
          </MemoryRouter>
        </SnackbarProvider>
      </MockedProvider>
    );

    // Correct value for name input
    const nameInput = screen.getByLabelText("Name") as HTMLInputElement;
    expect(nameInput.value).toEqual("test");
    // Correct value for description input
    const descriptionInput = screen.getByLabelText(
      "Description"
    ) as HTMLInputElement;
    expect(descriptionInput.value).toEqual("test-description");

    // A little clunky, but verify that macOS is selected by
    // getting its text.
    expect(screen.getByText("macOS")).toBeInTheDocument();
  });

  it("renders correct copy when fromInput=true", () => {
    render(
      <MockedProvider mocks={[]}>
        <MemoryRouter>
          <SnackbarProvider>
            <RawConfigWizard
              initialValues={initFormValues}
              fromImport={true}
              onSuccess={() => {}}
            />
          </SnackbarProvider>
        </MemoryRouter>
      </MockedProvider>
    );
    // Step one copy
    expect(
      screen.getByText(
        "We've provided some basic details for this configuration, just verify everything looks correct."
      )
    ).toBeInTheDocument();

    screen.getByTestId("step-one-next").click();

    // Step two copy
    expect(
      screen.getByText(
        "This is the OpenTelemetry configuration of the connected agent. If everything looks good, click Save to complete your import."
      )
    ).toBeInTheDocument();

    const uploadButton = screen.queryByTestId("file-input");
    expect(uploadButton).not.toBeInTheDocument();
  });

  it("will block going to step two if fields aren't valid", () => {
    render(
      <MockedProvider mocks={[]}>
        <MemoryRouter>
          <RawConfigWizard onSuccess={() => {}} />
        </MemoryRouter>
      </MockedProvider>
    );

    expect(screen.getByTestId("step-one")).toBeInTheDocument();
    screen.getByText("Next").click();

    expect(screen.getByTestId("step-one")).toBeInTheDocument();
  });

  it("can navigate to step two with valid form values", () => {
    render(
      <MockedProvider mocks={[]}>
        <SnackbarProvider>
          <MemoryRouter>
            <RawConfigWizard onSuccess={() => {}} />
          </MemoryRouter>
        </SnackbarProvider>
      </MockedProvider>
    );

    fireEvent.change(screen.getByLabelText("Name"), {
      target: { value: "test" },
    });

    fireEvent.mouseDown(screen.getByLabelText("Platform"));
    screen.getByText("Windows").click();

    screen.getByText("Next").click();

    expect(screen.getByTestId("step-two")).toBeInTheDocument();
  });
  it("can navigate to step two with valid form values if the platform has a secondary selection", () => {
    render(
      <MockedProvider mocks={[]}>
        <SnackbarProvider>
          <MemoryRouter>
            <RawConfigWizard onSuccess={() => {}} />
          </MemoryRouter>
        </SnackbarProvider>
      </MockedProvider>
    );

    fireEvent.change(screen.getByLabelText("Name"), {
      target: { value: "test" },
    });

    fireEvent.mouseDown(screen.getByLabelText("Platform"));
    screen.getByText("Kubernetes").click();

    const secondaryPlatformSelect = screen.getByTestId(
      "platform-secondary-select-input"
    );

    fireEvent.change(secondaryPlatformSelect, {
      target: { value: "kubernetes-deployment" },
    });

    screen.getByText("Next").click();

    expect(screen.getByTestId("step-two")).toBeInTheDocument();
  });

  it("contains correct doc links", () => {
    render(
      <MockedProvider mocks={[]}>
        <SnackbarProvider>
          <MemoryRouter>
            <RawConfigWizard onSuccess={() => {}} />
          </MemoryRouter>
        </SnackbarProvider>
      </MockedProvider>
    );

    expect(screen.getByText("sample files")).toHaveAttribute(
      "href",
      "https://github.com/observIQ/observiq-otel-collector/tree/main/config/google_cloud_exporter"
    );
    expect(screen.getByText("OpenTelemetry documentation")).toHaveAttribute(
      "href",
      "https://opentelemetry.io/docs/collector/configuration/"
    );
  });

  it("persists form data between steps", () => {
    render(
      <MockedProvider mocks={[]}>
        <SnackbarProvider>
          <MemoryRouter>
            <RawConfigWizard onSuccess={() => {}} />
          </MemoryRouter>
        </SnackbarProvider>
      </MockedProvider>
    );

    fireEvent.change(screen.getByLabelText("Name"), {
      target: { value: "test" },
    });

    fireEvent.mouseDown(screen.getByLabelText("Platform"));
    screen.getByText("Linux").click();

    fireEvent.change(screen.getByLabelText("Description"), {
      target: { value: "This is the description text." },
    });

    screen.getByText("Next").click();
    expect(screen.getByTestId("step-two")).toBeInTheDocument();

    screen.getByText("Back").click();
    expect(screen.getByTestId("step-one")).toBeInTheDocument();

    expect(screen.getByLabelText("Name")).toHaveValue("test");
    expect(screen.getByLabelText("Description")).toHaveValue(
      "This is the description text."
    );
    expect(screen.getByText("Linux")).toBeInTheDocument();
  });

  it("displays the expected default configuration", () => {
    render(
      <MockedProvider mocks={[]}>
        <MemoryRouter>
          <SnackbarProvider>
            <RawConfigWizard onSuccess={() => {}} />
          </SnackbarProvider>
        </MemoryRouter>
      </MockedProvider>
    );

    goToStepTwo();
    const editor = screen.getByTestId("yaml-editor");
    expect(editor).toHaveValue(DEFAULT_RAW_CONFIG);
  });

  it("posts the correct data to /v1/apply", async () => {
    render(
      <MockedProvider>
        <MemoryRouter>
          <SnackbarProvider>
            <RawConfigWizard onSuccess={() => {}} />
          </SnackbarProvider>
        </MemoryRouter>
      </MockedProvider>
    );

    // Rest Mock for POST /apply
    const restScope = nock("http://localhost:80")
      .post("/v1/apply", (body: ApplyPayload) => {
        gotApplyBody = body;
        return true;
      })
      .once()
      .reply(202, {
        updates: [
          {
            resource: { metadata: { name: "test" } },
            status: UpdateStatus.CREATED,
          },
        ],
      });

    let gotApplyBody: ApplyPayload = { resources: [] };

    goToStepTwo();

    const expectConfig = newConfiguration({
      name: "test",
      description: "",
      spec: {
        selector: { matchLabels: { configuration: "test" } },
        raw: "raw-config",
      },
      labels: { platform: "linux" },
    });

    const textarea = screen.getByTestId("yaml-editor");

    fireEvent.change(textarea, {
      target: { value: "raw-config" },
    });

    const saveButton = screen.getByText("Save");
    expect(saveButton).not.toBeDisabled();

    saveButton.click();

    await waitFor(() => {
      return expect(restScope.isDone()).toEqual(true);
    });
    expect(gotApplyBody).toStrictEqual({ resources: [expectConfig] });
  });

  it("can upload a file", async () => {
    render(
      <MockedProvider>
        <MemoryRouter>
          <SnackbarProvider>
            <RawConfigWizard onSuccess={() => {}} />
          </SnackbarProvider>
        </MemoryRouter>
      </MockedProvider>
    );

    goToStepTwo();

    const file: File = new File(["(⌐□_□)"], "raw-config.yaml");

    const fileInput = screen.getByTestId("file-input");
    expect(fileInput).not.toBeVisible();

    fireEvent.change(fileInput, { target: { files: [file] } });

    const fileChip = await screen.findByText("raw-config.yaml");
    expect(fileChip).toBeInTheDocument();

    screen.getByDisplayValue("(⌐□_□)");
  });

  it("calls onSuccess when apply is successful", async () => {
    let onSuccessCalled = false;

    render(
      <MockedProvider>
        <MemoryRouter>
          <SnackbarProvider>
            <RawConfigWizard
              initialValues={initFormValues}
              onSuccess={() => {
                onSuccessCalled = true;
              }}
            />{" "}
          </SnackbarProvider>
        </MemoryRouter>
      </MockedProvider>
    );

    nock("http://localhost:80")
      .post("/v1/apply", (body) => {
        return true;
      })
      .once()
      .reply(202, {
        updates: [
          {
            resource: { metadata: { name: "test" } },
            status: UpdateStatus.CREATED,
          },
        ],
      });

    screen.getByTestId("step-one-next").click();
    screen.getByTestId("save-button").click();

    await waitFor(() => expect(onSuccessCalled).toEqual(true));
  });

  it("calls onSuccess when apply is successful in import mode", async () => {
    let onSuccessCalled = false;

    render(
      <MockedProvider mocks={[]}>
        <MemoryRouter>
          <SnackbarProvider>
            <RawConfigWizard
              initialValues={initFormValues}
              fromImport={true}
              onSuccess={() => {
                onSuccessCalled = true;
              }}
            />
          </SnackbarProvider>
        </MemoryRouter>
      </MockedProvider>
    );

    nock("http://localhost:80")
      .post("/v1/apply", (body) => {
        return true;
      })
      .once()
      .reply(202, {
        updates: [
          {
            resource: { metadata: { name: "test" } },
            status: UpdateStatus.CREATED,
          },
        ],
      });

    nock("http://localhost")
      .patch(`/v1/agents/labels`, (body) => true)
      .once()
      .reply(200, { errors: [] });

    screen.getByTestId("step-one-next").click();
    screen.getByTestId("save-button").click();

    await waitFor(() => expect(onSuccessCalled).toEqual(true));
  });

  it("displays reason when apply returns update status invalid", async () => {
    render(
      <MockedProvider mocks={[]}>
        <MemoryRouter>
          <SnackbarProvider>
            <RawConfigWizard
              initialValues={initFormValues}
              fromImport={true}
              onSuccess={() => {}}
            />
          </SnackbarProvider>
        </MemoryRouter>
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
            resource: { metadata: { name: "test" } },
            status: UpdateStatus.INVALID,
            reason: invalidReasonText,
          },
        ],
      });

    screen.getByTestId("step-one-next").click();
    screen.getByTestId("save-button").click();

    await screen.findByText(invalidReasonText);
  });
});

function goToStepTwo() {
  fireEvent.change(screen.getByLabelText("Name"), {
    target: { value: "test" },
  });

  fireEvent.mouseDown(screen.getByLabelText("Platform"));
  screen.getByText("Linux").click();

  screen.getByText("Next").click();
  expect(screen.getByTestId("step-two")).toBeInTheDocument();
}
