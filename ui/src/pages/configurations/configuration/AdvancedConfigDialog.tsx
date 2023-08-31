import {
  Button,
  Dialog,
  DialogContent,
  DialogProps,
  Stack,
  TextField,
  Typography,
} from "@mui/material";
import { useState } from "react";
import { useSnackbar } from "notistack";
import { BPConfiguration } from "../../../utils/classes";
import { useGetConfigurationQuery } from "../../../graphql/generated";
import { UpdateStatus } from "../../../types/resources";
import { isFunction } from "lodash";
import { PERIODS } from "../../../components/MeasurementControlBar/MeasurementControlBar";

interface Props extends DialogProps {
  onSuccess: () => void;
  configName: string;
}

export const AdvancedConfigDialog: React.FC<Props> = ({
  onSuccess,
  configName,
  ...dialogProps
}) => {
  const [measurementInterval, setMeasurementInterval] = useState<string>();
  const [touched, setTouched] = useState(false);

  const { enqueueSnackbar } = useSnackbar();

  const { data, refetch } = useGetConfigurationQuery({
    variables: {
      name: configName,
    },
    onError(error) {
      console.error("useGetConfigurationQuery", error);
      enqueueSnackbar(`Failed to fetch configuration ${configName}.`, {
        variant: "error",
      });
    },
    onCompleted(data) {
      if (data.configuration == null) {
        enqueueSnackbar(`No configuration with name ${configName}.`, {
          variant: "error",
        });
        return;
      }
      setMeasurementInterval(
        data.configuration.spec.measurementInterval || "10s"
      );
    },
  });

  function clearState() {
    setTouched(false);
  }

  async function updateMeasurementInterval() {
    await refetch();
    if (!data?.configuration) {
      throw new Error("No configuration data to apply.");
    }

    const updatedConfig = new BPConfiguration(data?.configuration);
    if (measurementInterval == null) {
      throw new Error("No measurement interval to apply.");
    }
    updatedConfig.updateMeasurementInterval(measurementInterval);

    const update = await updatedConfig.apply();
    if (update.status === UpdateStatus.INVALID) {
      throw new Error("failed to update measurement interval.");
    }

    await refetch();
    onSuccess();
  }

  async function handleSave(e: React.FormEvent<HTMLFormElement>) {
    e.preventDefault();

    try {
      await updateMeasurementInterval();

      enqueueSnackbar("Measurements interval updated successfully", {
        variant: "success",
      });

      // Call the provided onSuccess function
      if (typeof onSuccess === "function") {
        onSuccess();
      }
    } catch (error) {
      console.error(error);
      enqueueSnackbar("Failed to update Measurements interval", {
        variant: "error",
      });
    }
  }

  function handleChange(
    e: React.ChangeEvent<HTMLInputElement | HTMLTextAreaElement>
  ) {
    if (!touched) {
      setTouched(true);
    }
    setMeasurementInterval(e.target.value);
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
          Advanced Configuration Options
        </Typography>
        <Typography>
          Here you can set the scrape interval for the agent's measurements.
        </Typography>
        <form onSubmit={handleSave}>
          <TextField
            value={measurementInterval}
            onChange={handleChange}
            size="small"
            label="Measurements Scrape Interval"
            name="interval"
            fullWidth
            margin="normal"
            required
            select
            onBlur={() => setTouched(true)}
            SelectProps={{ native: true }}
          >
            {Object.entries(PERIODS).map(([p, label]) => (
              <option key={p} value={label}>
                {label}
              </option>
            ))}
          </TextField>

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
              disabled={!measurementInterval}
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
