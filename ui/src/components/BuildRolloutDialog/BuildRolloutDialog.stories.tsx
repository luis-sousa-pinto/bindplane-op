import { ComponentMeta, ComponentStory } from "@storybook/react";
import { BuildRolloutDialog } from "./BuildRolloutDialog";
import { CURRENT_CONFIG_MOCK, NEW_CONFIG_MOCK } from "./__mocks__";

export default {
  title: "Build Rollout Dialog",
  component: BuildRolloutDialog,
} as ComponentMeta<typeof BuildRolloutDialog>;

const Template: ComponentStory<typeof BuildRolloutDialog> = (args) => {
  return (
    <BuildRolloutDialog
      {...args}
      open={true}
      configurationName="linux-metrics"
    />
  );
};

export const Default = Template.bind({});
Default.parameters = {
  apolloClient: {
    mocks: [CURRENT_CONFIG_MOCK, NEW_CONFIG_MOCK],
  },
};
