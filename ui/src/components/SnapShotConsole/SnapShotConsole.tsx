import {
  Alert,
  CircularProgress,
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
import { Log, Metric, Trace, useSnapshot } from "./SnapshotContext";

import styles from "./snap-shot-console.module.scss";
import mixins from "../../styles/mixins.module.scss";

const TOGGLE_WIDTH = 100;

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
      <Stack className={mixins["flex-grow"]}>
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
            <Stack
              direction="row"
              justifyContent={"space-between"}
              spacing={2}
              marginY={1}
              sx={{ width: "100%" }}
            >
              {showAgentSelector ? (
                <AgentSelector
                  agentID={agentID}
                  query={"-status:disconnected"}
                  onChange={setAgentID}
                  onError={setError}
                />
              ) : (
                <div />
              )}
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
              <IconButton
                color={"primary"}
                disabled={loading}
                onClick={refresh}
              >
                <RefreshIcon />
              </IconButton>
            </Stack>

            {error && (
              <Alert sx={{ marginTop: 2 }} color="error">
                {error.message}
              </Alert>
            )}
          </>
        )}
      </Stack>
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
  if (!display) {
    return null;
  }
  return (
    <div className={styles.stack}>
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
  );
};

const MessagesContainer = memo(MessagesContainerComponent);
