import { fireEvent, render, screen, waitFor } from "@testing-library/react";
import { SnackbarProvider } from "notistack";
import { MemoryRouter } from "react-router-dom";
import { DuplicateConfigDialog } from "./DuplicateConfigDialog";
import nock from "nock";
import { MockedProvider, MockedResponse } from "@apollo/client/testing";
import { GetConfigNamesDocument } from "../../../graphql/generated";

const NAMES_MOCK: MockedResponse = {
  request: {
    query: GetConfigNamesDocument,
  },
  result: {
    data: {
      configurations: {
        configurations: [
          {
            metadata: {
              name: "config-1",
              id: "123",
              version: 1,
            },
          },
        ],
      },
    },
  },
};

describe("DuplicateConfigDialog", () => {
  it("renders without error", () => {
    render(
      <SnackbarProvider>
        <MockedProvider mocks={[NAMES_MOCK]}>
          <MemoryRouter>
            <DuplicateConfigDialog
              open={true}
              currentConfigName={"current-config-name"}
              onSuccess={() => {}}
            />
          </MemoryRouter>
        </MockedProvider>
      </SnackbarProvider>
    );
  });

  it("disables save button by default", () => {
    render(
      <SnackbarProvider>
        <MockedProvider mocks={[NAMES_MOCK]}>
          <MemoryRouter>
            <DuplicateConfigDialog
              open={true}
              currentConfigName={"current-config-name"}
              onSuccess={() => {}}
            />
          </MemoryRouter>
        </MockedProvider>
      </SnackbarProvider>
    );

    const saveButton = screen.getByText("Save");
    expect(saveButton).toBeDisabled();
  });

  it("enables save button when name is valid", () => {
    render(
      <SnackbarProvider>
        <MockedProvider mocks={[NAMES_MOCK]}>
          <MemoryRouter>
            <DuplicateConfigDialog
              open={true}
              currentConfigName={"current-config-name"}
              onSuccess={() => {}}
            />
          </MemoryRouter>
        </MockedProvider>
      </SnackbarProvider>
    );

    const saveButton = screen.getByText("Save");
    expect(saveButton).toBeDisabled();

    const nameInput = screen.getByRole("textbox");
    fireEvent.change(nameInput, { target: { value: "new-config-name" } });

    nameInput.blur();

    expect(saveButton).not.toBeDisabled();
  });

  it("calls onSuccess when 201 status returns", async () => {
    var onSuccessCalled = false;

    nock("http://localhost")
      .post("/v1/configurations/current-config-name/copy")
      .once()
      .reply(201);

    render(
      <SnackbarProvider>
        <MockedProvider mocks={[NAMES_MOCK]}>
          <MemoryRouter>
            <DuplicateConfigDialog
              open={true}
              currentConfigName={"current-config-name"}
              onSuccess={() => {
                onSuccessCalled = true;
              }}
            />
          </MemoryRouter>
        </MockedProvider>
      </SnackbarProvider>
    );

    const saveButton = screen.getByText("Save");
    expect(saveButton).toBeDisabled();

    const nameInput = screen.getByRole("textbox");
    fireEvent.change(nameInput, { target: { value: "new-config-name" } });

    nameInput.blur();
    fireEvent.click(saveButton);
    await waitFor(() => expect(onSuccessCalled).toBe(true));
  });
});
