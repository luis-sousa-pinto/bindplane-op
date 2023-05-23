import { ComponentMeta, ComponentStory } from "@storybook/react";
import { RolloutProgressBar } from "./RolloutProgressBar";

export default {
  title: "RolloutProgress",
  component: RolloutProgressBar,
} as ComponentMeta<typeof RolloutProgressBar>;

const Template: ComponentStory<typeof RolloutProgressBar> = (args) => {
  return (
    <div style={{ width: 800 }}>
      <RolloutProgressBar {...args} />
    </div>
  );
};

export const Pending = Template.bind({});
Pending.args = {
  totalCount: 100,
  completedCount: 0,
  rolloutStatus: 0,
  onPause: () => {},
  onResume: () => {},
  onStart: () => {},
};

export const Started = Template.bind({});
Started.args = {
  totalCount: 100,
  completedCount: 50,
  rolloutStatus: 1,
  onPause: () => {},
  onResume: () => {},
  onStart: () => {},
};

export const Paused = Template.bind({});
Paused.args = {
  totalCount: 100,
  completedCount: 50,
  rolloutStatus: 2,
  onPause: () => {},
  onResume: () => {},
  onStart: () => {},
};

export const Errored = Template.bind({});
Errored.args = {
  totalCount: 100,
  completedCount: 50,
  rolloutStatus: 3,
  onPause: () => {},
  onResume: () => {},
  onStart: () => {},
};
