import { fireEvent, render, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { ConfigurationDetails } from "./ConfigurationDetails";
import { MockedProvider, MockedResponse } from "@apollo/client/testing";
import {
  DETAILS_MOCKS,
  LATEST_DESCRIPTION_BODY,
  CURRENT_VERSION,
  LATEST_VERSION,
  NEW_DESCRIPTION_BODY,
} from "./__test__";
import {
  EditConfigDescriptionDocument,
  GetLatestConfigDescriptionDocument,
} from "../../graphql/generated";
import { SnackbarProvider } from "notistack";
import { MemoryRouter } from "react-router-dom";

const Wrapper: React.FC<{ mocks: MockedResponse[] }> = ({
  children,
  mocks,
}) => {
  return (
    <SnackbarProvider>
      <MemoryRouter>
        <MockedProvider mocks={mocks}>{children}</MockedProvider>
      </MemoryRouter>
    </SnackbarProvider>
  );
};

describe("ConfigurationDetails component", () => {
  const configurationName = "linux-metrics";
  it("renders", () => {
    render(
      <Wrapper mocks={DETAILS_MOCKS}>
        <ConfigurationDetails configurationName={configurationName} />
      </Wrapper>
    );
  });
  it("shows latest description", async () => {
    render(
      <Wrapper mocks={DETAILS_MOCKS}>
        <ConfigurationDetails configurationName={configurationName} />
      </Wrapper>
    );
    await screen.findByText(LATEST_DESCRIPTION_BODY);
  });
  it("shows current version", async () => {
    render(
      <Wrapper mocks={DETAILS_MOCKS}>
        <ConfigurationDetails configurationName={configurationName} />
      </Wrapper>
    );
    await screen.findByText(`${CURRENT_VERSION}`);
    expect(screen.queryByText(`${LATEST_VERSION}`)).not.toBeInTheDocument();
  });
  it("can edit the description with the editConfigurationDescription mutation", async () => {
    var mutationCalled = false;
    const editDescriptionMutationMock: MockedResponse = {
      request: {
        query: EditConfigDescriptionDocument,
        variables: {
          input: {
            name: "linux-metrics",
            description: NEW_DESCRIPTION_BODY,
          },
        },
      },
      result: () => {
        mutationCalled = true;
        return {
          data: {
            editConfigurationDescription: null,
          },
        };
      },
    };

    const latestAfterMutationMock: MockedResponse = {
      request: {
        query: GetLatestConfigDescriptionDocument,
        variables: {
          configurationName: "linux-metrics:latest",
        },
      },
      result: {
        data: {
          configuration: {
            metadata: {
              name: "linux-metrics",
              id: "linux-metrics-id",
              version: LATEST_VERSION,
              description: NEW_DESCRIPTION_BODY,
            },
          },
        },
      },
    };

    render(
      <Wrapper
        mocks={[
          ...DETAILS_MOCKS,
          editDescriptionMutationMock,
          latestAfterMutationMock,
        ]}
      >
        <ConfigurationDetails configurationName={configurationName} />
      </Wrapper>
    );

    await screen.findByText(LATEST_DESCRIPTION_BODY);
    const editButton = screen.getByTestId("edit-description-button");
    editButton.click();

    const textbox = await screen.findByRole("textbox");
    fireEvent.change(textbox, { target: { value: NEW_DESCRIPTION_BODY } });

    // click away
    await userEvent.click(screen.getByText("linux-metrics"));

    await waitFor(() => {
      expect(mutationCalled).toBe(true);
    });
  });

  it("hides buttons when disableEdit is true", async () => {
    render(
      <Wrapper mocks={DETAILS_MOCKS}>
        <ConfigurationDetails
          configurationName={configurationName}
          disableEdit={true}
        />
      </Wrapper>
    );
    await screen.findByText(LATEST_DESCRIPTION_BODY);
    expect(
      screen.queryByTestId("edit-description-button")
    ).not.toBeInTheDocument();

    expect(screen.queryByTestId("config-menu-button")).not.toBeInTheDocument();
  });
});
