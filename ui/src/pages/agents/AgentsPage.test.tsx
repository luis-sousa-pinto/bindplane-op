import { MockedProvider, MockedResponse } from "@apollo/client/testing";
import { render, screen } from "@testing-library/react";
import { SnackbarProvider } from "notistack";
import { MemoryRouter } from "react-router-dom";
import {
  AgentChangesDocument,
  AgentsTableDocument,
  AgentsTableMetricsDocument,
  Role,
} from "../../graphql/generated";
import { AgentsPageContent } from ".";
import { RBACContext } from "../../contexts/RBAC";

const INSTALL_AGENT_BUTTON = "Install Agent";

const MOCKS: MockedResponse[] = [
  {
    request: {
      query: AgentsTableMetricsDocument,
      variables: {
        period: "10s",
      },
    },
    result: {
      data: {
        agentMetrics: [],
      },
    },
  },
  {
    request: {
      query: AgentsTableDocument,
      variables: {
        selector: undefined,
        query: "",
      },
    },
    result: {
      data: {
        agents: {
          agents: [],
          query: "",
          suggestions: [],
          latestVersion: "",
        },
      },
    },
  },
  {
    request: {
      query: AgentChangesDocument,
      variables: {
        selector: undefined,
        query: "",
      },
    },
    result: {
      data: {
        agentChanges: [],
      },
    },
  },
];

describe("AgentsPage RBAC", () => {
  it("shows the install button when user is not viewer", async () => {
    render(
      <MockedProvider mocks={MOCKS}>
        <MemoryRouter>
          <SnackbarProvider>
            <RBACContext.Provider
              value={{
                role: Role.User,
              }}
            ></RBACContext.Provider>
            <AgentsPageContent />
          </SnackbarProvider>
        </MemoryRouter>
      </MockedProvider>
    );

    await screen.findByText(INSTALL_AGENT_BUTTON);
  });

  it("hides the install button when user is viewer", async () => {
    render(
      <MockedProvider mocks={MOCKS}>
        <MemoryRouter>
          <SnackbarProvider>
            <RBACContext.Provider value={{ role: Role.Viewer }}>
              <AgentsPageContent />
            </RBACContext.Provider>
          </SnackbarProvider>
        </MemoryRouter>
      </MockedProvider>
    );

    // wait for page load
    await screen.findByText("Agents");

    const button = screen.queryByText(INSTALL_AGENT_BUTTON);
    expect(button).not.toBeInTheDocument();
  });
});
