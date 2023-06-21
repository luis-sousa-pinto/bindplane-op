import { MemoryRouter } from "react-router-dom";
import {
  ConfigurationChangesDocument,
  GetConfigurationTableDocument,
  GetConfigurationTableQuery,
} from "../../../graphql/generated";
import { MockedProvider, MockedResponse } from "@apollo/client/testing";
import { render, screen } from "@testing-library/react";
import { GridRowSelectionModel } from "@mui/x-data-grid";
import { useState } from "react";
import { ConfigurationsTable } from ".";

const TEST_CONFIGS: GetConfigurationTableQuery["configurations"]["configurations"] =
  [
    {
      metadata: {
        id: "config-1",
        name: "config-1",
        version: 1,
        description: "description for config-1",
        labels: {
          env: "test",
          foo: "bar",
        },
      },
      agentCount: 10,
    },
    {
      metadata: {
        id: "config-2",
        name: "config-2",
        version: 1,
        description: "description for config-2",
        labels: {
          env: "test",
          foo: "bar",
        },
      },
      agentCount: 30,
    },
  ];

const QUERY_RESULT: GetConfigurationTableQuery = {
  configurations: {
    configurations: TEST_CONFIGS,
    query: "",
    suggestions: [],
  },
};

const mocks: MockedResponse<Record<string, any>>[] = [
  {
    request: {
      query: GetConfigurationTableDocument,
      variables: {
        query: "",
        onlyDeployedConfigurations: false,
      },
    },
    result: () => {
      return { data: QUERY_RESULT };
    },
  },
  {
    request: {
      query: GetConfigurationTableDocument,
      variables: {
        query: "",
        onlyDeployedConfigurations: true,
      },
    },
    result: () => {
      return { data: QUERY_RESULT };
    },
  },
  {
    request: {
      query: ConfigurationChangesDocument,
      variables: {
        query: "",
      },
    },
    result: () => {
      return {
        data: { configurationChanges: [] },
      };
    },
  },
];

describe("OverviewConfigurationsTable", () => {
  const Wrapper = () => {
    const [selected, setSelected] = useState<GridRowSelectionModel>([]);
    return (
      <MemoryRouter>
        <MockedProvider mocks={mocks} addTypename={false}>
          <ConfigurationsTable
            allowSelection
            overviewPage={true}
            enableDelete={false}
            setSelected={setSelected}
            selected={selected}
          />
        </MockedProvider>
      </MemoryRouter>
    );
  };
  it("renders rows of configs", async () => {
    render(<Wrapper />);

    const rowOne = await screen.findByText("config-1");
    expect(rowOne).toBeInTheDocument();
    const rowTwo = await screen.findByText("config-2");
    expect(rowTwo).toBeInTheDocument();
  });
  it("does not show delete button after selecting row", async () => {
    render(<Wrapper />);
    // sanity check
    const row1 = await screen.findByText("config-1");
    expect(row1).toBeInTheDocument();
    const checkbox = await screen.findByLabelText("Select all rows");
    checkbox.click();
    expect(() => screen.getByText("Delete 2 Configs")).toThrow();
  });
});

describe("ConfigurationsTable", () => {
  const Wrapper = () => {
    const [selected, setSelected] = useState<GridRowSelectionModel>([]);
    return (
      <MemoryRouter>
        <MockedProvider mocks={mocks} addTypename={false}>
          <ConfigurationsTable
            allowSelection
            overviewPage={false}
            setSelected={setSelected}
            selected={selected}
          />
        </MockedProvider>
      </MemoryRouter>
    );
  };
  it("renders rows of configs", async () => {
    render(<Wrapper />);

    const rowOne = await screen.findByText("config-1");
    expect(rowOne).toBeInTheDocument();
    const rowTwo = await screen.findByText("config-2");
    expect(rowTwo).toBeInTheDocument();
  });
  it("shows delete button after selecting row", async () => {
    render(<Wrapper />);
    // sanity check
    const row1 = await screen.findByText("config-1");
    expect(row1).toBeInTheDocument();
    const checkbox = await screen.findByLabelText("Select all rows");
    checkbox.click();
    const deleteButton = await screen.findByText("Delete 2 Configs");
    expect(deleteButton).toBeInTheDocument();
  });
  it("opens the delete dialog after clicking delete", async () => {
    render(<Wrapper />);
    const row1 = await screen.findByText("config-1");
    expect(row1).toBeInTheDocument();
    const checkbox = await screen.findByLabelText("Select all rows");
    checkbox.click();
    const deleteButton = await screen.findByText("Delete 2 Configs");
    deleteButton.click();
    const dialog = await screen.findByTestId("delete-dialog");
    expect(dialog).toBeInTheDocument();
  });
});
