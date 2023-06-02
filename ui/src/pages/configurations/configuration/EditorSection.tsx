import { gql } from "@apollo/client";
import {
  Box,
  Card,
  CardContent,
  CardHeader,
  CircularProgress,
  Divider,
  Stack,
} from "@mui/material";
import { ConfigurationEditor } from "../../../components/ConfigurationEditor";
import { RolloutHistory } from "../../../components/RolloutHistory";
import {
  useGetConfigRolloutAgentsQuery,
  useGetRenderedConfigLazyQuery,
} from "../../../graphql/generated";
import { useRefetchOnConfigurationChange } from "../../../hooks/useRefetchOnConfigurationChanges";
import { RawOrTopologyControl } from "../../../components/PipelineGraph/RawOrTopologyControl";
import { useEffect, useState } from "react";
import { YamlEditor } from "../../../components/YamlEditor";

import styles from "./configuration-page.module.scss";

gql`
  query getConfigRolloutAgents($name: String!) {
    configuration(name: $name) {
      metadata {
        name
        id
        version
      }
      agentCount
    }
  }

  query getRenderedConfigValue($name: String!) {
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

interface EditorSectionProps {
  configurationName: string;
  isOtel: boolean;
  hideRolloutActions?: boolean;
}

/**
 * EditorSection renders the configuration editor and rollout history.
 *
 * @param configurationName should be the non-versioned name of the configuration.
 * @param isOtel determines whether to display a Topology Graph or a Yaml Editor.
 * @param hideRolloutActions determines whether to hide the rollout actions.
 * @returns
 */
export const EditorSection: React.FC<EditorSectionProps> = ({
  configurationName,
  isOtel,
  hideRolloutActions,
}) => {
  const [rawOrTopology, setRawOrTopology] = useState<"raw" | "topology">(
    "topology"
  );

  const { data, refetch } = useGetConfigRolloutAgentsQuery({
    variables: { name: configurationName },
    fetchPolicy: "cache-and-network",
  });

  const [fetchRawConfig, { data: rawData, refetch: refetchRaw }] =
    useGetRenderedConfigLazyQuery({
      variables: { name: configurationName },
      fetchPolicy: "cache-and-network",
    });

  useEffect(() => {
    if (rawOrTopology === "raw") {
      fetchRawConfig();
    }
  }, [rawOrTopology, fetchRawConfig]);

  function refetchQueries() {
    refetch();
    refetchRaw();
  }

  useRefetchOnConfigurationChange(configurationName, refetchQueries);

  const shouldShowRolloutHistory =
    (data?.configuration?.agentCount ?? 1) > 0 ||
    data?.configuration?.metadata.version !== 1;

  return (
    <Card className={styles["section-card"]}>
      <CardHeader
        title={isOtel ? "Configuration" : "Topology"}
        titleTypographyProps={{ fontWeight: 600 }}
        classes={{
          root: styles.padding,
        }}
      />
      <Divider />
      <CardContent classes={{ root: styles["card-content"] }}>
        {!isOtel && (
          <Box marginBottom={2}>
            <RawOrTopologyControl
              rawOrTopology={rawOrTopology}
              setTopologyOrRaw={setRawOrTopology}
            />
          </Box>
        )}

        {rawOrTopology === "topology" && (
          <ConfigurationEditor
            configurationName={configurationName}
            isOtel={isOtel}
            hideRolloutActions={hideRolloutActions}
          />
        )}

        {rawOrTopology === "raw" && (
          <YamlOrLoading value={rawData?.configuration?.rendered} />
        )}
      </CardContent>
      <Divider />
      {shouldShowRolloutHistory && (
        <CardContent classes={{ root: styles["card-content"] }}>
          <RolloutHistory configurationName={configurationName} />
        </CardContent>
      )}
    </Card>
  );
};

interface YamlOrLoadingProps {
  value?: null | string;
}
const YamlOrLoading: React.FC<YamlOrLoadingProps> = ({ value }) => {
  if (value === undefined) {
    return (
      <Stack
        width="100%"
        height="100%"
        minHeight={200}
        alignItems="center"
        justifyContent="center"
      >
        <CircularProgress />
      </Stack>
    );
  }
  return <YamlEditor value={value!} readOnly />;
};
