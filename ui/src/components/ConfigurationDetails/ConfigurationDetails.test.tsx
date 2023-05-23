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
import { EditConfigDescriptionDocument } from "../../graphql/generated";

describe("ConfigurationDetails component", () => {
  const configurationName = "linux-metrics";
  it("renders", () => {
    render(
      <MockedProvider mocks={DETAILS_MOCKS}>
        <ConfigurationDetails configurationName={configurationName} />
      </MockedProvider>
    );
  });
  it("shows latest description", async () => {
    render(
      <MockedProvider mocks={DETAILS_MOCKS}>
        <ConfigurationDetails configurationName={configurationName} />
      </MockedProvider>
    );
    await screen.findByText(LATEST_DESCRIPTION_BODY);
  });
  it("shows current version", async () => {
    render(
      <MockedProvider mocks={DETAILS_MOCKS}>
        <ConfigurationDetails configurationName={configurationName} />
      </MockedProvider>
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

    render(
      <MockedProvider mocks={[...DETAILS_MOCKS, editDescriptionMutationMock]}>
        <ConfigurationDetails configurationName={configurationName} />
      </MockedProvider>
    );

    await screen.findByText(LATEST_DESCRIPTION_BODY);
    const editButton = screen.getByRole("button");
    editButton.click();

    const textbox = await screen.findByRole("textbox");
    fireEvent.change(textbox, { target: { value: NEW_DESCRIPTION_BODY } });

    // click away
    await userEvent.click(screen.getByText("linux-metrics"));

    await waitFor(() => {
      expect(mutationCalled).toBe(true);
    });
  });

  it("hides the edit button when disableDescriptionEdit is true", async () => {
    render(
      <MockedProvider mocks={DETAILS_MOCKS}>
        <ConfigurationDetails
          configurationName={configurationName}
          disableDescriptionEdit={true}
        />
      </MockedProvider>
    );
    await screen.findByText(LATEST_DESCRIPTION_BODY);
    expect(screen.queryByRole("button")).not.toBeInTheDocument();
  });
});
