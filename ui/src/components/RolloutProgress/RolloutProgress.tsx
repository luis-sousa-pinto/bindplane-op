import { gql } from "@apollo/client";
import { useSnackbar } from "notistack";
import { useEffect, useState } from "react";
import { useGetConfigRolloutStatusQuery } from "../../graphql/generated";
import {
  pauseRollout,
  resumeRollout,
  startRollout,
} from "../../utils/rest/rollouts-rest-fns";
import { RolloutProgressBar } from "../RolloutProgressBar";
import { RolloutProgressData } from "./rollout-progress-data";
import { nameAndVersion } from "../../utils/version-helpers";
import { useRefetchOnConfigurationChange } from "../../hooks/useRefetchOnConfigurationChanges";

gql`
  query getConfigRolloutStatus($name: String!) {
    configuration(name: $name) {
      metadata {
        name
        id
        version
      }
      agentCount
      status {
        pending
        current
        latest

        rollout {
          status
          phase
          completed
          errors
          pending
          waiting
        }
      }
    }
  }
`;

interface RolloutProgressProps {
  configurationName: string;
  configurationVersion: string;
  hideActions?: boolean;
  setShowCompareVersions: (show: boolean) => void;
}

/**
 * RolloutProgress wraps the RolloutProgressBar component with a query
 * and subscription for the data.
 * The progress bar is only shown if agents are in the rollout or it's not version 1.
 *
 * @param configurationName The name of the configuration, should not contain a version
 * @param configurationVersion The version of the configuration, should be a string "latest" or "pending"
 * @param actionDisabled Whether the user should be able to start or pause the rollout
 */
export const RolloutProgress: React.FC<RolloutProgressProps> = ({
  configurationName,
  configurationVersion,
  hideActions,
  setShowCompareVersions,
}) => {
  const { enqueueSnackbar } = useSnackbar();

  const [progressData, setProgressData] = useState<RolloutProgressData>();
  const [loading, setLoading] = useState<boolean>(true);

  const versionedName = nameAndVersion(configurationName, configurationVersion);

  const { refetch } = useGetConfigRolloutStatusQuery({
    variables: { name: versionedName },
    onCompleted(data) {
      if (data.configuration) {
        const newData = new RolloutProgressData(data.configuration);
        setProgressData(newData);
        setLoading(false);
      }
    },
  });

  useRefetchOnConfigurationChange(configurationName, refetch);

  /**
   * handleStartRollout is passed to the BuildRolloutDialog and starts the rollout with default options.
   */
  async function handleStartRollout() {
    setLoading(true);
    try {
      await startRollout(versionedName);
      await refetch();
    } catch (err) {
      console.error(err);
      enqueueSnackbar("Failed to start rollout", {
        variant: "error",
        key: "start-failed",
      });
    } finally {
      setLoading(false);
    }
  }

  /**
   * handlePauseRollout is called when the user clicks the "Pause Rollout" button.
   */
  async function handlePauseRollout() {
    setLoading(true);
    try {
      await pauseRollout(versionedName);
      await refetch();
    } catch (err) {
      console.error(err);
      enqueueSnackbar("Failed to pause rollout", {
        variant: "error",
        key: "pause-failed",
      });
    } finally {
      setLoading(false);
    }
  }

  /**
   * handleResumeRollout is called when the user clicks the "Resume Rollout" button.
   */
  async function handleResumeRollout() {
    setLoading(true);
    try {
      await resumeRollout(versionedName);
      await refetch();
    } catch (err) {
      console.error(err);
      enqueueSnackbar("Failed to resume rollout", {
        variant: "error",
        key: "resume-failed",
      });
    } finally {
      setLoading(false);
    }
  }

  useEffect(() => {
    if (progressData?.configuration.metadata.version === 1) {
      setShowCompareVersions(false);
      return;
    }
    if (progressData?.rolloutStatus() === 4) {
      setShowCompareVersions(false);
      return;
    }
    setShowCompareVersions(true);
  }, [progressData, setShowCompareVersions]);

  if (progressData == null) {
    // TODO(dsvanlani): Show a loading indicator
    return null;
  }

  const totalCount =
    progressData.completed() || progressData.rolloutIsStarted()
      ? progressData.total()
      : progressData.agentCount();

  return (
    <>
      {(totalCount > 0 ||
        progressData.configuration.metadata.version !== 1) && (
        <RolloutProgressBar
          totalCount={totalCount}
          errors={progressData.errored()}
          completedCount={progressData.completed()}
          rolloutStatus={progressData.rolloutStatus()}
          hideActions={hideActions}
          paused={!progressData.rolloutIsStarted()}
          loading={loading}
          onPause={handlePauseRollout}
          onStart={handleStartRollout}
          onResume={handleResumeRollout}
        />
      )}
    </>
  );
};
