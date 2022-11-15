import { gql } from "@apollo/client";
import {
  Button,
  Card,
  CircularProgress,
  Stack,
  Typography,
} from "@mui/material";
import { useSnackbar } from "notistack";
import { useEffect, useState } from "react";
import ReactFlow, { Controls, useReactFlow } from "react-flow-renderer";
import { useNavigate } from "react-router-dom";
import {
  DEFAULT_PERIOD,
  DEFAULT_TELEMETRY_TYPE,
  MeasurementControlBar,
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

import global from "../../styles/global.module.scss";

gql`
  query getOverviewPage {
    overviewPage {
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

  subscription OverviewMetrics($period: String!) {
    overviewMetrics(period: $period) {
      metrics {
        name
        nodeID
        pipelineType
        value
        unit
      }
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
  const [selectedPeriod, setPeriod] = useState(DEFAULT_PERIOD);
  const { selectedTelemetry, onTelemetryTypeChange } = useOverviewPage();
  const { enqueueSnackbar } = useSnackbar();
  const reactFlowInstance = useReactFlow();
  const navigate = useNavigate();

  const { data, error, loading } = useGetOverviewPageQuery({
    fetchPolicy: "network-only",
  });

  const { data: overviewMetricsData } = useOverviewMetricsSubscription({
    variables: {
      period: selectedPeriod,
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
    if (data?.overviewPage?.graph?.attributes != null) {
      onTelemetryTypeChange(
        firstActiveTelemetry(data.overviewPage.graph.attributes) ??
          DEFAULT_TELEMETRY_TYPE
      );
    }
  }, [data?.overviewPage.graph.attributes, onTelemetryTypeChange]);

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
    500,
    null,
    () => {}
  );
  updateMetricData(
    Page.Overview,
    nodes,
    edges,
    overviewMetricsData?.overviewMetrics.metrics ?? [],
    selectedPeriod,
    selectedTelemetry
  );

  return hasPipeline ? (
    <Card style={{ height: "calc(100vh - 120px)", width: "100%" }}>
      <div style={{ height: "100%", width: "100%", paddingBottom: 50 }}>
        <MeasurementControlBar
          telemetry={selectedTelemetry}
          onTelemetryTypeChange={onTelemetryTypeChange}
          period={selectedPeriod}
          onPeriodChange={(r: string) => setPeriod(r)}
        />
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
          className={global["graph"]}
        >
          <Controls showZoom={false} showInteractive={false} />
        </ReactFlow>
      </div>
    </Card>
  ) : (
    <NoConfigurationMessage navigate={navigate} />
  );
};

const NoConfigurationMessage: React.FC<{ navigate: (to: string) => void }> = ({
  navigate,
}) => {
  return (
    <Stack
      width="100%"
      height="calc(100vh - 200px)"
      justifyContent="center"
      alignItems="center"
      spacing={1}
    >
      <Typography variant="h4">
        You haven’t built any configurations yet.
      </Typography>
      <Typography>
        Once you&apos;ve created one, you&apos;ll see your data topology here.
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
