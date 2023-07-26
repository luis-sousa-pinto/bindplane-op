import { Dialog, DialogProps } from "@mui/material";
import {
  AdditionalInfo,
  Parameter,
  ParameterDefinition,
} from "../../graphql/generated";
import { ResourceConfigForm } from "../ResourceConfigForm";
import { ResourceDialogContextProvider } from "./ResourceDialogContext";
import { isFunction } from "lodash";

interface EditResourceDialogProps extends DialogProps {
  resourceTypeDisplayName: string;
  displayName?: string;
  description: string;
  additionalInfo?: AdditionalInfo | null;
  onSave: (values: { [key: string]: any }) => void;
  onDelete?: () => void;
  onCancel: () => void;
  parameters: Parameter[];
  parameterDefinitions: ParameterDefinition[];
  includeNameField?: boolean;
  kind: "source" | "destination";
  // The supported telemetry types of the resource type that is
  // being configured.  a subset of ['logs', 'metrics', 'traces']
  telemetryTypes?: string[];
  paused?: boolean;
  onTogglePause?: () => void;
  readOnly?: boolean;
}

const EditResourceDialogComponent: React.FC<EditResourceDialogProps> = ({
  displayName,
  onSave,
  onDelete,
  onTogglePause,
  onCancel,
  resourceTypeDisplayName,
  description,
  additionalInfo,
  parameters,
  parameterDefinitions,
  kind,
  telemetryTypes,
  includeNameField = false,
  paused = false,
  readOnly,
  ...dialogProps
}) => {
  return (
    <Dialog
      {...dialogProps}
      onClose={onCancel}
      fullWidth
      maxWidth="md"
      PaperProps={{
        style: {
          height: "85vh",
        },
      }}
    >
      <ResourceConfigForm
        includeNameField={includeNameField}
        includeDisplayNameField={kind === "source"}
        resourceTypeDisplayName={resourceTypeDisplayName}
        displayName={displayName}
        additionalInfo={additionalInfo}
        description={description}
        kind={kind}
        parameterDefinitions={parameterDefinitions}
        parameters={parameters}
        onSave={onSave}
        onDelete={onDelete}
        telemetryTypes={telemetryTypes}
        paused={paused}
        onTogglePause={onTogglePause}
        readOnly={readOnly}
      />
    </Dialog>
  );
};

export const EditResourceDialog: React.FC<EditResourceDialogProps> = (
  props
) => {
  function handleClose() {
    if (isFunction(props.onClose)) {
      props.onClose({}, "backdropClick");
    }
  }
  return (
    <ResourceDialogContextProvider purpose="edit" onClose={handleClose}>
      <EditResourceDialogComponent {...props} />
    </ResourceDialogContextProvider>
  );
};
