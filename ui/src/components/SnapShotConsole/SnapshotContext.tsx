import { ApolloError, NetworkStatus, gql } from "@apollo/client";
import { createContext, useContext, useEffect, useMemo, useState } from "react";
import {
  PipelineType,
  ResourceConfiguration,
  SnapshotQuery,
  useSnapshotQuery,
} from "../../graphql/generated";

// while the query includes all three pipeline types, only the pipelineType specified will have results
gql`
  query snapshot(
    $agentID: String!
    $pipelineType: PipelineType!
    $position: String
    $resourceName: String
  ) {
    snapshot(
      agentID: $agentID
      pipelineType: $pipelineType
      position: $position
      resourceName: $resourceName
    ) {
      metrics {
        name
        timestamp
        value
        unit
        type
        attributes
        resource
      }
      logs {
        timestamp
        body
        severity
        attributes
        resource
      }
      traces {
        name
        traceID
        spanID
        parentSpanID
        start
        end
        attributes
        resource
      }
    }
  }
`;

export type Metric = SnapshotQuery["snapshot"]["metrics"][0];
export type Log = SnapshotQuery["snapshot"]["logs"][0];
export type Trace = SnapshotQuery["snapshot"]["traces"][0];

export interface SnapshotContextValue {
  logs: Log[];
  metrics: Metric[];
  traces: Trace[];

  setLogs(logs: Log[]): void;
  setMetrics(metrics: Metric[]): void;
  setTraces(traces: Trace[]): void;

  // true during initial loading and refetching
  loading: boolean;

  // true if a dropdown of agents should be included
  showAgentSelector: boolean;

  error?: ApolloError;
  setError(error: ApolloError): void;

  agentID?: string;
  setAgentID(agentID: string | undefined): void;

  pipelineType: PipelineType;
  setPipelineType(type: PipelineType): void;

  refresh: () => void;
}

const defaultValue: SnapshotContextValue = {
  logs: [],
  metrics: [],
  traces: [],

  setLogs: () => {},
  setMetrics: () => {},
  setTraces: () => {},

  loading: false,
  showAgentSelector: false,

  error: undefined,
  setError: () => {},

  agentID: undefined,
  setAgentID: () => {},

  pipelineType: PipelineType.Traces,
  setPipelineType: () => {},

  refresh: () => {},
};

export const SnapshotContext = createContext(defaultValue);

export interface SnapshotProviderProps {
  pipelineType: PipelineType;
  agentID?: string;
  showAgentSelector?: boolean;
  position?: "s0" | "d0";
  resourceName?: string;
  processors?: ResourceConfiguration[];
}

export const SnapshotContextProvider: React.FC<SnapshotProviderProps> = ({
  children,
  pipelineType: initialPipelineType,
  agentID: initialAgentID,
  showAgentSelector,
  position,
  resourceName,
}) => {
  const [pipelineType, setPipelineType] =
    useState<PipelineType>(initialPipelineType);

  const [logs, setLogs] = useState<Log[]>([]);
  const [metrics, setMetrics] = useState<Metric[]>([]);
  const [traces, setTraces] = useState<Trace[]>([]);

  const [agentID, setAgentID] = useState<string | undefined>(initialAgentID);
  const [error, setError] = useState<ApolloError>();

  const { loading, refetch, networkStatus } = useSnapshotQuery({
    variables: {
      agentID: agentID ?? "",
      pipelineType,
      position,
      resourceName,
    },
    skip: agentID == null,
    onCompleted: (data) => {
      const { snapshot } = data;
      setLogs(snapshot.logs.slice().reverse());
      setMetrics(snapshot.metrics.slice().reverse());
      setTraces(snapshot.traces.slice().reverse());
      setError(undefined);
    },
    onError: (error) => {
      setError(error);
    },
    fetchPolicy: "network-only",
    notifyOnNetworkStatusChange: true,
  });

  useEffect(() => {
    if (agentID == null) {
      return;
    }
    refetch({ agentID, pipelineType });
  }, [refetch, agentID, pipelineType]);

  const anyLoading = useMemo(
    () => loading || networkStatus === NetworkStatus.refetch,
    [loading, networkStatus]
  );

  return (
    <SnapshotContext.Provider
      value={{
        logs,
        metrics,
        traces,

        setLogs,
        setMetrics,
        setTraces,

        loading: anyLoading,
        showAgentSelector: showAgentSelector ?? false,

        error,
        setError,

        agentID,
        setAgentID,

        pipelineType,
        setPipelineType,

        refresh: () => refetch({ agentID: agentID ?? "", pipelineType }),
      }}
    >
      {children}
    </SnapshotContext.Provider>
  );
};

export function useSnapshot(): SnapshotContextValue {
  return useContext(SnapshotContext);
}
