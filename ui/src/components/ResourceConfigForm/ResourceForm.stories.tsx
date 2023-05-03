import { Meta, StoryFn } from "@storybook/react";
import { ResourceConfigForm } from ".";
import { ResourceType1, ResourceType2 } from "./__test__/dummyResources";

export default {
  title: "Resource Form",
  component: ResourceConfigForm,
} as Meta<typeof ResourceConfigForm>;

const Template: StoryFn<typeof ResourceConfigForm> = (args) => (
  <div style={{ width: 400 }}>
    <ResourceConfigForm {...args} />
  </div>
);

export const Default = Template.bind({});
export const RelevantIf = Template.bind({});

Default.args = {
  displayName: ResourceType1.metadata.displayName!,
  description: ResourceType1.metadata.description!,
  kind: "source",
  parameterDefinitions: ResourceType1.spec.parameters,
};
RelevantIf.args = {
  displayName: ResourceType2.metadata.displayName!,
  description: ResourceType2.metadata.description!,
  kind: "source",
  parameterDefinitions: ResourceType2.spec.parameters,
};
