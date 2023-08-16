import { ApolloError, gql } from "@apollo/client";
import { Button, CircularProgress, Stack, Typography } from "@mui/material";
import { useSnackbar } from "notistack";
import { useEffect, useState } from "react";
import ReactFlow, {
  Controls,
  useReactFlow,
  useStore,
  Node,
  Edge,
} from "reactflow";
import { useNavigate } from "react-router-dom";
import {
  DEFAULT_OVERVIEW_GRAPH_PERIOD,
  DEFAULT_TELEMETRY_TYPE,
  TELEMETRY_SIZE_METRICS,
} from "../../components/MeasurementControlBar/MeasurementControlBar";
import { firstActiveTelemetry } from "../../components/PipelineGraph/Nodes/nodeUtils";
import {
  Role,
  useGetOverviewPageQuery,
  useOverviewPageMetricsSubscription,
} from "../../graphql/generated";
import {
  getNodesAndEdges,
  Page,
  updateOverviewMetricData,
} from "../../utils/graph/utils";
import { OverviewDestinationNode, ConfigurationNode } from "./nodes";
import OverviewEdge from "./OverviewEdge";
import { useOverviewPage } from "./OverviewPageContext";
import colors from "../../styles/colors";
import { GraphGradient } from "../../components/GraphComponents";
import { hasPermission } from "../../utils/has-permission";
import { useRole } from "../../hooks/useRole";

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
`;

interface LastDataRecieved {
  query?: Date;
  subscription?: Date;
}

const nodeTypes = {
  destinationNode: OverviewDestinationNode,
  configurationNode: ConfigurationNode,
};

const edgeTypes = {
  overviewEdge: OverviewEdge,
};

export const OverviewGraph: React.FC = () => {
  const [nodes, setNodes] = useState<Node[]>();
  const [edges, setEdges] = useState<Edge[]>();
  const [hasPipeline, setHasPipeline] = useState<boolean>(true);
  const [lastDataRecieved, setLastDataRecieved] = useState<LastDataRecieved>(
    {}
  );

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

  function onError(error: ApolloError) {
    console.error(error.message);
    enqueueSnackbar("Oops! Something went wrong.", {
      variant: "error",
      key: error.message,
    });
  }

  const { loading } = useGetOverviewPageQuery({
    notifyOnNetworkStatusChange: true,
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
    onCompleted(data) {
      setLastDataRecieved((prev) => ({
        ...prev,
        query: new Date(),
      }));

      const { nodes: gotNodes, edges: gotEdges } = getNodesAndEdges(
        Page.Overview,
        data!.overviewPage.graph,
        700,
        null,
        () => {},
        () => {},
        () => {},
        true,
        selectedTelemetry
      );

      setNodes(gotNodes);
      setEdges(gotEdges);

      const determineHasPipeline =
        data.overviewPage.graph.sources.length > 0 &&
        data.overviewPage.graph.targets.length > 0;

      setHasPipeline(determineHasPipeline);

      if (
        selectedTelemetry == null &&
        data.overviewPage.graph.attributes != null
      ) {
        const activeTelemetry = firstActiveTelemetry(
          data.overviewPage.graph.attributes
        );
        if (activeTelemetry) {
          setSelectedTelemetry(activeTelemetry);
        }
      }
    },
    onError,
  });

  const { data: subscriptionData } = useOverviewPageMetricsSubscription({
    skip: loading,
    variables: {
      period: selectedPeriod || DEFAULT_OVERVIEW_GRAPH_PERIOD,
      configIDs: configIDs,
      destinationIDs: destinationIDs,
    },
    onData({ data: subscriptionData }) {
      setLastDataRecieved((prev) => ({
        ...prev,
        subscription: new Date(),
      }));

      const overviewMetrics = subscriptionData.data?.overviewMetrics;
      if (overviewMetrics) {
        setMaxValues({
          maxMetricValue: overviewMetrics.maxMetricValue,
          maxLogValue: overviewMetrics.maxLogValue,
          maxTraceValue: overviewMetrics.maxTraceValue,
        });
      }
    },
    onError,
  });

  useEffect(() => {
    if (!edges || !nodes || !subscriptionData) {
      return;
    }

    // Update metric data if the subscription is newer than the query
    if (
      lastDataRecieved.subscription &&
      lastDataRecieved.query &&
      lastDataRecieved.subscription > lastDataRecieved.query
    ) {
      updateOverviewMetricData(
        subscriptionData?.overviewMetrics,
        edges,
        nodes,
        selectedPeriod || DEFAULT_OVERVIEW_GRAPH_PERIOD,
        selectedTelemetry || DEFAULT_TELEMETRY_TYPE
      );
    }
  }, [
    edges,
    nodes,
    lastDataRecieved,
    selectedPeriod,
    selectedTelemetry,
    subscriptionData,
  ]);

  const reactFlowWidth = useStore((state: { width: any }) => state.width);
  const reactFlowHeight = useStore((state: { height: any }) => state.height);
  const reactFlowNodeCount = useStore(
    (state: { nodeInternals: any }) =>
      Array.from(state.nodeInternals.values()).length || 0
  );
  useEffect(() => {
    reactFlowInstance.fitView();
  }, [reactFlowWidth, reactFlowHeight, reactFlowNodeCount, reactFlowInstance]);

  if (loading || nodes == null) {
    return <LoadingIndicator />;
  }

  function onNodesChange() {
    reactFlowInstance.fitView();
  }

  return hasPipeline ? (
    <div style={{ height: "100%", width: "100%", paddingBottom: 75 }}>
      <ReactFlow
        nodes={nodes}
        edges={edges}
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
  const role = useRole();

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
        Once you&apos;ve created a configuration and rolled it out to an agent,
        you&apos;ll see your data topology here.
      </Typography>
      <Button
        disabled={role === Role.Viewer}
        variant="contained"
        onClick={
          !hasPermission(Role.User, role)
            ? undefined
            : () => navigate("/configurations/new")
        }
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
