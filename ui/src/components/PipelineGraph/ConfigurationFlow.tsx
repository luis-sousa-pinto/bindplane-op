import { useEffect, useMemo } from "react";
import ReactFlow, { Controls, useReactFlow, useStoreApi } from "reactflow";
import SourceNode from "./Nodes/SourceNode";
import DestinationNode from "./Nodes/DestinationNode";
import UIControlNode from "./Nodes/UIControlNode";
import {
  getNodesAndEdges,
  GRAPH_NODE_OFFSET,
  GRAPH_PADDING,
  Page,
  updateMetricData,
} from "../../utils/graph/utils";
import { ProcessorNode } from "./Nodes/ProcessorNode";
import { ConfigurationMetricsSubscription } from "../../graphql/generated";
import OverviewEdge from "../../pages/overview/OverviewEdge";
import ConfigurationEdge from "./Nodes/ConfigurationEdge";
import { DummyProcessorNode } from "./Nodes/DummyProcessorNode";
import { CircularProgress, Stack } from "@mui/material";
import { usePipelineGraph } from "./PipelineGraphContext";
import { MinimumRequiredConfig } from "./PipelineGraph";

import styles from "./pipeline-graph.module.scss";

const nodeTypes = {
  sourceNode: SourceNode,
  destinationNode: DestinationNode,
  uiControlNode: UIControlNode,
  processorNode: ProcessorNode,
  dummyProcessorNode: DummyProcessorNode,
};

const edgeTypes = {
  overviewEdge: OverviewEdge,
  configurationEdge: ConfigurationEdge,
};

export const TARGET_OFFSET_MULTIPLIER = 250;

interface ConfigurationFlowProps {
  period: string;
  selectedTelemetry: string;
  page: Page.Agent | Page.Configuration;
  loading?: boolean;
  measurementData?: ConfigurationMetricsSubscription;
}

/**
 * ConfigurationFlow renders the PipelineGraph for a configuration
 * @param period the period on which to query for metrics
 * @param selectedTelemetry the telemetry type on which to query for metrics
 * @param page either the agent page or configuration page, used to determine edit buttons
 * @param loading whether the graph is loading, will show a loading spinner if true
 * @param measurementData optional data from the ConfigurationMetrics subscription
 * @returns
 */
export const ConfigurationFlow: React.FC<ConfigurationFlowProps> = ({
  period,
  selectedTelemetry,
  page,
  loading,
  measurementData,
}) => {
  const reactFlowInstance = useReactFlow();
  const {
    readOnlyGraph,
    configuration,
    refetchConfiguration,
    setAddDestinationOpen,
    setAddSourceOpen,
  } = usePipelineGraph();

  const { getState } = useStoreApi();
  const { width, height } = getState();

  const viewPortHeight = getViewPortHeight(configuration);

  const { nodes, edges } = useMemo(() => {
    if (configuration?.graph == null) {
      return {
        nodes: [],
        edges: [],
      };
    }

    const { nodes, edges } = getNodesAndEdges(
      page,
      configuration.graph,
      TARGET_OFFSET_MULTIPLIER,
      configuration,
      refetchConfiguration,
      setAddSourceOpen,
      setAddDestinationOpen,
      Boolean(readOnlyGraph)
    );

    if (measurementData) {
      updateMetricData(
        Page.Configuration,
        nodes,
        edges,
        measurementData.configurationMetrics.metrics,
        period,
        selectedTelemetry
      );
    }

    return {
      nodes: nodes,
      edges: edges,
    };
  }, [
    page,
    configuration,
    measurementData,
    period,
    readOnlyGraph,
    refetchConfiguration,
    selectedTelemetry,
    setAddDestinationOpen,
    setAddSourceOpen,
  ]);

  useEffect(() => {
    // Fit the view to the graph when the window is resized
    const fitView = function () {
      reactFlowInstance.fitView();
    };

    window.addEventListener("resize", fitView);

    return () => window.removeEventListener("resize", fitView);
  }, [reactFlowInstance]);

  useEffect(() => {
    // Refit the view when our nodes have changed, i.e. one is deleted.
    // By placing this in a setTimeout with delay 0 we
    // ensure this is called on the next event cycle, not right away.
    setTimeout(() => reactFlowInstance.fitView(), 0);
  }, [nodes, reactFlowInstance]);

  useEffect(() => {
    // Refit the view window when the bounds of the react flow have changed.
    setTimeout(() => reactFlowInstance.fitView(), 0);
  }, [height, width, reactFlowInstance]);

  if (loading) {
    return (
      <div style={{ height: viewPortHeight, width: "100%" }}>
        <Stack
          className={styles.grey}
          width="100%"
          height="100%"
          justifyContent="center"
          alignItems="center"
        >
          <CircularProgress />
        </Stack>
      </div>
    );
  }

  return (
    <div style={{ height: viewPortHeight, width: "100%" }}>
      <ReactFlow
        defaultNodes={nodes}
        defaultEdges={edges}
        nodeTypes={nodeTypes}
        edgeTypes={edgeTypes}
        // Called by a react flow node when entering the viewport,
        // use to refit when we add a destination or source.
        onNodesChange={() => reactFlowInstance.fitView()}
        proOptions={{ account: "paid-pro", hideAttribution: true }}
        nodesConnectable={false}
        nodesDraggable={false}
        fitView={true}
        deleteKeyCode={null}
        zoomOnScroll={false}
        panOnDrag={true}
        minZoom={0.1}
        preventScrolling={false}
        className={styles.grey}
        // This is a hack to hide the graph from the viewport until
        // we have called a fitView on the reactFlowInstance.  This
        // is to prevent the graph from appearing out of view on
        // first load.
        defaultViewport={{ x: 1000, y: 1000, zoom: 1 }}
      >
        <Controls showZoom={false} showInteractive={false} />
      </ReactFlow>
    </div>
  );
};

function getViewPortHeight(configuration: MinimumRequiredConfig) {
  return (
    GRAPH_PADDING +
    Math.max(
      configuration?.graph?.sources?.length || 0,
      configuration?.graph?.targets?.length || 0
    ) *
      GRAPH_NODE_OFFSET
  );
}
