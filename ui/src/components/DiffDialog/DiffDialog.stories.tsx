import { Meta, StoryFn } from "@storybook/react";
import { DiffDialog } from "./DiffDialog";
import { CURRENT_CONFIG_MOCK, NEW_CONFIG_MOCK } from "./__mocks__";

export default {
  title: "Diff Dialog",
  component: DiffDialog,
} as Meta<typeof DiffDialog>;

const Template: StoryFn<typeof DiffDialog> = (args) => {
  return <DiffDialog {...args} open={true} configurationName="linux-metrics" />;
};

export const Default = Template.bind({});
Default.parameters = {
  apolloClient: {
    mocks: [CURRENT_CONFIG_MOCK, NEW_CONFIG_MOCK],
  },
};
