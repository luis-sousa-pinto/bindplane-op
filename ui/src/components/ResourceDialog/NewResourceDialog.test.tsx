import { render, screen } from "@testing-library/react";
import { NewResourceDialog } from ".";
import {
  Destination1,
  ResourceType1,
  ResourceType2,
  WindowsOnlyResourceType,
} from "../ResourceConfigForm/__test__/dummyResources";
import { SelectView } from "./SelectView";

describe("ResourceDialog", () => {
  it("renders without error", () => {
    render(
      <NewResourceDialog
        platform="linux"
        onClose={() => {}}
        resourceTypes={[ResourceType1, ResourceType2]}
        title={""}
        kind={"source"}
        open={true}
      />
    );
  });

  it("renders ResourceTypes", () => {
    render(
      <NewResourceDialog
        platform="linux"
        onClose={() => {}}
        resourceTypes={[ResourceType1, ResourceType2]}
        title={""}
        kind={"source"}
        open={true}
      />
    );

    screen.getByText(ResourceType1.metadata.displayName!);
    screen.getByText(ResourceType2.metadata.displayName!);
  });

  it("displays ResourceType form when clicking next", () => {
    render(
      <NewResourceDialog
        platform="linux"
        onClose={() => {}}
        resourceTypes={[ResourceType1, ResourceType2]}
        title={""}
        kind={"source"}
        open={true}
      />
    );

    screen.getByText("ResourceType One").click();
    screen.getByTestId("resource-form");
  });

  it("will offer to use an existing destination", () => {
    render(
      <NewResourceDialog
        platform="linux"
        onClose={() => {}}
        resourceTypes={[ResourceType1, ResourceType2]}
        resources={[Destination1]}
        title={""}
        kind={"destination"}
        open={true}
      />
    );

    screen.getByText(ResourceType1.metadata.displayName!).click();
    screen.getByText(Destination1.metadata.name);
    screen.getByText("Create New");
  });

  it("can still create a new destination with existing of the same type", async () => {
    render(
      <NewResourceDialog
        platform="linux"
        onClose={() => {}}
        resourceTypes={[ResourceType1, ResourceType2]}
        resources={[Destination1]}
        title={""}
        kind={"destination"}
        open={true}
      />
    );
    screen.getByText(ResourceType1.metadata.displayName!).click();
    screen.getByText(Destination1.metadata.name);
    screen.getByText("Create New").click();

    // Look for the Name input
    screen.getByTestId("name-field");
  });
});

describe("SelectView", () => {
  it("shows all resource types for supported platform", () => {
    render(
      <SelectView
        resourceTypes={[ResourceType1, ResourceType2, WindowsOnlyResourceType]}
        resources={[]}
        setSelected={() => {}}
        setCreateNew={() => {}}
        kind={"source"}
        platform={"windows"}
      />
    );

    screen.getByText(WindowsOnlyResourceType.metadata.displayName!);
    screen.getByText(ResourceType1.metadata.displayName!);
    screen.getByText(ResourceType2.metadata.displayName!);
  });

  it("filters out resource types based on platform", () => {
    render(
      <SelectView
        resourceTypes={[ResourceType1, ResourceType2, WindowsOnlyResourceType]}
        resources={[]}
        setSelected={() => {}}
        setCreateNew={() => {}}
        kind={"source"}
        platform={"linux"}
      />
    );

    expect(
      screen.queryByText(WindowsOnlyResourceType.metadata.displayName!)
    ).not.toBeInTheDocument();
    screen.getByText(ResourceType1.metadata.displayName!);
    screen.getByText(ResourceType2.metadata.displayName!);
  });
});
