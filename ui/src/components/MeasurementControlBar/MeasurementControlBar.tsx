import { Box, Stack, Tabs, Tab } from "@mui/material";

import styles from "./measurement-control-bar.module.scss";

export const PERIODS: { [period: string]: string } = {
  "10s": "10s",
  "1m": "1m",
  "5m": "5m",
  "1h": "1h",
  "24h": "24h",
};
export const DEFAULT_PERIOD = "10s";
export const DEFAULT_AGENTS_TABLE_PERIOD = "10s";
export const DEFAULT_CONFIGURATION_TABLE_PERIOD = "10s";
export const DEFAULT_OVERVIEW_GRAPH_PERIOD = "1h";
export const DEFAULT_DESTINATIONS_TABLE_PERIOD = "24h";

export const TELEMETRY_TYPES: { [telemetryType: string]: string } = {
  logs: "Logs",
  metrics: "Metrics",
  traces: "Traces",
};
export const DEFAULT_TELEMETRY_TYPE = "logs";

export const TELEMETRY_SIZE_METRICS: { [telemetryType: string]: string } = {
  logs: "log_data_size",
  metrics: "metric_data_size",
  traces: "trace_data_size",
};

interface MeasurementControlBarProps {
  telemetry: string;
  onTelemetryTypeChange: (telemetry: string) => void;
  period: string;
  onPeriodChange: (period: string) => void;
}

/**
 * MeasurementControlBar is a component that allows the user to change the telemetry type and period
 * for the topology graph.
 *
 * @param onTelemetryTypeChange called when the user changes the telemetry type
 * @param onPeriodChange called when the user changes the period
 * @param telemetry the current stateful telemetry type
 * @param period the current stateful period
 * @returns
 */
export const MeasurementControlBar: React.FC<MeasurementControlBarProps> = ({
  onTelemetryTypeChange,
  onPeriodChange,
  telemetry,
  period,
}) => {
  function handleTelemetryChange(
    _event: React.SyntheticEvent<Element, Event>,
    value: any
  ) {
    onTelemetryTypeChange(value);
  }

  function handlePeriodChange(
    _event: React.SyntheticEvent<Element, Event>,
    value: any
  ) {
    onPeriodChange(value);
  }

  return (
    <Box className={styles.box}>
      <Stack
        direction="row"
        justifyContent="space-between"
        className={styles.stack}
      >
        <Tabs value={telemetry} onChange={handleTelemetryChange}>
          {Object.entries(TELEMETRY_TYPES).map(([t, label]) => (
            <Tab
              key={`telemetry-tab-${t}`}
              value={t}
              label={label}
              classes={{ root: styles.tab }}
            />
          ))}
        </Tabs>

        <Tabs value={period} onChange={handlePeriodChange}>
          {Object.entries(PERIODS).map(([p, label]) => (
            <Tab
              key={`period-tab-${p}`}
              value={p}
              label={label}
              classes={{ root: styles.tab }}
            />
          ))}
        </Tabs>
      </Stack>
    </Box>
  );
};
