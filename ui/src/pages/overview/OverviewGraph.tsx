import { gql } from "@apollo/client";
import { Button, CircularProgress, Stack, Typography } from "@mui/material";
import { useSnackbar } from "notistack";
import { useEffect } from "react";
import ReactFlow, {
  Controls,
  useReactFlow,
  useStore,
} from "react-flow-renderer";
import { useNavigate } from "react-router-dom";
import {
  DEFAULT_OVERVIEW_GRAPH_PERIOD,
  DEFAULT_TELEMETRY_TYPE,
  TELEMETRY_SIZE_METRICS,
} from "../../components/MeasurementControlBar/MeasurementControlBar";
import { firstActiveTelemetry } from "../../components/PipelineGraph/Nodes/nodeUtils";
import {
  useGetOverviewPageQuery,
  useOverviewMetricsSubscription,
} from "../../graphql/generated";
import {
  getNodesAndEdges,
  Page,
  updateMetricData,
} from "../../utils/graph/utils";
import { OverviewDestinationNode, ConfigurationNode } from "./nodes";
import OverviewEdge from "./OverviewEdge";
import { useOverviewPage } from "./OverviewPageContext";
import colors from "../../styles/colors";
import { GraphGradient } from "../../components/GraphComponents";

gql`
  query getOverviewPage(
    $configIDs: [ID!]
    $destinationIDs: [ID!]
    $period: String!
    $telemetryType: String!
  ) {
    overviewPage(
      configIDs: $configIDs
      destinationIDs: $destinationIDs
      period: $period
      telemetryType: $telemetryType
    ) {
      graph {
        attributes
        sources {
          id
          label
          type
          attributes
        }

        intermediates {
          id
          label
          type
          attributes
        }

        targets {
          id
          label
          type
          attributes
        }

        edges {
          id
          source
          target
        }
      }
    }
  }

  subscription OverviewMetrics(
    $period: String!
    $configIDs: [ID!]
    $destinationIDs: [ID!]
  ) {
    overviewMetrics(
      period: $period
      configIDs: $configIDs
      destinationIDs: $destinationIDs
    ) {
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
  destinationNode: OverviewDestinationNode,
  configurationNode: ConfigurationNode,
};

const edgeTypes = {
  overviewEdge: OverviewEdge,
};

export const OverviewGraph: React.FC = () => {
  const {
    selectedTelemetry,
    setSelectedTelemetry,
    selectedPeriod,
    selectedConfigs,
    selectedDestinations,
    setMaxValues,
  } = useOverviewPage();
  const { enqueueSnackbar } = useSnackbar();
  const reactFlowInstance = useReactFlow();
  const navigate = useNavigate();

  // map the selectedDestinations to an array of strings
  const destinationIDs = selectedDestinations.map((id) => id.toString());

  // map the selectedConfigs to an array of strings
  const configIDs = selectedConfigs.map((id) => id.toString());

  const { data, error, loading } = useGetOverviewPageQuery({
    fetchPolicy: "network-only",
    variables: {
      configIDs: configIDs,
      destinationIDs: destinationIDs,
      period: selectedPeriod || DEFAULT_OVERVIEW_GRAPH_PERIOD,
      telemetryType:
        selectedTelemetry != null
          ? TELEMETRY_SIZE_METRICS[selectedTelemetry]
          : DEFAULT_TELEMETRY_TYPE,
    },
  });

  const { data: overviewMetricsData } = useOverviewMetricsSubscription({
    variables: {
      period: selectedPeriod || DEFAULT_OVERVIEW_GRAPH_PERIOD,
      configIDs: configIDs,
      destinationIDs: destinationIDs,
    },
    onData({ data: subscriptionData }) {
      if (subscriptionData.data?.overviewMetrics) {
        setMaxValues({
          maxMetricValue: subscriptionData.data.overviewMetrics.maxMetricValue,
          maxLogValue: subscriptionData.data.overviewMetrics.maxLogValue,
          maxTraceValue: subscriptionData.data.overviewMetrics.maxTraceValue,
        });
      }
    },
  });

  useEffect(() => {
    if (error != null) {
      console.error(error);
      enqueueSnackbar("There was a problem loading the overview graph.", {
        variant: "error",
      });
    }
  }, [enqueueSnackbar, error]);

  useEffect(() => {
    // Set the first selected telemetry to the first active after we load.
    if (
      data?.overviewPage?.graph?.attributes != null &&
      selectedTelemetry == null
    ) {
      setSelectedTelemetry(
        firstActiveTelemetry(data.overviewPage.graph.attributes) ??
          DEFAULT_TELEMETRY_TYPE
      );
    }
  });

  const reactFlowWidth = useStore((state: { width: any }) => state.width);
  const reactFlowHeight = useStore((state: { height: any }) => state.height);
  const reactFlowNodeCount = useStore(
    (state: { nodeInternals: any }) =>
      Array.from(state.nodeInternals.values()).length || 0
  );
  useEffect(() => {
    reactFlowInstance.fitView();
  }, [reactFlowWidth, reactFlowHeight, reactFlowNodeCount, reactFlowInstance]);

  if (loading || data == null || data?.overviewPage == null) {
    return <LoadingIndicator />;
  }

  if (data?.overviewPage.graph == null) {
    enqueueSnackbar("There was a problem loading the overview graph.", {
      variant: "error",
    });
    return null;
  }

  function onNodesChange() {
    reactFlowInstance.fitView();
  }

  const hasPipeline =
    data.overviewPage.graph.sources.length > 0 &&
    data.overviewPage.graph.targets.length > 0;

  const { nodes, edges } = getNodesAndEdges(
    Page.Overview,
    data!.overviewPage.graph,
    700,
    null,
    () => {},
    () => {},
    () => {},
    true
  );
  updateMetricData(
    Page.Overview,
    nodes,
    edges,
    overviewMetricsData?.overviewMetrics.metrics ?? [],
    selectedPeriod || DEFAULT_OVERVIEW_GRAPH_PERIOD,
    selectedTelemetry || DEFAULT_TELEMETRY_TYPE
  );

  return hasPipeline ? (
    <div style={{ height: "100%", width: "100%", paddingBottom: 75 }}>
      <ReactFlow
        defaultNodes={nodes}
        defaultEdges={edges}
        nodeTypes={nodeTypes}
        edgeTypes={edgeTypes}
        nodesConnectable={false}
        nodesDraggable={false}
        proOptions={{ account: "paid-pro", hideAttribution: true }}
        fitView={true}
        deleteKeyCode={null}
        zoomOnScroll={false}
        panOnDrag={true}
        minZoom={0.1}
        maxZoom={1.75}
        onWheel={(event) => {
          window.scrollBy(event.deltaX, event.deltaY);
        }}
        onNodesChange={onNodesChange}
        style={{ backgroundColor: colors.backgroundGrey }}
      >
        <Controls showZoom={false} showInteractive={false} />
      </ReactFlow>
      <GraphGradient />
    </div>
  ) : (
    <NoDeployedConfigurationsMessage navigate={navigate} />
  );
};

const NoDeployedConfigurationsMessage: React.FC<{
  navigate: (to: string) => void;
}> = ({ navigate }) => {
  return (
    <Stack
      width="100%"
      height="calc(100vh - 200px)"
      justifyContent="center"
      alignItems="center"
      spacing={2}
      padding={4}
    >
      <Typography variant="h4" textAlign={"center"}>
        You haven&apos;t deployed any configurations.
      </Typography>
      <Typography textAlign={"center"}>
        Once you&apos;ve created a configuration and applied it to an agent,
        you&apos;ll see your data topology here.
      </Typography>
      <Button
        variant="contained"
        onClick={() => navigate("/configurations/new")}
      >
        Create Configuration Now
      </Button>
    </Stack>
  );
};

const LoadingIndicator: React.FC = () => {
  return (
    <Stack
      width="100%"
      height="calc(100vh - 200px)"
      justifyContent="center"
      alignItems="center"
    >
      <CircularProgress />
    </Stack>
  );
};
