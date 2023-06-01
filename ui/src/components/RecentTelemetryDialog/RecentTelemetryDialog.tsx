import { Dialog, DialogContent, DialogProps } from "@mui/material";
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
      <Dialog fullWidth maxWidth={"xl"} {...dialogProps}>
        <DialogContent>
          <RecentTelemetryBody handleClose={handleClose} />
        </DialogContent>
      </Dialog>
    </SnapshotContextProvider>
  );
};

const RecentTelemetryBody: React.FC<{
  handleClose: () => void;
}> = ({ handleClose, children }) => {
  const { logs, metrics, traces, pipelineType } = useSnapshot();
  const footer = `Showing recent ${pipelineType}`;
  return (
    <DialogContainer title="Recent Telemetry" description="Showing a snapshot of recent telemetry taken from the current agent configuration before it is sent to a destination" onClose={handleClose}>
      <SnapshotConsole
        logs={logs}
        metrics={metrics}
        traces={traces}
        footer={footer}
      />
    </DialogContainer>
  );
};
