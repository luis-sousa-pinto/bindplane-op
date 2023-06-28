import { TabContext, TabPanel } from "@mui/lab";
import {
  Button,
  Dialog,
  DialogProps,
  Stack,
  Tab,
  Tabs,
  Typography,
} from "@mui/material";
import { useState } from "react";
import { DiffSection } from "./DiffSection";
import { ContentSection, TitleSection } from "../DialogComponents";
import { gql } from "@apollo/client";
import { useGetLatestConfigVersionQuery } from "../../graphql/generated";
import { useRefetchOnConfigurationChange } from "../../hooks/useRefetchOnConfigurationChanges";

gql`
  query getLatestConfigVersion($name: String!) {
    configuration(name: $name) {
      metadata {
        id
        name
        version
      }
    }
  }
`;

interface BuildRolloutDialogProps extends DialogProps {
  onClose: () => void;
  onStartRollout: () => void;
  configurationName: string;
}

type BuildRolloutDialogTab = "build" | "diff";

/**
 * BuildRolloutDialog is used to start a rollout for the first time.
 *
 * @param onClose callback to close the Dialog
 * @param onStartRollout callback to start the rollout
 * @param configurationName the name of the configuration to rollout, should not contain a version
 * @returns
 */
export const BuildRolloutDialog: React.FC<BuildRolloutDialogProps> = ({
  open,
  configurationName,
  onClose,
  onStartRollout,
}) => {
  const [tab, setTab] = useState<BuildRolloutDialogTab>("build");
  const [showDiff, setShowDiff] = useState(false);

  const { refetch } = useGetLatestConfigVersionQuery({
    variables: { name: configurationName },
    fetchPolicy: "network-only",
    onCompleted(data) {
      // Don't show the diff tab if the latest version is 1
      if (data.configuration?.metadata.version !== 1) {
        setShowDiff(true);
      }
    },
  });

  useRefetchOnConfigurationChange(configurationName, refetch);

  function handleTabChange(
    _e: React.SyntheticEvent,
    value: BuildRolloutDialogTab
  ) {
    setTab(value);
  }

  function resetForm() {
    setTab("build");
  }

  return (
    <TabContext value={tab}>
      <Dialog
        open={open}
        onClose={onClose}
        fullWidth
        maxWidth="lg"
        TransitionProps={{
          onExited: resetForm,
        }}
      >
        <TitleSection onClose={onClose} title="Build Rollout" />

        <Stack
          justifyContent="center"
          width="100%"
          alignItems="center"
          marginBottom={1}
        >
          <Tabs onChange={handleTabChange} value={tab}>
            <Tab label="Start New Rollout" value="build" />
            {showDiff && <Tab label="Compare Versions" value="diff" />}
          </Tabs>
        </Stack>
        <ContentSection>
          <TabPanel value="build" sx={{ padding: 0 }}>
            <Typography fontSize={18} marginBottom={2}>
              <strong>Rollout Type:</strong> Immediate
            </Typography>
            <Typography marginBottom={5}>
              Starting this rollout will begin deploying the new configuration
              to connected agents immediately. It will first configure a batch
              of 3 agents and increase the batch size by a multiple of 5 on each
              successful deployment. A maximum of 100 agents will be configured
              at a time.
            </Typography>

            <Stack justifyContent={"center"} alignItems={"center"}>
              <Button size="large" variant="contained" onClick={onStartRollout}>
                Start Rollout
              </Button>
            </Stack>
          </TabPanel>

          <TabPanel value="diff" sx={{ padding: 0 }}>
            <DiffSection configurationName={configurationName} />
          </TabPanel>
        </ContentSection>
      </Dialog>
    </TabContext>
  );
};
