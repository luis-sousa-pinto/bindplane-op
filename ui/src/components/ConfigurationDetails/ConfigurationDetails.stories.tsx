import { ComponentMeta, ComponentStory } from "@storybook/react";
import { ConfigurationDetails } from "./ConfigurationDetails";
import { DETAILS_MOCKS } from "./__test__";

export default {
  title: "Configuration Details",
  Component: ConfigurationDetails,
} as ComponentMeta<typeof ConfigurationDetails>;

const Template: ComponentStory<typeof ConfigurationDetails> = (args) => {
  return (
    <div style={{ width: "1000px" }}>
      <ConfigurationDetails {...args} />
    </div>
  );
};

export const Default = Template.bind({});
Default.args = {
  configurationName: "linux-metrics",
};
Default.parameters = {
  apolloClient: {
    mocks: DETAILS_MOCKS,
  },
};

export const NoEditDescription = Template.bind({});
NoEditDescription.args = {
  configurationName: "linux-metrics",
  disableEdit: true,
};
NoEditDescription.parameters = {
  apolloClient: {
    mocks: DETAILS_MOCKS,
  },
};

export const Loading = Template.bind({});
Loading.args = {
  configurationName: "linux-metrics",
};
