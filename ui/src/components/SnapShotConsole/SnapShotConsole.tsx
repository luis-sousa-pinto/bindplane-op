import {
  Alert,
  CircularProgress,
  Grid,
  IconButton,
  Stack,
  ToggleButton,
  ToggleButtonGroup,
  Typography,
} from "@mui/material";
import { memo } from "react";
import { PipelineType } from "../../graphql/generated";
import { RefreshIcon } from "../Icons";
import { SnapshotRow } from "./SnapShotRow";

import { AgentSelector } from "./AgentSelector";
import styles from "./snap-shot-console.module.scss";
import { Log, Metric, Trace, useSnapshot } from "./SnapshotContext";

const TOGGLE_WIDTH = 150;

interface Props {
  hideControls?: boolean;
  logs: Log[];
  metrics: Metric[];
  traces: Trace[];
  footer: string;
}

export const SnapshotConsole: React.FC<Props> = memo(
  ({ hideControls, logs, metrics, traces, footer }) => {
    const {
      loading,
      showAgentSelector,
      pipelineType,
      setPipelineType,
      agentID,
      setAgentID,
      error,
      setError,
      refresh,
    } = useSnapshot();

    return (
      <>
        <MessagesContainer
          type={PipelineType.Logs}
          display={pipelineType === PipelineType.Logs}
          loading={loading}
          messages={logs}
          footer={footer}
        />

        <MessagesContainer
          type={PipelineType.Metrics}
          display={pipelineType === PipelineType.Metrics}
          loading={loading}
          messages={metrics}
          footer={footer}
        />

        <MessagesContainer
          type={PipelineType.Traces}
          display={pipelineType === PipelineType.Traces}
          loading={loading}
          messages={traces}
          footer={footer}
        />

        {!hideControls && (
          <>
            <Grid
              container
              spacing={2}
              sx={{ width: "100%" }}
              alignItems={"center"}
              marginY={1}
            >
              <Grid
                item
                xs={"auto"}
                sx={{ minWidth: "25%", textAlign: "left" }}
              >
                {showAgentSelector && (
                  <AgentSelector
                    agentID={agentID}
                    onChange={setAgentID}
                    onError={setError}
                  />
                )}
              </Grid>
              <Grid item xs={6} sx={{ textAlign: "center" }}>
                <ToggleButtonGroup
                  size={"small"}
                  color="primary"
                  value={pipelineType}
                  exclusive
                  onChange={(_, value) => {
                    if (value != null) {
                      setPipelineType(value);
                    }
                  }}
                  aria-label="Telemetry Type"
                >
                  <ToggleButton
                    sx={{ width: TOGGLE_WIDTH }}
                    value={PipelineType.Logs}
                  >
                    Logs
                  </ToggleButton>
                  <ToggleButton
                    sx={{ width: TOGGLE_WIDTH }}
                    value={PipelineType.Metrics}
                  >
                    Metrics
                  </ToggleButton>
                  <ToggleButton
                    sx={{ width: TOGGLE_WIDTH }}
                    value={PipelineType.Traces}
                  >
                    Traces
                  </ToggleButton>
                </ToggleButtonGroup>
              </Grid>
              <Grid item xs={3} sx={{ textAlign: "right" }}>
                <IconButton
                  color={"primary"}
                  disabled={loading}
                  onClick={refresh}
                >
                  <RefreshIcon />
                </IconButton>
              </Grid>
            </Grid>

            {error && (
              <Alert sx={{ marginTop: 2 }} color="error">
                {error.message}
              </Alert>
            )}
          </>
        )}
      </>
    );
  }
);

const MessagesContainerComponent: React.FC<{
  messages: (Log | Metric | Trace)[] | null;
  type: PipelineType;
  display: boolean;
  loading?: boolean;
  footer: string;
}> = ({ messages, type, display, loading, footer }) => {
  return (
    <div style={{ display: display ? "inline" : "none" }}>
      <div className={styles.container}>
        <div className={styles.console}>
          <div className={styles.stack}>
            {loading ? (
              <Stack
                height="90%"
                width={"100%"}
                justifyContent="center"
                alignItems="center"
              >
                <CircularProgress disableShrink />
              </Stack>
            ) : (
              <>
                {!messages?.length && (
                  <Stack
                    height="100%"
                    width={"100%"}
                    justifyContent="center"
                    alignItems="center"
                    bgcolor={"#fcfcfc"}
                  >
                    <Typography color="secondary">No recent {type}</Typography>
                  </Stack>
                )}
                {messages?.map((m, ix) => (
                  <SnapshotRow key={`${type}-${ix}`} message={m} type={type} />
                ))}
              </>
            )}
          </div>

          <div className={styles.footer}>
            <Typography color="secondary" fontSize={12}>
              {footer}
            </Typography>
          </div>
        </div>
      </div>
    </div>
  );
};

const MessagesContainer = memo(MessagesContainerComponent);
