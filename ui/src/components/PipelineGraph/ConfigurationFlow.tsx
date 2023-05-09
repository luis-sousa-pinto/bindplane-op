import { useCallback, useEffect } from "react";
import ReactFlow, {
  Controls,
  useReactFlow,
  useStore,
} from "react-flow-renderer";
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
import { gql } from "@apollo/client";
import { useConfigurationMetricsSubscription } from "../../graphql/generated";
import OverviewEdge from "../../pages/overview/OverviewEdge";
import ConfigurationEdge from "./Nodes/ConfigurationEdge";
import { MinimumRequiredConfig } from "./PipelineGraph";
import { useConfigurationPage } from "../../pages/configurations/configuration/ConfigurationPageContext";
import { DummyProcessorNode } from "./Nodes/DummyProcessorNode";
import { usePipelineGraph } from "./PipelineGraphContext";
import { GraphGradient } from "../GraphComponents";

import globals from "../../styles/global.module.scss";

gql`
  subscription ConfigurationMetrics(
    $period: String!
    $name: String!
    $agent: String
  ) {
    configurationMetrics(period: $period, name: $name, agent: $agent) {
      metrics {
        name
        nodeID
        pipelineType
        value
        unit
      }
      maxMetricValue
      maxLogValue
      maxTraceValue
    }
  }
`;

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
  configuration: MinimumRequiredConfig;
  refetchConfiguration: () => void;
  agent: string;
}

export const ConfigurationFlow: React.FC<ConfigurationFlowProps> = ({
  period,
  selectedTelemetry,
  configuration,
  refetchConfiguration,
  agent,
}) => {
  const reactFlowInstance = useReactFlow();
  const { setMaxValues } = usePipelineGraph();

  const { setAddSourceDialogOpen, setAddDestDialogOpen } =
    useConfigurationPage();

  const { nodes, edges } = getNodesAndEdges(
    Page.Configuration,
    configuration?.graph!,
    TARGET_OFFSET_MULTIPLIER,
    configuration,
    refetchConfiguration,
    setAddSourceDialogOpen,
    setAddDestDialogOpen,
    agent === "" ? true : false
  );

  const { data } = useConfigurationMetricsSubscription({
    variables: {
      period,
      name: configuration?.metadata?.name || "",
      agent: agent,
    },
    onData({ data }) {
      if (data.data?.configurationMetrics) {
        setMaxValues({
          maxMetricValue: data.data.configurationMetrics.maxMetricValue,
          maxLogValue: data.data.configurationMetrics.maxLogValue,
          maxTraceValue: data.data.configurationMetrics.maxTraceValue,
        });
      }
    },
  });

  updateMetricData(
    Page.Configuration,
    nodes,
    edges,
    data?.configurationMetrics.metrics ?? [],
    period,
    selectedTelemetry
  );

  const viewPortHeight =
    GRAPH_PADDING +
    Math.max(
      configuration?.graph?.sources?.length || 0,
      configuration?.graph?.targets?.length || 0
    ) *
      GRAPH_NODE_OFFSET;
  const onNodesChange = useCallback(() => {
    reactFlowInstance.fitView();
  }, [reactFlowInstance]);

  const reactFlowWidth = useStore((state: { width: any }) => state.width);
  const reactFlowHeight = useStore((state: { height: any }) => state.height);
  const reactFlowNodeCount = useStore(
    (state: { nodeInternals: any }) =>
      Array.from(state.nodeInternals.values()).length || 0
  );
  useEffect(() => {
    setTimeout(() => {
      reactFlowInstance.fitView();
    }, 0);
  }, [reactFlowWidth, reactFlowHeight, reactFlowNodeCount, reactFlowInstance]);

  return (
    <div style={{ height: viewPortHeight, width: "100%" }}>
      <ReactFlow
        defaultNodes={nodes}
        defaultEdges={edges}
        nodeTypes={nodeTypes}
        edgeTypes={edgeTypes}
        proOptions={{ account: "paid-pro", hideAttribution: true }}
        nodesConnectable={false}
        nodesDraggable={false}
        fitView={true}
        deleteKeyCode={null}
        zoomOnScroll={false}
        panOnDrag={true}
        minZoom={0.1}
        preventScrolling={false}
        // This callback only fires when a node is deleted in the graph (ie select node and press the delete button)
        // Need to figure out how to update after node is deleted from the App
        // This callback *does* fire when a Node is added in the App (?)
        onNodesChange={onNodesChange}
        className={globals["graph"]}
      >
        <Controls showZoom={false} showInteractive={false} />
      </ReactFlow>

      <GraphGradient />
    </div>
  );
};
