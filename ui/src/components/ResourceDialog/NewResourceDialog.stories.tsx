import { Meta, StoryFn } from "@storybook/react";
import { NewResourceDialog } from ".";
import {
  Destination1,
  Destination2,
  ResourceType1,
  ResourceType2,
  SupportsBoth,
  SupportsLogs,
  SupportsMetrics,
} from "../ResourceConfigForm/__test__/dummyResources";

export default {
  title: "Resource Dialog",
  component: NewResourceDialog,
} as Meta<typeof NewResourceDialog>;

const Template: StoryFn<typeof NewResourceDialog> = (args) => (
  <NewResourceDialog {...args} />
);

export const Destination = Template.bind({});
export const DestinationWithExistingResources = Template.bind({});
export const Source = Template.bind({});

Destination.args = {
  open: true,
  resourceTypes: [ResourceType1, ResourceType2],
  title: "Title",
  kind: "destination",
};

DestinationWithExistingResources.args = {
  open: true,
  resourceTypes: [ResourceType1, ResourceType2],
  resources: [Destination1, Destination2],
  title: "Title",
  kind: "destination",
};

Source.args = {
  open: true,
  resourceTypes: [SupportsLogs, SupportsMetrics, SupportsBoth],
  resources: [],
  title: "Title",
  kind: "source",
};
