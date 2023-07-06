import nock from "nock";
import { fireEvent, render, screen } from "@testing-library/react";
import { ConfigDetailsMenu } from "./ConfigDetailsMenu";
import { SnackbarProvider } from "notistack";
import { MemoryRouter } from "react-router-dom";
import { MockedProvider, MockedResponse } from "@apollo/client/testing";
import {
  VERSION_MOCK_NO_HISTORY,
  VERSION_MOCK_WITH_HISTORY,
} from "../ConfigurationEditor/__test__/mocks";
import { GetConfigNamesDocument } from "../../graphql/generated";

const Wrapper: React.FC<{ mocks: MockedResponse[] }> = ({
  children,
  mocks,
}) => {
  return (
    <MockedProvider mocks={mocks}>
      <MemoryRouter>
        <SnackbarProvider>{children}</SnackbarProvider>
      </MemoryRouter>
    </MockedProvider>
  );
};

describe("ConfigDetailsMenu", () => {
  it("renders", () => {
    render(
      <Wrapper mocks={[]}>
        <ConfigDetailsMenu configName="linux-metrics" />
      </Wrapper>
    );
  });

  it("shows delete option", async () => {
    render(
      <Wrapper mocks={[VERSION_MOCK_WITH_HISTORY]}>
        <ConfigDetailsMenu configName="linux-metrics" />
      </Wrapper>
    );

    screen.getByTestId("config-menu-button").click();
    await screen.findByText("Delete");
  });

  it("opens the delete dialog when delete is clicked", async () => {
    render(
      <Wrapper mocks={[VERSION_MOCK_WITH_HISTORY]}>
        <ConfigDetailsMenu configName="linux-metrics" />
      </Wrapper>
    );

    screen.getByTestId("config-menu-button").click();
    fireEvent.click(screen.getByText("Delete"));
    await screen.findByText(
      "Are you sure you want to delete this configuration?"
    );
  });

  it("can delete the config", async () => {
    nock("http://localhost")
      .post("/v1/delete", (body) => {
        return true;
      })
      .once()
      .reply(202);

    render(
      <Wrapper mocks={[VERSION_MOCK_WITH_HISTORY]}>
        <ConfigDetailsMenu configName="linux-metrics" />
      </Wrapper>
    );

    screen.getByTestId("config-menu-button").click();
    fireEvent.click(screen.getByText("Delete"));
    await screen.findByText(
      "Are you sure you want to delete this configuration?"
    );

    const deleteBtn = await screen.findByTestId(
      "confirm-delete-dialog-delete-button"
    );
    fireEvent.click(deleteBtn);
  });

  it("shows duplicate option when there is a current version", async () => {
    render(
      <Wrapper mocks={[VERSION_MOCK_WITH_HISTORY]}>
        <ConfigDetailsMenu configName="linux-metrics" />
      </Wrapper>
    );

    fireEvent.click(screen.getByTestId("config-menu-button"));
    await screen.findByText("Duplicate current version");
  });

  it("opens the duplicate dialog when duplicate is clicked", async () => {
    const namesMock: MockedResponse = {
      request: {
        query: GetConfigNamesDocument,
      },
      result: {
        data: {
          configurations: {
            configurations: [
              {
                metadata: {
                  name: "linux-metrics",
                  id: "1",
                  version: 1,
                },
              },
            ],
          },
        },
      },
    };
    render(
      <Wrapper mocks={[VERSION_MOCK_WITH_HISTORY, namesMock]}>
        <ConfigDetailsMenu configName="linux-metrics" />
      </Wrapper>
    );

    fireEvent.click(screen.getByTestId("config-menu-button"));
    const duplicateOption = await screen.findByText(
      "Duplicate current version"
    );
    fireEvent.click(duplicateOption);

    await screen.findByText("Duplicate Configuration");
  });

  it("hides duplicate option when there is no current version", async () => {
    render(
      <Wrapper mocks={[VERSION_MOCK_NO_HISTORY]}>
        <ConfigDetailsMenu configName="linux-metrics" />
      </Wrapper>
    );

    screen.getByTestId("config-menu-button").click();
    expect(screen.queryByText("Duplicate")).not.toBeInTheDocument();
  });
});
