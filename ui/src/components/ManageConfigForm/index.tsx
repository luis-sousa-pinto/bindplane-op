import { gql } from "@apollo/client";
import { Button, Typography } from "@mui/material";
import { useSnackbar } from "notistack";
import { Link } from "react-router-dom";
import {
  GetAgentAndConfigurationsQuery,
  Role,
  useRemoveAgentConfigurationMutation,
} from "../../graphql/generated";
import { platformIsContainer } from "../../pages/agents/install";
import mixins from "../../styles/mixins.module.scss";
import { patchConfigLabel } from "../../utils/patch-config-label";
import { classes } from "../../utils/styles";
import styles from "./apply-config-form.module.scss";
import { Config } from "./types";
import { RBACWrapper } from "../RBACWrapper/RBACWrapper";
import { hasPermission } from "../../utils/has-permission";
import { useRole } from "../../hooks/useRole";

gql`
  mutation removeAgentConfiguration($input: RemoveAgentConfigurationInput!) {
    removeAgentConfiguration(input: $input) {
      id
      configurationResource {
        rendered
      }
    }
  }
`;

interface ManageConfigFormProps {
  agent: NonNullable<GetAgentAndConfigurationsQuery["agent"]>;
  configurations: Config[];
  onImport: () => void;
  editing: boolean;
  setEditing: React.Dispatch<React.SetStateAction<boolean>>;
  selectedConfig: Config | undefined;
  setSelectedConfig: React.Dispatch<React.SetStateAction<Config | undefined>>;
}

export const ManageConfigForm: React.FC<ManageConfigFormProps> = ({
  agent,
  configurations,
  onImport,
  editing,
  setEditing,
  selectedConfig,
  setSelectedConfig,
}) => {
  const snackbar = useSnackbar();
  const role = useRole();
  const [removeAgentConfiguration] = useRemoveAgentConfigurationMutation({
    variables: {
      input: {
        agentId: agent.id,
      },
    },
  });

  const configResourceName = agent?.configurationResource?.metadata.name;
  const isRawConfig = configResourceName == null;

  async function onApplyConfiguration() {
    try {
      await patchConfigLabel(agent.id, selectedConfig!.metadata.name);

      setEditing(false);
    } catch (err) {
      console.error("Failed to apply new configuration", err);
      snackbar.enqueueSnackbar("Failed to change configuration.", {
        variant: "error",
        autoHideDuration: 5000,
      });
    }
  }

  function onCancelEdit() {
    setEditing(false);
    setSelectedConfig(
      configurations.find((c) => c.metadata.name === configResourceName)
    );
  }

  // Remove the 'configuration' label and refetch the agent
  async function onRemoveConfiguration() {
    try {
      await removeAgentConfiguration();
      setEditing(false);
    } catch (err) {
      setEditing(false);
      console.error("Failed to remove configuration", err);
      snackbar.enqueueSnackbar("Failed to change configuration.", {
        variant: "error",
        autoHideDuration: 5000,
      });
    }
  }

  const ShowConfiguration: React.FC = () => {
    return (
      <>
        {isRawConfig ? (
          <>
            <Typography variant={"body2"} classes={{ root: mixins["mb-2"] }}>
              This agent configuration is not currently managed by BindPlane.
              {hasPermission(Role.User, role) &&
                " Click import to pull this agent's configuration in as a new managed configuration."}
            </Typography>
          </>
        ) : (
          <>
            <Link to={`/configurations/${configResourceName}`}>
              {configResourceName}
            </Link>
          </>
        )}
      </>
    );
  };

  return (
    <>
      <div
        className={classes([
          mixins.flex,
          mixins["align-center"],
          mixins["mb-3"],
        ])}
      >
        <Typography variant="h6">
          Configuration - {editing ? <></> : <ShowConfiguration />}
        </Typography>

        <div className={styles["title-button-group"]}>
          {editing ? (
            <>
              <Button variant="outlined" onClick={onCancelEdit}>
                Cancel
              </Button>
              {!isRawConfig && (
                <Button
                  className={mixins["ml-2"]}
                  variant="outlined"
                  color="error"
                  onClick={onRemoveConfiguration}
                >
                  Detach
                </Button>
              )}
              <Button
                variant="contained"
                onClick={onApplyConfiguration}
                classes={{ root: mixins["ml-2"] }}
              >
                Apply
              </Button>
            </>
          ) : (
            <>
              {isRawConfig && (
                <RBACWrapper requiredRole={Role.User}>
                  <Button variant="contained" onClick={onImport}>
                    Import
                  </Button>
                </RBACWrapper>
              )}
              {/* k8s agents cannot change their configuration */}
              {!platformIsContainer(agent.platform ?? "") &&
                configurations.length > 0 && (
                  <RBACWrapper requiredRole={Role.User}>
                    <Button
                      className={classes([
                        mixins["ml-2"],
                        styles["choose-button"],
                      ])}
                      variant="text"
                      onClick={() => setEditing(true)}
                    >
                      Choose Another Configuration
                    </Button>
                  </RBACWrapper>
                )}
            </>
          )}
        </div>
      </div>
    </>
  );
};
