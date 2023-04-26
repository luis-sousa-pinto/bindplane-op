import { Meta, StoryFn } from "@storybook/react";
import { AssistedConfigWizard } from ".";
import {
  ResourceType1,
  ResourceType2,
} from "../../../../components/ResourceConfigForm/__test__/dummyResources";
import { SourceTypesDocument } from "../../../../graphql/generated";

export default {
  title: "Assisted Config Wizard",
  component: AssistedConfigWizard,
} as Meta<typeof AssistedConfigWizard>;

const Template: StoryFn<typeof AssistedConfigWizard> = (args) => (
  <AssistedConfigWizard {...args} />
);

export const Default = Template.bind({});

const mockParams = {
  apolloClient: {
    mocks: [
      {
        request: {
          query: SourceTypesDocument,
        },
        result: {
          data: {
            sourceTypes: [ResourceType1, ResourceType2],
          },
        },
      },
    ],
  },
};

Default.args = {};
Default.parameters = mockParams;
