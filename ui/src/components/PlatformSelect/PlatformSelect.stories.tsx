import { Meta, StoryFn } from "@storybook/react";
import { PlatformSelect } from ".";

export default {
  title: "Platform Select",
  component: PlatformSelect,
} as Meta<typeof PlatformSelect>;

const Template: StoryFn<typeof PlatformSelect> = (args) => (
  <PlatformSelect {...args} />
);

export const Default = Template.bind({});

Default.args = {
  onPlatformSelected: (v: string) => console.log(v),
};
