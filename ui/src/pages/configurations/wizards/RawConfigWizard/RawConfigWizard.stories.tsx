import { Meta, StoryFn } from "@storybook/react";
import { RawConfigWizard } from ".";

export default {
  title: "Raw Config Wizard",
  component: RawConfigWizard,
} as Meta<typeof RawConfigWizard>;

const Template: StoryFn<typeof RawConfigWizard> = (args) => (
  <RawConfigWizard {...args} />
);

export const StepOne = Template.bind({});

StepOne.args = {};
