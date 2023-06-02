import { ComponentMeta, ComponentStory } from "@storybook/react";
import { GetRolloutHistoryDocument } from "../../graphql/generated";
import { RolloutHistory } from "./RolloutHistory";
import { mockHistory } from "./__test__/rollout-history-mocks";

export default {
  title: "RolloutHistory",
  component: RolloutHistory,
} as ComponentMeta<typeof RolloutHistory>;

const Template: ComponentStory<typeof RolloutHistory> = (args) => {
  return (
    <div style={{ width: 800 }}>
      <RolloutHistory {...args} />
    </div>
  );
};

export const Default = Template.bind({});
export const Loading = Template.bind({});

Default.args = {
  configurationName: "test",
};
Default.parameters = {
  apolloClient: {
    mocks: [
      {
        request: {
          query: GetRolloutHistoryDocument,
          variables: {
            name: "test",
          },
        },
        result: {
          data: mockHistory,
        },
      },
    ],
  },
};

Loading.args = {
  configurationName: "test",
};
