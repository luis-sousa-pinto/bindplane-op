import { Meta, StoryFn } from "@storybook/react";
import { AgentsTable } from ".";
import {
  AgentChangesDocument,
  AgentsTableDocument,
  AgentsTableQuery,
} from "../../../graphql/generated";
import { AgentsTableField } from "./AgentsDataGrid";
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
        AgentsTableField.NAME,
        AgentsTableField.STATUS,
        AgentsTableField.VERSION,
        AgentsTableField.CONFIGURATION,
        AgentsTableField.CONFIGURATION_VERSION,
        AgentsTableField.OPERATING_SYSTEM,
        AgentsTableField.LABELS,
        AgentsTableField.LOGS,
        AgentsTableField.METRICS,
        AgentsTableField.TRACES,
      ],
      control: "multi-select",
    },
  },
} as Meta<typeof AgentsTable>;

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

Default.args = {
  allowSelection: true,
};
Default.parameters = mockParams;

Selectable.args = {
  onAgentsSelected: (agentIds) => console.log({ agentIds }),
};
Selectable.parameters = mockParams;
