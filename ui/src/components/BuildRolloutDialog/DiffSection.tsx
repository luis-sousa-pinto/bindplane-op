import { gql } from "@apollo/client";
import { Box, CircularProgress, Typography } from "@mui/material";
import { Stack } from "@mui/system";
import { useState } from "react";
import { useGetRenderedConfigQuery } from "../../graphql/generated";
import CodeDiff, { ReactDiffViewerProps } from "react-diff-viewer-continued";
import { RenderedConfigData } from "./get-rendered-config-data";
import { asCurrentVersion, asLatestVersion } from "../../utils/version-helpers";
import colors from "../../styles/colors";

import styles from "./diff-section.module.scss";
interface DiffSectionProps {
  // configurationName should be the non-versioned name of the configuration
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

export const DiffSection: React.FC<DiffSectionProps> = ({
  configurationName,
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
      <Stack
        height="616px"
        width="100%"
        justifyContent="center"
        alignItems="center"
      >
        <CircularProgress />
      </Stack>
    );
  }

  return (
    <Stack>
      <Stack direction="row" justifyContent="space-around" width="100%">
        <Typography fontWeight={600}>{currentConfigData.title()}</Typography>
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
        />
      </Box>
    </Stack>
  );
};
