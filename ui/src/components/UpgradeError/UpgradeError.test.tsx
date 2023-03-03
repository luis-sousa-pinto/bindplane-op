import { MockedProvider, MockedResponse } from "@apollo/client/testing";
import { render, screen, waitFor } from "@testing-library/react";
import { UpgradeError } from ".";
import { ClearAgentUpgradeErrorDocument } from "../../graphql/generated";

describe("UpgradeError", () => {
  it("renders", () => {
    render(
      <MockedProvider>
        <UpgradeError
          upgradeError="error"
          agentId="1"
          onClearFailure={() => {}}
          onClearSuccess={() => {}}
        />
      </MockedProvider>
    );
  });

  it("displays nothing when upgradeError is undefined", () => {
    render(
      <MockedProvider>
        <UpgradeError
          upgradeError={undefined}
          agentId="1"
          onClearFailure={() => {}}
          onClearSuccess={() => {}}
        />
      </MockedProvider>
    );

    const found = screen.queryByRole("alert");
    expect(found).toBeNull();
  });

  it("displays when upgradeError is defined", () => {
    render(
      <MockedProvider>
        <UpgradeError
          upgradeError="error"
          agentId="1"
          onClearFailure={() => {}}
          onClearSuccess={() => {}}
        />
      </MockedProvider>
    );

    screen.getByRole("alert");
  });

  it("can clear the error with the mutation", async () => {
    var mutationCalled = false;
    const mockedMutation: MockedResponse = {
      request: {
        query: ClearAgentUpgradeErrorDocument,
        variables: {
          input: {
            agentId: "1",
          },
        },
      },
      result: () => {
        mutationCalled = true;
        return {
          data: {
            clearAgentUpgradeError: true,
          },
        };
      },
    };

    render(
      <MockedProvider mocks={[mockedMutation]}>
        <UpgradeError
          upgradeError="error"
          agentId="1"
          onClearFailure={() => {}}
          onClearSuccess={() => {}}
        />
      </MockedProvider>
    );

    screen.getByRole("button").click();
    await waitFor(() => {
      expect(mutationCalled).toBe(true);
    });
  });
});
