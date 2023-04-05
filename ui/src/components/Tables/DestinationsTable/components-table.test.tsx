import { MockedProvider, MockedResponse } from "@apollo/client/testing";
import { GridRowSelectionModel } from "@mui/x-data-grid";
import { render, screen, waitFor } from "@testing-library/react";
import { GetDestinationTypeDisplayInfoDocument } from "../../../graphql/generated";
import { resourcesFromSelected } from "../../../pages/destinations/DestinationsPage";
import { ResourceKind } from "../../../types/resources";

import {
  Destination1,
  Destination2,
} from "../../ResourceConfigForm/__test__/dummyResources";
import { DestinationsDataGrid } from "./DestinationsDataGrid";

describe("resourcesFromSelected", () => {
  it("Destination|gcp", () => {
    const selected = ["Destination|gcp"];

    const want = [
      {
        kind: ResourceKind.DESTINATION,
        metadata: {
          name: "gcp",
        },
      },
    ];

    const got = resourcesFromSelected(selected);

    expect(got).toEqual(want);
  });
});

const MOCKS: MockedResponse[] = [
  {
    request: {
      query: GetDestinationTypeDisplayInfoDocument,
      variables: {
        name: Destination1.metadata.name,
      },
    },
    result: {
      data: {
        metadata: {
          name: "destination-1-name",
          icon: "",
          displayName: "",
        },
      },
    },
  },
  {
    request: {
      query: GetDestinationTypeDisplayInfoDocument,
      variables: {
        name: Destination2.metadata.name,
      },
    },
    result: {
      data: {
        metadata: {
          name: "destination-2-name",
          icon: "",
          displayName: "",
        },
      },
    },
  },
];

describe("DestinationsDataGrid", () => {
  const destinationData = [Destination1, Destination2];

  it("renders without error", () => {
    render(
      <MockedProvider mocks={MOCKS}>
        <DestinationsDataGrid
          loading={false}
          rows={destinationData}
          setSelectionModel={() => {}}
          disableRowSelectionOnClick
          checkboxSelection
          onEditDestination={() => {}}
        />
      </MockedProvider>
    );
  });

  it("displays destinations", () => {
    render(
      <MockedProvider mocks={MOCKS}>
        <DestinationsDataGrid
          loading={false}
          rows={destinationData}
          setSelectionModel={() => {}}
          disableRowSelectionOnClick
          checkboxSelection
          onEditDestination={() => {}}
        />
      </MockedProvider>
    );

    screen.getByText(Destination1.metadata.name);
    screen.getByText(Destination2.metadata.name);
  });

  it("uses the expected GridRowSelectionModel", () => {
    function onDestinationsSelected(m: GridRowSelectionModel) {
      expect(m).toEqual([
        `Destination|${Destination1.metadata.name}`,
        `Destination|${Destination2.metadata.name}`,
      ]);
    }
    render(
      <MockedProvider mocks={MOCKS}>
        <DestinationsDataGrid
          loading={false}
          rows={destinationData}
          setSelectionModel={onDestinationsSelected}
          disableRowSelectionOnClick
          checkboxSelection
          onEditDestination={() => {}}
        />
      </MockedProvider>
    );

    screen.getByLabelText("Select all rows").click();
  });

  it("calls onEditDestination when destinations are selected", async () => {
    let editCalled: boolean = false;
    function onEditDestination() {
      editCalled = true;
    }
    render(
      <MockedProvider mocks={MOCKS}>
        <DestinationsDataGrid
          loading={false}
          rows={destinationData}
          setSelectionModel={() => {}}
          disableRowSelectionOnClick
          checkboxSelection
          onEditDestination={onEditDestination}
        />
      </MockedProvider>
    );

    screen.getByText(Destination1.metadata.name).click();

    await waitFor(() => expect(editCalled).toEqual(true));
  });
});
