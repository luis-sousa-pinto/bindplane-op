import { MockedProvider, MockedResponse } from "@apollo/client/testing";
import { render, screen } from "@testing-library/react";
import { GetRolloutHistoryDocument } from "../../graphql/generated";
import { RolloutHistory } from "./RolloutHistory";
import { mockHistory } from "./__test__/rollout-history-mocks";

const MOCKS: MockedResponse[] = [
  {
    request: {
      query: GetRolloutHistoryDocument,
      variables: {
        name: "test",
      },
    },
    result: {
      data: mockHistory,
    },
  },
];

describe("RolloutHistory", () => {
  it("renders", () => {
    render(
      <MockedProvider mocks={MOCKS}>
        <RolloutHistory configurationName="test" />
      </MockedProvider>
    );
  });

  it("renders loading state", async () => {
    render(
      <MockedProvider mocks={MOCKS}>
        <RolloutHistory configurationName="test" />
      </MockedProvider>
    );

    await screen.findByTestId("circular-progress");
  });

  it("displays rollout history messages", async () => {
    render(
      <MockedProvider mocks={MOCKS}>
        <RolloutHistory configurationName="test" />
      </MockedProvider>
    );

    await screen.findByText("Rollout History");
    // These tests need to know the local timezone to pass everywhere
    // because dev happens -5/-4UTC and CI happens -8/-7UTC.

    // Get the local timezone offset in hours from -4UTC, the values in the test data
    //             hours UTC ahead test data  -  hours UTC ahead local
    const offset = 4 - new Date().getTimezoneOffset() / 60;

    // interpolate the local timezone offset into the expected string
    // to make the test pass within timezones close enough that the offset won't change the day
    // Testable hypothesis: these tests will fail in Japan
    await screen.findByText(
      `Version 6 pending rollout on 3/30/2023 at ${(13 + offset)
        .toString()
        .padStart(2, "0")}:09`
    );
    await screen.findByText(
      `Version 5 rollout started on 3/30/2023 at ${(12 + offset)
        .toString()
        .padStart(2, "0")}:55`
    );
    await screen.findByText(
      `Version 4 rollout paused on 3/30/2023 at ${(12 + offset)
        .toString()
        .padStart(2, "0")}:42`
    );

    expect(screen.getByText("Version 3", { exact: false })).toHaveTextContent(
      `Version 3 rollout paused with 2 errors on 3/30/2023 at ${(12 + offset)
        .toString()
        .padStart(2, "0")}:27`
    );

    await screen.findByText(
      `Version 2 completed on 3/30/2023 at ${(12 + offset)
        .toString()
        .padStart(2, "0")}:13`
    );
    await screen.findByText(
      `Version 1 rollout replaced on 3/30/2023 at ${(12 + offset)
        .toString()
        .padStart(2, "0")}:00`
    );
  });
});
