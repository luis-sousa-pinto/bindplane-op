import { Stack } from "@mui/material";
import {
  DataGrid,
  GridCellParams,
  GridColDef,
  GridDensity,
  GridRowParams,
  GridRowSelectionModel,
  GridValueFormatterParams,
  GridValueGetterParams,
} from "@mui/x-data-grid";
import React, { memo, useEffect, useState } from "react";
import { renderAgentLabels, renderAgentStatus } from "../utils";
import { AgentsTableMetricsSubscription } from "../../../graphql/generated";
import { AgentStatus } from "../../../types/agents";
import { isFunction } from "lodash";
import { AgentsTableAgent } from ".";
import { formatMetric } from "../../../utils/graph/utils";
import {
  DEFAULT_AGENTS_TABLE_PERIOD,
  TELEMETRY_SIZE_METRICS,
  TELEMETRY_TYPES,
} from "../../MeasurementControlBar/MeasurementControlBar";
import { SearchLink } from "../../../utils/state";

export enum AgentsTableField {
  NAME = "name",
  STATUS = "status",
  VERSION = "version",
  CONFIGURATION = "configuration",
  CONFIGURATION_VERSION = "configurationVersion",
  OPERATING_SYSTEM = "operatingSystem",
  LABELS = "labels",
  LOGS = "logs",
  METRICS = "metrics",
  TRACES = "traces",
}

interface AgentsDataGridProps {
  onAgentsSelected?: (agentIds: GridRowSelectionModel) => void;
  isRowSelectable?: (params: GridRowParams<AgentsTableAgent>) => boolean;
  clearSelectionModelFnRef?: React.MutableRefObject<(() => void) | null>;
  density?: GridDensity;
  loading: boolean;
  minHeight?: string;
  agents?: AgentsTableAgent[];
  agentMetrics?: AgentsTableMetricsSubscription;
  columnFields?: AgentsTableField[];
  allowSelection: boolean;
}

const AgentsDataGridComponent: React.FC<AgentsDataGridProps> = ({
  clearSelectionModelFnRef,
  onAgentsSelected,
  isRowSelectable,
  minHeight,
  loading,
  agents,
  agentMetrics,
  columnFields,
  density,
  allowSelection,
}) => {
  const [selectionModel, setSelectionModel] = useState<GridRowSelectionModel>(
    []
  );

  useEffect(() => {
    if (clearSelectionModelFnRef == null) {
      return;
    }
    clearSelectionModelFnRef.current = function () {
      setSelectionModel([]);
    };
  }, [setSelectionModel, clearSelectionModelFnRef]);

  const columns: GridColDef[] = (columnFields || []).map((field) => {
    switch (field) {
      case AgentsTableField.STATUS:
        return {
          field: AgentsTableField.STATUS,
          headerName: "Status",
          width: 150,
          renderCell: renderStatusDataCell,
        };
      case AgentsTableField.VERSION:
        return {
          field: AgentsTableField.VERSION,
          headerName: "Version",
          width: 100,
        };
      case AgentsTableField.CONFIGURATION:
        return {
          field: AgentsTableField.CONFIGURATION,
          headerName: "Config",
          width: 200,
          renderCell: renderConfigurationCell,
          valueGetter: (params: GridValueGetterParams<AgentsTableAgent>) => {
            const configuration = params.row.configurationResource;
            return configuration?.metadata?.name;
          },
        };
      case AgentsTableField.OPERATING_SYSTEM:
        return {
          field: AgentsTableField.OPERATING_SYSTEM,
          headerName: "Operating System",
          width: 200,
        };
      case AgentsTableField.LABELS:
        return {
          sortable: false,
          field: AgentsTableField.LABELS,
          headerName: "Labels",
          width: 200,
          renderCell: renderLabelDataCell,
          valueGetter: (params: GridValueGetterParams<AgentsTableAgent>) => {
            return params.row.labels;
          },
        };
      case AgentsTableField.LOGS:
        return createMetricRateColumn(field, "logs", 100, agentMetrics);
      case AgentsTableField.METRICS:
        return createMetricRateColumn(field, "metrics", 100, agentMetrics);
      case AgentsTableField.TRACES:
        return createMetricRateColumn(field, "traces", 100, agentMetrics);
      case AgentsTableField.CONFIGURATION_VERSION:
        return {
          sortable: true,
          field: AgentsTableField.CONFIGURATION_VERSION,
          headerName: "Config Version",
          width: 200,
          valueGetter: (params: GridValueGetterParams<AgentsTableAgent>) => {
            if (!params.row.configurationResource) {
              return "-";
            }

            const configuration = params.row.configurationResource;
            const configurationLabel = params.row.labels?.configuration;

            const matches = configuration.metadata.name === configurationLabel;

            return matches ? configuration.metadata.version : "-";
          },
        };
      default:
        return {
          field: AgentsTableField.NAME,
          headerName: "Name",
          valueGetter: (params: GridValueGetterParams<AgentsTableAgent>) => {
            return params.row.name;
          },
          renderCell: renderNameDataCell,
          width: 240,
        };
    }
  });

  function handleSelect(s: GridRowSelectionModel) {
    setSelectionModel(s);

    isFunction(onAgentsSelected) && onAgentsSelected(s);
  }

  return (
    <DataGrid
      checkboxSelection={isFunction(onAgentsSelected) && allowSelection}
      isRowSelectable={isRowSelectable}
      onRowSelectionModelChange={handleSelect}
      rowSelectionModel={selectionModel}
      density={density}
      components={{
        NoRowsOverlay: () => (
          <Stack height="100%" alignItems="center" justifyContent="center">
            No Agents
          </Stack>
        ),
      }}
      style={{ minHeight }}
      loading={loading}
      disableRowSelectionOnClick
      columns={columns}
      rows={agents ?? []}
    />
  );
};

function renderConfigurationCell(cellParams: GridCellParams<any, string>) {
  const configName = cellParams.value;
  if (configName == null) {
    return <></>;
  }
  return (
    <SearchLink
      path={`/configurations/${configName}`}
      displayName={configName}
    />
  );
}

function renderNameDataCell(
  cellParams: GridCellParams<{ name: string; id: string }, AgentsTableAgent>
): JSX.Element {
  return (
    <SearchLink
      path={`/agents/${cellParams.row.id}`}
      displayName={cellParams.row.name}
    />
  );
}

function renderLabelDataCell(
  cellParams: GridCellParams<any, Record<string, string>>
): JSX.Element {
  return renderAgentLabels(cellParams.value);
}

function renderStatusDataCell(
  cellParams: GridCellParams<any, AgentStatus>
): JSX.Element {
  return renderAgentStatus(cellParams.value);
}

function createMetricRateColumn(
  field: string,
  telemetryType: string,
  width: number,
  agentMetrics?: AgentsTableMetricsSubscription
): GridColDef[][0] {
  return {
    field,
    width: width,
    headerName: TELEMETRY_TYPES[telemetryType],
    valueGetter: (params: GridValueGetterParams) => {
      if (agentMetrics == null) {
        return "";
      }
      // should probably have a lookup table here rather than interpolate in two places
      const metricName = TELEMETRY_SIZE_METRICS[telemetryType];
      const agentName = params.id;

      // get all metrics for this agent that match the pattern /^destination\/\w+$/
      // those are metrics for data received by a destination, ignoring values before the processors
      const metrics = agentMetrics.agentMetrics.metrics.filter(
        (m) =>
          m.name === metricName &&
          m.agentID! === agentName &&
          m.nodeID.startsWith("destination/") &&
          !m.nodeID.endsWith("/processors")
      );
      if (metrics == null) {
        return 0;
      }
      // to make this sortable, we use the raw value and provide a valueFormatter implementation to show units
      return metrics.reduce((a, b) => a + b.value, 0);
    },
    valueFormatter: (params: GridValueFormatterParams<number>): string => {
      if (params.value === 0) {
        return "";
      }

      const metricName = TELEMETRY_SIZE_METRICS[telemetryType];
      const agentName = params.id;

      const metrics = agentMetrics?.agentMetrics.metrics.find(
        (m) => m.name === metricName && m.agentID! === agentName
      );
      return formatMetric(
        { value: params.value, unit: metrics?.unit || "B/s" },
        DEFAULT_AGENTS_TABLE_PERIOD
      );
    },
  };
}

AgentsDataGridComponent.defaultProps = {
  minHeight: "calc(100vh - 300px)",
  columnFields: [
    AgentsTableField.NAME,
    AgentsTableField.STATUS,
    AgentsTableField.VERSION,
    AgentsTableField.CONFIGURATION,
    AgentsTableField.LOGS,
    AgentsTableField.METRICS,
    AgentsTableField.TRACES,
    AgentsTableField.OPERATING_SYSTEM,
    AgentsTableField.LABELS,
  ],
};

export const AgentsDataGrid = memo(AgentsDataGridComponent);
