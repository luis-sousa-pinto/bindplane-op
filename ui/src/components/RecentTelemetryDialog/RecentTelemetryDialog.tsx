import { Dialog, DialogContent, DialogProps, Stack } from "@mui/material";
import { isFunction } from "lodash";
import { PipelineType } from "../../graphql/generated";
import { SnapshotConsole } from "../SnapShotConsole/SnapShotConsole";
import {
  SnapshotContextProvider,
  useSnapshot,
} from "../SnapShotConsole/SnapshotContext";
import { DialogContainer } from "../DialogComponents/DialogContainer";

interface RecentTelemetryDialogProps extends DialogProps {
  agentID: string;
}

export const RecentTelemetryDialog: React.FC<RecentTelemetryDialogProps> = ({
  agentID,
  ...dialogProps
}) => {
  function handleClose() {
    isFunction(dialogProps.onClose) && dialogProps.onClose({}, "backdropClick");
  }

  return (
    <SnapshotContextProvider pipelineType={PipelineType.Logs} agentID={agentID}>
      <Dialog
        fullWidth
        maxWidth={"xl"}
        {...dialogProps}
        PaperProps={{
          style: {
            height: "90vh",
            minHeight: "550px",
          },
        }}
      >
        <DialogContent
          style={{
            height: "90vh",
            minHeight: "500px",
          }}
        >
          <Stack
            flexGrow={1}
            height="calc(90vh - 48px)"
            minHeight="500px"
            display="flex"
          >
            <RecentTelemetryBody handleClose={handleClose} />
          </Stack>
        </DialogContent>
      </Dialog>
    </SnapshotContextProvider>
  );
};

const RecentTelemetryBody: React.FC<{
  handleClose: () => void;
}> = ({ handleClose }) => {
  const { logs, metrics, traces, pipelineType } = useSnapshot();
  const footer = `Showing recent ${pipelineType}`;
  return (
    <DialogContainer
      title="Recent Telemetry"
      description="Showing a snapshot of recent telemetry taken before it is sent to a destination"
      onClose={handleClose}
    >
      <Stack flexGrow={1} height="100%">
        <SnapshotConsole
          logs={logs}
          metrics={metrics}
          traces={traces}
          footer={footer}
        />
      </Stack>
    </DialogContainer>
  );
};
