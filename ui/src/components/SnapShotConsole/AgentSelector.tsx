import { ApolloError, gql } from "@apollo/client";
import {
  FormControl,
  InputLabel,
  MenuItem,
  Select,
  SelectChangeEvent,
} from "@mui/material";
import { useAgentsWithConfigurationQuery } from "../../graphql/generated";
import { selectorString } from "../../types/configuration";
import { usePipelineGraph } from "../PipelineGraph/PipelineGraphContext";

gql`
  query agentsWithConfiguration($selector: String, $query: String) {
    agents(selector: $selector, query: $query) {
      agents {
        id
        name
      }
    }
  }
`;

interface AgentSelectorProps {
  label?: string;
  agentID?: string;
  onChange: (agentID: string) => void;
  onError: (error: ApolloError) => void;
}

export const AgentSelector: React.FC<AgentSelectorProps> = ({
  label,
  agentID,
  onChange,
  onError,
}) => {
  label ||= "Agent";

  const { configuration } = usePipelineGraph();
  const { data, loading } = useAgentsWithConfigurationQuery({
    variables: {
      selector: selectorString(configuration?.spec?.selector),
    },
    onCompleted(data) {
      if (!agentID && data?.agents?.agents?.length) {
        // ensure that we have a selection if we have agents
        onChange(data.agents.agents[0].id);
      }
    },
    onError: onError,
    fetchPolicy: "cache-and-network",
  });

  return (
    <FormControl size="small" sx={{ minWidth: "150px", maxWidth: "200px" }}>
      <InputLabel id="agent-label">{label}</InputLabel>
      <Select
        inputProps={{ "data-testid": "agent-select" }}
        labelId="agent-label"
        id="agent-id"
        label={label}
        onChange={(e: SelectChangeEvent<string>) => {
          onChange(e.target.value);
        }}
        value={agentID ?? (loading ? "_loading_" : "")}
      >
        {loading && (
          <MenuItem key="loading" value="_loading_" disabled>
            Loading...
          </MenuItem>
        )}
        {data?.agents.agents.map((agent) => (
          <MenuItem
            key={agent.id}
            value={agent.id}
            data-testid={`agent-${agent.id}`}
          >
            {agent.name}
          </MenuItem>
        ))}
      </Select>
    </FormControl>
  );
};
