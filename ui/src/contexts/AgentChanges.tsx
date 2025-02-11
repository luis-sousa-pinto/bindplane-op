import { createContext } from "react";
import { AgentChangesSubscriptionResult, useAgentChangesSubscription } from "../graphql/generated";

export type AgentChange = NonNullable<AgentChangesSubscriptionResult["data"]>["agentChanges"][0]

interface AgentChangesContextValue {
  agentChanges: AgentChange[];
}

export const AgentChangesContext = createContext<AgentChangesContextValue>({
  agentChanges: [],
});

export const AgentChangesProvider: React.FC = ({ children }) => {
  const { data } = useAgentChangesSubscription();
  return (
    <AgentChangesContext.Provider
      value={{ agentChanges: data?.agentChanges || [] }}
    >
      {children}
    </AgentChangesContext.Provider>
  );
};
