import { render, screen, waitFor } from "@testing-library/react";
import { ResourceSourceCard } from "./ResourceSourceCard";
import { Wrapper } from "./__test__/wrapper";
import {
  mockedSourceAndTypeResponse,
  mockedSourceAndTypeResponse_PAUSED,
} from "./__test__/mocks";
import { source1, source1Name, testConfig } from "./__test__/test-resources";
import { ApplyPayload } from "../../types/rest";
import { Configuration, Source } from "../../graphql/generated";
import { UpdateStatus } from "../../types/resources";
import nock from "nock";

describe("ResourceSourceCard", () => {
  it("opens the edit dialog", async () => {
    render(
      <Wrapper mocks={[mockedSourceAndTypeResponse]}>
        <ResourceSourceCard name={source1Name} sourceIndex={1} />
      </Wrapper>
    );

    const card = await screen.findByText(source1Name);
    card.click();

    screen.getByText(`Edit Source: ${source1Name}`);
  });

  it("can remove a source", async () => {
    nock("http://localhost:80")
      .post("/v1/apply")
      .once()
      .reply(202, (_url, body) => {
        const payload = JSON.parse(body.toString()) as ApplyPayload;
        expect(payload.resources.length).toBe(1);
        const config = payload.resources[0] as Configuration;
        expect(config?.spec?.sources?.length).toBe(1);

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
      <Wrapper mocks={[mockedSourceAndTypeResponse]}>
        <ResourceSourceCard name={source1Name} sourceIndex={1} />
      </Wrapper>
    );

    const card = await screen.findByText(source1Name);
    card.click();

    const deleteBtn = await screen.findByText("Delete");
    deleteBtn.click();

    await screen.findByText("Are you sure you want to remove this source?");
    const confirmButton = screen.getByText("Remove");
    confirmButton.click();

    await waitFor(() => expect(nock.isDone()).toBe(true));
  });

  it("can pause a source", async () => {
    nock("http://localhost:80")
      .post("/v1/apply")
      .once()
      .reply(202, (_url, body) => {
        const payload = JSON.parse(body.toString()) as ApplyPayload;
        expect(payload.resources.length).toBe(1);
        const appliedSource = payload.resources[0] as Source;

        expect(appliedSource.spec.disabled).toBe(true);

        return {
          updates: [
            {
              resource: source1,
              status: UpdateStatus.CONFIGURED,
            },
          ],
        };
      });

    render(
      <Wrapper mocks={[mockedSourceAndTypeResponse]}>
        <ResourceSourceCard name={source1Name} sourceIndex={1} />
      </Wrapper>
    );

    const card = await screen.findByText(source1Name);
    card.click();

    const pauseBtn = await screen.findByText("Pause");
    pauseBtn.click();

    await waitFor(() => expect(nock.isDone()).toBe(true));
  });

  it("can resume a source", async () => {
    nock("http://localhost:80")
      .post("/v1/apply")
      .once()
      .reply(202, (_url, body) => {
        const payload = JSON.parse(body.toString()) as ApplyPayload;
        expect(payload.resources.length).toBe(1);
        const appliedSource = payload.resources[0] as Source;

        expect(appliedSource.spec.disabled).toBe(false);

        return {
          updates: [
            {
              resource: source1,
              status: UpdateStatus.CONFIGURED,
            },
          ],
        };
      });

    render(
      <Wrapper mocks={[mockedSourceAndTypeResponse_PAUSED]}>
        <ResourceSourceCard name={source1Name} sourceIndex={1} />
      </Wrapper>
    );

    const card = await screen.findByText(source1Name);
    screen.getByText("Paused");
    card.click();

    const resumeBtn = screen.getByText("Resume");
    resumeBtn.click();

    await waitFor(() => expect(nock.isDone()).toBe(true));
  });
});
