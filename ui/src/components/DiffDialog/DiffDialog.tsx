import {
  Alert,
  AlertTitle,
  Box,
  CircularProgress,
  Dialog,
  DialogProps,
  Stack,
  Typography,
} from "@mui/material";
import { ContentSection, TitleSection } from "../DialogComponents";
import { gql } from "@apollo/client";
import CodeDiff, { ReactDiffViewerProps } from "react-diff-viewer-continued";

import colors from "../../styles/colors";
import styles from "./diff-section.module.scss";
import { RenderedConfigData } from "./get-rendered-config-data";
import { useState } from "react";
import { useGetRenderedConfigQuery } from "../../graphql/generated";
import { asCurrentVersion, asLatestVersion } from "../../utils/version-helpers";
interface DiffDialogProps extends DialogProps {
  onClose: () => void;
  configurationName: string;
}

gql`
  query getRenderedConfig($name: String!) {
    configuration(name: $name) {
      metadata {
        name
        id
        version
      }
      rendered
    }
  }
`;

const CODE_DIFF_STYLES: ReactDiffViewerProps["styles"] = {
  contentText: {
    lineBreak: "anywhere",
    fontSize: 12,
  },
  diffContainer: {
    backgroundColor: colors.backgroundGrey,
    pre: {
      lineHeight: "14px",
    },
    a: {
      textDecoration: "none",
    },
  },
  lineNumber: {
    cursor: "default",
    fontSize: 12,
  },
  gutter: {
    ":hover": {
      cursor: "default",
    },
  },
};

/**
 * DiffDialog is used to show the diff between the latest version and the current version.
 *
 * @param onClose callback to close the Dialog
 * @param configurationName the name of the configuration, should not contain a version
 * @returns
 */
export const DiffDialog: React.FC<DiffDialogProps> = ({
  open,
  configurationName,
  onClose,
}) => {
  const [currentConfigData, setCurrentConfigData] =
    useState<RenderedConfigData>();
  const [latestConfigData, setLatestConfigData] =
    useState<RenderedConfigData>();

  useGetRenderedConfigQuery({
    variables: { name: asCurrentVersion(configurationName) },
    fetchPolicy: "network-only",
    onCompleted(data) {
      if (data.configuration) {
        setCurrentConfigData(new RenderedConfigData(data.configuration));
      }
    },
  });

  useGetRenderedConfigQuery({
    variables: { name: asLatestVersion(configurationName) },
    fetchPolicy: "network-only",
    onCompleted(data) {
      if (data.configuration) {
        setLatestConfigData(new RenderedConfigData(data.configuration));
      }
    },
  });

  if (!currentConfigData || !latestConfigData) {
    return (
      <Dialog open={open} onClose={onClose} fullWidth maxWidth="lg">
        <TitleSection onClose={onClose} title="Compare Versions" />
        <ContentSection>
          <Stack
            height="616px"
            width="100%"
            justifyContent="center"
            alignItems="center"
          >
            <CircularProgress />
          </Stack>
        </ContentSection>
      </Dialog>
    );
  }
  const noChanges =
    currentConfigData.value().localeCompare(latestConfigData.value()) === 0;

  return (
    <Dialog open={open} onClose={onClose} fullWidth maxWidth="lg">
      <TitleSection onClose={onClose} title="Compare Versions" />
      <ContentSection>
        <Stack spacing={1}>
          {noChanges && (
            <Alert severity="info">
              <AlertTitle>Identical Rendered Configurations</AlertTitle>
              There are no differences between the New Version and the Current
              Version. This is most likely because the changes you've made
              result in the same rendered OTEL configuration.
            </Alert>
          )}
          <Stack direction="row" justifyContent="space-around" width="100%">
            <Typography fontWeight={600}>
              {currentConfigData.title()}
            </Typography>
            <Typography fontWeight={600}>{latestConfigData.title()}</Typography>
          </Stack>

          <Box className={styles.box}>
            <CodeDiff
              oldValue={currentConfigData.value()}
              newValue={latestConfigData.value()}
              disableWordDiff
              codeFoldMessageRenderer={(total) => (
                <Stack width="200%" justifyContent="center" alignItems="center">
                  <Typography fontSize={14}>Expand {total} lines</Typography>
                </Stack>
              )}
              styles={CODE_DIFF_STYLES}
              showDiffOnly={!noChanges}
            />
          </Box>
        </Stack>{" "}
      </ContentSection>
    </Dialog>
  );
};
