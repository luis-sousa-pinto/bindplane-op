import { render, screen, waitFor } from "@testing-library/react";
import { ResourceDestinationCard } from "./ResourceDestinationCard";
import {
  mockedDestationAndTypeResponse_PAUSED,
  mockedDestinationAndTypeResponse,
} from "./__test__/mocks";
import {
  destination0,
  destination0Name,
  testConfig,
} from "./__test__/test-resources";
import nock from "nock";
import { MinimumRequiredConfig } from "../PipelineGraph/PipelineGraph";
import { ApplyPayload } from "../../types/rest";
import { UpdateStatus } from "../../types/resources";
import { Destination } from "../../graphql/generated";
import { Wrapper } from "./__test__/wrapper";

describe("ResourceDestinationCard", () => {
  it("opens the edit dialog", async () => {
    render(
      <Wrapper mocks={[mockedDestinationAndTypeResponse]}>
        <ResourceDestinationCard name={destination0Name} destinationIndex={0} />
      </Wrapper>
    );

    const card = await screen.findByText(destination0Name);
    card.click();

    await screen.findByText(`Edit Destination: ${destination0Name}`);
  });

  it("can remove a destination", async () => {
    // mock the apply call for delete
    nock("http://localhost:80")
      .post("/v1/apply")
      .once()
      .reply(202, (_url, body) => {
        const payload = JSON.parse(body.toString()) as ApplyPayload;
        expect(payload.resources.length).toBe(1);

        const appliedConfig = payload.resources[0] as MinimumRequiredConfig;
        expect(appliedConfig?.spec?.destinations?.length).toBe(0);
        return {
          updates: [
            {
              resource: testConfig,
              status: UpdateStatus.CONFIGURED,
            },
          ],
        };
      });

    render(
      <Wrapper mocks={[mockedDestinationAndTypeResponse]}>
        <ResourceDestinationCard name={destination0Name} destinationIndex={0} />
      </Wrapper>
    );

    const card = await screen.findByText(destination0Name);
    card.click();

    await screen.findByText(`Edit Destination: ${destination0Name}`);

    screen.getByText("Delete").click();
    await screen.findByText(
      "Are you sure you want to remove this destination?"
    );

    screen.getByText("Remove").click();
    await waitFor(() => expect(nock.isDone()).toBe(true));
  });

  it("can pause a destination", async () => {
    nock("http://localhost:80")
      .post("/v1/apply")
      .once()
      .reply(202, (_url, body) => {
        const payload = JSON.parse(body.toString()) as ApplyPayload;
        expect(payload.resources.length).toBe(1);

        const appliedDestination = payload.resources[0] as Destination;
        expect(appliedDestination.kind).toBe("Destination");
        expect(appliedDestination.spec.disabled).toBe(true);
        return {
          updates: [
            {
              resource: destination0,
              status: UpdateStatus.CONFIGURED,
            },
          ],
        };
      });

    render(
      <Wrapper mocks={[mockedDestinationAndTypeResponse]}>
        <ResourceDestinationCard name={destination0Name} destinationIndex={0} />
      </Wrapper>
    );

    const card = await screen.findByText(destination0Name);
    card.click();
    screen.getByText("Pause").click();

    await waitFor(() => expect(nock.isDone()).toBe(true));
  });

  it("can resume a destination", async () => {
    nock("http://localhost:80")
      .post("/v1/apply")
      .once()
      .reply(202, (_url, body) => {
        const payload = JSON.parse(body.toString()) as ApplyPayload;
        expect(payload.resources.length).toBe(1);

        const appliedDestination = payload.resources[0] as Destination;
        expect(appliedDestination.kind).toBe("Destination");
        expect(appliedDestination.spec.disabled).toBe(false);
        return {
          updates: [
            {
              resource: destination0,
              status: UpdateStatus.CONFIGURED,
            },
          ],
        };
      });

    render(
      <Wrapper mocks={[mockedDestationAndTypeResponse_PAUSED]}>
        <ResourceDestinationCard name={destination0Name} destinationIndex={0} />
      </Wrapper>
    );

    const card = await screen.findByText(destination0Name);
    screen.getByText("Paused");
    card.click();

    screen.getByText("Resume").click();

    await waitFor(() => expect(nock.isDone()).toBe(true));
  });
});
