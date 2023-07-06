import {
  Button,
  Dialog,
  DialogContent,
  DialogProps,
  Stack,
  TextField,
  Typography,
} from "@mui/material";
import { isFunction } from "lodash";
import { useSnackbar } from "notistack";
import { useEffect, useState } from "react";
import { useNavigate } from "react-router-dom";
import { useGetConfigNamesLazyQuery } from "../../../graphql/generated";
import { validateNameField } from "../../../utils/forms/validate-name-field";
import { copyConfig } from "../../../utils/rest/copy-config";

interface Props extends DialogProps {
  currentConfigName: string;
  onSuccess: () => void;
}

export const DuplicateConfigDialog: React.FC<Props> = ({
  currentConfigName,
  onSuccess,
  ...dialogProps
}) => {
  const [newName, setNewName] = useState("");
  const [touched, setTouched] = useState(false);
  const [existingConfigNames, setExistingConfigNames] = useState<string[]>([]);

  const [fetchConfigNames] = useGetConfigNamesLazyQuery({
    fetchPolicy: "network-only",
    onCompleted: (data) => {
      setExistingConfigNames(
        data.configurations.configurations.map((c) => c.metadata.name)
      );
    },
    onError: (error) => {
      console.error(error);
      enqueueSnackbar("Error retrieving config names.", {
        variant: "error",
      });
    },
  });

  const formError = validateNameField(
    newName,
    "configuration",
    existingConfigNames
  );

  const { enqueueSnackbar } = useSnackbar();
  const navigate = useNavigate();

  function clearState() {
    setTouched(false);
    setNewName("");
  }

  useEffect(() => {
    if (dialogProps.open) {
      fetchConfigNames();
    }
  }, [dialogProps.open, fetchConfigNames]);

  async function handleSave(e: React.FormEvent<HTMLFormElement>) {
    e.preventDefault();

    const status = await copyConfig({
      existingName: currentConfigName,
      newName: newName,
    });

    let message: string;
    switch (status) {
      case "conflict":
        message = "Looks like a configuration with that name already exists.";
        enqueueSnackbar(message, { key: message, variant: "warning" });
        break;
      case "error":
        message = "Oops, something went wrong. Failed to duplicate.";
        enqueueSnackbar(message, { key: message, variant: "error" });
        break;
      case "created":
        message = "Successfully duplicated!";
        onSuccess();
        enqueueSnackbar(message, { key: message, variant: "success" });
        navigate(`/configurations/${newName}`);
        break;
    }
  }

  function handleChange(
    e: React.ChangeEvent<HTMLInputElement | HTMLTextAreaElement>
  ) {
    if (!touched) {
      setTouched(true);
    }
    setNewName(e.target.value);
  }

  return (
    <Dialog
      {...dialogProps}
      TransitionProps={{
        onExited: clearState,
      }}
    >
      <DialogContent>
        <Typography variant="h6" marginBottom={2}>
          Duplicate Configuration
        </Typography>
        <Typography>
          Clicking save will create a new Configuration with identical sources
          and destinations.
        </Typography>
        <form onSubmit={handleSave}>
          <TextField
            value={newName}
            autoComplete="off"
            onChange={handleChange}
            size="small"
            label="Name"
            helperText={touched && formError ? formError : undefined}
            name="name"
            fullWidth
            error={touched && formError != null}
            margin="normal"
            required
            onBlur={() => setTouched(true)}
          />

          <Stack
            direction="row"
            justifyContent="end"
            spacing={1}
            marginTop="8px"
          >
            <Button
              color="secondary"
              variant="outlined"
              onClick={() => {
                isFunction(dialogProps.onClose) &&
                  dialogProps.onClose({}, "backdropClick");
              }}
            >
              Cancel
            </Button>
            <Button
              variant="contained"
              disabled={formError != null}
              type="submit"
            >
              Save
            </Button>
          </Stack>
        </form>
      </DialogContent>
    </Dialog>
  );
};
