import { Meta, StoryFn } from "@storybook/react";
import { AgentsTable } from ".";
import {
  AgentChangesDocument,
  AgentsTableDocument,
  AgentsTableQuery,
} from "../../../graphql/generated";
import { AgentTable } from "../AgentTable";
import { generateAgents } from "./__testutil__/generate-agents";

export default {
  title: "Agents Table",
  component: AgentsTable,
  argTypes: {
    density: {
      options: ["standard", "comfortable", "compact"],
      control: "radio",
    },
    columnFields: {
      options: [
        "name",
        "status",
        "version",
        "configuration",
        "operatingSystem",
        "labels",
      ],
      control: "multi-select",
    },
  },
} as Meta<typeof AgentTable>;

const Template: StoryFn<typeof AgentsTable> = (args) => (
  <div style={{ width: "80vw", height: "500px" }}>
    <AgentsTable {...args} />
  </div>
);

export const Default = Template.bind({});
export const Selectable = Template.bind({});

const resultData: AgentsTableQuery = {
  agents: {
    agents: generateAgents(50),
    suggestions: [],
    query: "",
    latestVersion: "v1.5.0",
  },
};

const mockParams = {
  apolloClient: {
    mocks: [
      {
        request: {
          query: AgentsTableDocument,
          variables: {
            query: "",
          },
        },
        result: {
          data: resultData,
        },
      },
      {
        request: {
          query: AgentChangesDocument,
          variables: {
            query: "",
          },
        },
        result: {
          data: {
            agentChanges: [],
          },
        },
      },
    ],
  },
};

Default.args = {};
Default.parameters = mockParams;

Selectable.args = {
  onAgentsSelected: (agentIds) => console.log({ agentIds }),
};
Selectable.parameters = mockParams;
