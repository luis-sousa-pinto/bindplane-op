import { LoadingButton } from "@mui/lab";
import { Box, Button, LinearProgress, Stack, Typography } from "@mui/material";

import styles from "./rollout-progress.module.scss";
import { useMemo } from "react";

interface RolloutProgressProps {
  totalCount: number;
  errors: number;
  completedCount: number;
  rolloutStatus: number;
  hideActions?: boolean;
  paused: boolean;
  loading?: boolean;
  onPause: () => void;
  onStart: () => void;
  onResume: () => void;
}

/**
 * RolloutProgress is a component that displays the progress of a rollout
 * and allows the user to pause or start the rollout.
 *
 * @param totalCount the total number of agents in the rollout
 * @param errors the number of errored agents in the rollout
 * @param completedCount the number of agents that have completed the rollout
 * @param rolloutStatus used to determine the verbiage of the control button
 * @param hideActions whether to hide the pause/resume/start buttons
 * @param paused whether the rollout is paused, if true,
 * the control button will be "Start Rollout", otherwise it will be "Pause"
 * @param loading whether to display a loading state in the action button
 * @param onPause callback for when the "Pause" button is clicked
 * @param onStartRollout callback for when the "Start Rollout" button is clicked
 * @returns
 */
export const RolloutProgressBar: React.FC<RolloutProgressProps> = ({
  totalCount,
  errors,
  completedCount,
  rolloutStatus,
  hideActions,
  loading,
  onPause,
  onStart,
  onResume,
}) => {
  const value = (completedCount / totalCount) * 100;

  const actionButton = useMemo(() => {
    if (hideActions) {
      return null;
    }

    switch (rolloutStatus) {
      case 0: // pending
        return (
          <Button
            classes={{ root: styles.button }}
            color="secondary"
            variant="contained"
            onClick={onStart}
          >
            Start Rollout
          </Button>
        );
      case 1: // started
        return (
          <LoadingButton
            classes={{ root: styles.button }}
            color="secondary"
            variant="contained"
            onClick={onPause}
            loading={loading}
          >
            Pause
          </LoadingButton>
        );
      case 2: // paused
      case 3: // errored
        return (
          <LoadingButton
            classes={{ root: styles.button }}
            color="secondary"
            variant="contained"
            onClick={onResume}
            loading={loading}
          >
            Resume
          </LoadingButton>
        );
    }
  }, [hideActions, loading, onPause, onResume, onStart, rolloutStatus]);

  const label = useMemo(() => {
    switch (rolloutStatus) {
      case 1: // started
        return "Rollout in Progress";
      case 2: // paused
      case 3: // errored
        return "Rollout Paused";
      case 4: // stable
        return "Rollout Complete";
      default:
        return "Rollout";
    }
  }, [rolloutStatus]);

  return (
    <Box>
      <Stack direction="row" width="100%" alignItems={"center"}>
        <Box flexGrow={1}>
          <Stack
            direction="row"
            justifyContent="space-between"
            alignItems="flex-end"
          >
            <Stack
              direction="row"
              alignItems="center"
              justifyContent="center"
              marginBottom="8px"
              spacing={2}
            >
              <Typography fontSize={18} fontWeight={600}>
                {label}
              </Typography>
              {errors > 0 && (
                <Typography color="error" fontSize={14}>
                  {errors} error{errors > 1 ? "s" : ""}
                </Typography>
              )}
            </Stack>

            <Typography fontSize={16} fontWeight={600} marginBottom="4px">
              {completedCount}/{totalCount}
            </Typography>
          </Stack>

          <LinearProgress variant="determinate" value={value} />
        </Box>
        <Box className={styles["control-box"]}>{actionButton}</Box>
      </Stack>
    </Box>
  );
};
