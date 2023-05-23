import { gql } from "@apollo/client";
import { Card, CardContent, CardHeader, Divider } from "@mui/material";
import { ConfigurationEditor } from "../../../components/ConfigurationEditor";
import { RolloutHistory } from "../../../components/RolloutHistory";
import { useGetConfigRolloutAgentsQuery } from "../../../graphql/generated";
import styles from "./configuration-page.module.scss";
import { useRefetchOnConfigurationChange } from "../../../hooks/useRefetchOnConfigurationChanges";

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
  const { data, refetch } = useGetConfigRolloutAgentsQuery({
    variables: { name: configurationName },
    fetchPolicy: "cache-and-network",
  });

  useRefetchOnConfigurationChange(configurationName, refetch);

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
        <ConfigurationEditor
          configurationName={configurationName}
          isOtel={isOtel}
          hideRolloutActions={hideRolloutActions}
        />
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
