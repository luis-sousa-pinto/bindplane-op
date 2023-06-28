import { useSnackbar } from "notistack";
import { YamlEditor } from "../YamlEditor";
import { Role, useGetConfigurationQuery } from "../../graphql/generated";
import { useState } from "react";
import { BPConfiguration } from "../../utils/classes";
import { Alert, Button, IconButton, Stack } from "@mui/material";
import { EditIcon } from "../Icons";
import { UpdateStatus } from "../../types/resources";
import { RBACWrapper } from "../RBACWrapper/RBACWrapper";

interface OtelConfigProps {
  configurationName: string;
  readOnly?: boolean;
}

export const OtelConfigEditor: React.FC<OtelConfigProps> = ({
  configurationName,
  readOnly,
}) => {
  const { enqueueSnackbar } = useSnackbar();

  const [editing, setEditing] = useState<boolean>(false);
  const [editValue, setEditValue] = useState<string>("");
  const [invalidReason, setInvalidReason] = useState<string | null>(null);

  const { data, refetch } = useGetConfigurationQuery({
    variables: {
      name: configurationName,
    },
    onError(error) {
      console.error(error);
      enqueueSnackbar(`Failed to fetch configuration ${configurationName}.`, {
        variant: "error",
      });
    },
    onCompleted(data) {
      if (data.configuration == null) {
        enqueueSnackbar(`No configuration with name ${configurationName}.`, {
          variant: "error",
        });
        return;
      }
      setEditValue(data.configuration.spec.raw ?? "");
    },
  });

  async function handleSave() {
    try {
      if (data?.configuration == null) {
        throw new Error("No configuration data to apply.");
      }

      const newConfig = new BPConfiguration(data.configuration);
      newConfig.setRaw(editValue);

      const resourceStatus = await newConfig.apply();
      switch (resourceStatus.status) {
        case UpdateStatus.CONFIGURED:
        case UpdateStatus.UNCHANGED:
          setEditing(false);
          setInvalidReason(null);
          await refetch();
          return;
        case UpdateStatus.INVALID:
          setInvalidReason(resourceStatus.reason ?? "Invalid configuration.");
          return;
        default:
          throw new Error(
            `Got unexpected update status: ${resourceStatus.status}`
          );
      }
    } catch (err) {
      console.error(err);
      enqueueSnackbar("Failed to save configuration.", {
        variant: "error",
      });
      return;
    }
  }

  function handleEditValueChange(e: React.ChangeEvent<HTMLTextAreaElement>) {
    setEditValue(e.target.value);
    setInvalidReason(null);
  }

  function handleCancelEdit() {
    setEditing(false);
    setEditValue(data?.configuration?.spec?.raw ?? "");
  }

  const EditAction = readOnly ? null : (
    <RBACWrapper requiredRole={Role.User}>
      <IconButton
        size="small"
        onClick={() => setEditing(true)}
        data-testid="edit-configuration-button"
        sx={{ float: "right" }}
      >
        <EditIcon />
      </IconButton>
    </RBACWrapper>
  );

  const SaveCancelActions = (
    <>
      <Button size="small" color="inherit" onClick={handleCancelEdit}>
        Cancel
      </Button>
      <Button
        data-testid="save-button"
        size="small"
        color="primary"
        variant="outlined"
        onClick={handleSave}
      >
        Save
      </Button>
    </>
  );

  return (
    <Stack marginBottom={4}>
      <YamlEditor
        value={editing ? editValue : data?.configuration?.spec?.raw ?? ""}
        onValueChange={handleEditValueChange}
        readOnly={!editing || readOnly}
        actions={editing ? SaveCancelActions : EditAction}
      />
      {invalidReason != null && (
        <Alert severity="error" sx={{ mt: 2 }}>
          {invalidReason}
        </Alert>
      )}
    </Stack>
  );
};
