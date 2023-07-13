import { Button, Card, Paper, Stack, Tooltip } from "@mui/material";
import { ReactFlowProvider } from "reactflow";
import { withNavBar } from "../../components/NavBar";
import { ConfigurationsTable } from "../../components/Tables/ConfigurationTable";
import { withRequireLogin } from "../../contexts/RequireLogin";
import {
  useOverviewPageMetricsSubscription,
  useDeployedConfigsQuery,
  useDestinationsQuery,
} from "../../graphql/generated";
import { OverviewGraph } from "./OverviewGraph";
import { OverviewPageProvider, useOverviewPage } from "./OverviewPageContext";
import {
  DEFAULT_OVERVIEW_GRAPH_PERIOD,
  DEFAULT_TELEMETRY_TYPE,
  MeasurementControlBar,
  TELEMETRY_SIZE_METRICS,
} from "../../components/MeasurementControlBar/MeasurementControlBar";
import { gql } from "@apollo/client";
import { DestinationsTableField } from "../../components/Tables/DestinationsTable/DestinationsDataGrid";
import { ConfigurationsTableField } from "../../components/Tables/ConfigurationTable/ConfigurationsDataGrid";
import { DestinationsPageSubContent } from "../destinations/DestinationsPage";
import { useCallback, useEffect, useState } from "react";

import colors from "../../styles/colors";
import mixins from "../../styles/mixins.module.scss";
import styles from "./overview-page.module.scss";
import { GridSortItem } from "@mui/x-data-grid";

gql`
  query DeployedConfigs {
    configurations(onlyDeployedConfigurations: true) {
      configurations {
        metadata {
          id
          name
          version
        }
      }
    }
  }

  subscription OverviewPageMetrics(
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
    }
  }
`;

const OverviewPageSubContent: React.FC = () => {
  const {
    selectedTelemetry,
    selectedConfigs,
    setSelectedConfigs,
    selectedDestinations,
    setSelectedDestinations,
    setSelectedTelemetry,
    selectedPeriod,
    setPeriod,
    editingDestination,
    setEditingDestination,
    loadTop,
    setLoadTop,
  } = useOverviewPage();

  const { data: deployedConfigs } = useDeployedConfigsQuery();
  const { data: destinationsInConfigs } = useDestinationsQuery({
    variables: { filterUnused: true },
  });
  // we need these metrics to select the top three configs on load
  const { data: metrics } = useOverviewPageMetricsSubscription({
    variables: {
      period: selectedPeriod || DEFAULT_OVERVIEW_GRAPH_PERIOD,
      configIDs: deployedConfigs?.configurations?.configurations.map(
        (c) => c.metadata.name
      ),
      destinationIDs: destinationsInConfigs?.destinations.map(
        (d) => d.metadata.name
      ),
    },
  });

  const selectTopResources = useCallback(
    (count: number, resourceType: "configuration" | "destination") => {
      const filteredMetrics =
        metrics?.overviewMetrics.metrics
          .filter((metric) => metric.nodeID.startsWith(resourceType))
          .filter(
            (metric) =>
              metric.name ===
              TELEMETRY_SIZE_METRICS[
                selectedTelemetry || DEFAULT_TELEMETRY_TYPE
              ]
          )
          .sort((a, b) => {
            const valueDiff = b.value - a.value;
            if (valueDiff !== 0) {
              return valueDiff;
            }
            return a.nodeID.localeCompare(b.nodeID);
          }) || [];

      const topResources = filteredMetrics.slice(0, count).map((metric) => {
        return metric.nodeID.split("/")[1];
      });
      return topResources;
    },
    [metrics, selectedTelemetry]
  );

  const selectTopConfigs = useCallback(
    (count: number) => {
      setSortModel([
        {
          field: selectedTelemetry ?? ConfigurationsTableField.LOGS,
          sort: "desc",
        },
      ]);
      const topConfigs = selectTopResources(count, "configuration");
      if (topConfigs) {
        setSelectedConfigs(topConfigs);
      }
    },
    [setSelectedConfigs, selectTopResources, selectedTelemetry]
  );

  const selectTopDestinations = useCallback(
    (count: number) => {
      setDestinationsSortModel([
        {
          field: selectedTelemetry ?? DestinationsTableField.LOGS,
          sort: "desc",
        },
      ]);
      const topDests = selectTopResources(count, "destination").map((name) => {
        // map from name to "Destination|name"
        return `Destination|${name}`;
      });
      if (topDests) {
        setSelectedDestinations(topDests);
      }
    },
    [setSelectedDestinations, selectTopResources, selectedTelemetry]
  );

  useEffect(() => {
    // select top three configs on load
    if (loadTop && metrics && selectedTelemetry) {
      selectTopConfigs(3);
      selectTopDestinations(3);
      setLoadTop(false);
    }
  }, [
    metrics,
    loadTop,
    setLoadTop,
    selectTopConfigs,
    selectTopDestinations,
    selectedTelemetry,
  ]);

  const [sortModel, setSortModel] = useState<GridSortItem[]>([
    { field: selectedTelemetry ?? ConfigurationsTableField.LOGS, sort: "desc" },
  ]);

  const [destinationsSortModel, setDestinationsSortModel] = useState<
    GridSortItem[]
  >([
    { field: selectedTelemetry ?? ConfigurationsTableField.LOGS, sort: "desc" },
  ]);

  return (
    <Stack direction={"row"} spacing={2} height={"calc(100vh - 120px)"}>
      <Stack spacing={1} minWidth="400px">
        <Paper className={styles["overview-table-paper"]}>
          <Tooltip
            enterDelay={1000}
            title="Limit the displayed configurations to the three receiving the most data of the selected telemetry type over the selected period."
          >
            <Button
              variant="contained"
              classes={{ root: mixins["float-right"] }}
              onClick={() => selectTopConfigs(3)}
            >
              Top Three
            </Button>
          </Tooltip>
          <ConfigurationsTable
            allowSelection
            selected={selectedConfigs}
            setSelected={setSelectedConfigs}
            enableDelete={false}
            enableNew={false}
            columns={[
              ConfigurationsTableField.NAME,
              ConfigurationsTableField.LOGS,
              ConfigurationsTableField.METRICS,
              ConfigurationsTableField.TRACES,
            ]}
            overviewPage
            minHeight="calc(50vh - 200px)"
            maxHeight="calc(50vh - 200px)"
            initialState={{
              columns: {
                columnVisibilityModel: {
                  logs: false,
                  metrics: false,
                  traces: false,
                },
              },
            }}
            sortModel={sortModel}
            onSortModelChange={(model) => setSortModel(model)}
          />
        </Paper>

        <Paper className={styles["overview-table-paper"]}>
          <Tooltip
            enterDelay={1000}
            title="Limit the displayed destinations to the three receiving the most data of the selected telemetry type over the selected period."
          >
            <Button
              variant="contained"
              classes={{ root: mixins["float-right"] }}
              onClick={() => selectTopDestinations(3)}
            >
              Top Three
            </Button>
          </Tooltip>
          <DestinationsPageSubContent
            allowSelection
            selected={selectedDestinations}
            setSelected={setSelectedDestinations}
            destinationsPage={false}
            columnFields={[
              DestinationsTableField.ICON_AND_NAME,
              DestinationsTableField.LOGS,
              DestinationsTableField.METRICS,
              DestinationsTableField.TRACES,
            ]}
            initialState={{
              columns: {
                columnVisibilityModel: {
                  logs: false,
                  metrics: false,
                  traces: false,
                },
              },
            }}
            sortModel={destinationsSortModel}
            onSortModelChange={(model) => setDestinationsSortModel(model)}
            editingDestination={editingDestination}
            setEditingDestination={setEditingDestination}
            minHeight="calc(50vh - 180px)"
            maxHeight="calc(50vh - 180px)"
          />
        </Paper>
      </Stack>

      <Card
        style={{
          height: "100%",
          width: "100%",
          backgroundColor: colors.backgroundGrey,
        }}
      >
        <MeasurementControlBar
          telemetry={selectedTelemetry || DEFAULT_TELEMETRY_TYPE}
          onTelemetryTypeChange={setSelectedTelemetry}
          period={selectedPeriod || DEFAULT_OVERVIEW_GRAPH_PERIOD}
          onPeriodChange={setPeriod}
        />
        <ReactFlowProvider>
          <OverviewGraph />
        </ReactFlowProvider>
      </Card>
    </Stack>
  );
};

export const OverviewPageContent: React.FC = () => {
  return (
    <OverviewPageProvider>
      <OverviewPageSubContent />
    </OverviewPageProvider>
  );
};

export const OverviewPage: React.FC = withRequireLogin(
  withNavBar(OverviewPageContent)
);
