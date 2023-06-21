import { gql } from "@apollo/client";
import { Button, Typography } from "@mui/material";
import { useSnackbar } from "notistack";
import React, { useEffect, useState } from "react";
import { Navigate, useParams } from "react-router-dom";
import { CardContainer } from "../../../components/CardContainer";
import { PlusCircleIcon } from "../../../components/Icons";
import { withNavBar } from "../../../components/NavBar";
import { AgentsTable } from "../../../components/Tables/AgentsTable";
import { AgentsTableField } from "../../../components/Tables/AgentsTable/AgentsDataGrid";
import { withRequireLogin } from "../../../contexts/RequireLogin";
import {
  GetConfigurationQuery,
  Role,
  useGetConfigurationLazyQuery,
} from "../../../graphql/generated";
import { selectorString } from "../../../types/configuration";
import { platformIsContainer } from "../../agents/install";
import { ApplyConfigDialog } from "./ApplyConfigDialog";
import { isEmpty } from "lodash";
import { ConfigurationDetails } from "../../../components/ConfigurationDetails";
import { EditorSection } from "./EditorSection";
import { RBACWrapper } from "../../../components/RBACWrapper/RBACWrapper";
import { hasPermission } from "../../../utils/has-permission";
import { useRole } from "../../../hooks/useRole";

import styles from "./configuration-page.module.scss";

gql`
  query GetConfiguration($name: String!) {
    configuration(name: $name) {
      metadata {
        id
        name
        description
        labels
        version
      }
      agentCount
      spec {
        raw
        sources {
          type
          name
          displayName
          parameters {
            name
            value
          }
          processors {
            type
            displayName
            parameters {
              name
              value
            }
            disabled
          }
          disabled
        }
        destinations {
          type
          name
          displayName
          parameters {
            name
            value
          }
          processors {
            type
            displayName
            parameters {
              name
              value
            }
            disabled
          }
          disabled
        }
        selector {
          matchLabels
        }
      }
      graph {
        attributes
        sources {
          id
          type
          label
          attributes
        }
        intermediates {
          id
          type
          label
          attributes
        }
        targets {
          id
          type
          label
          attributes
        }
        edges {
          id
          source
          target
        }
      }
    }
  }
`;

export type ShowPageConfig = GetConfigurationQuery["configuration"];

export const ConfigPageContent: React.FC = () => {
  const { name } = useParams();
  const { enqueueSnackbar } = useSnackbar();
  const role = useRole();

  const [fetchConfig, { data }] = useGetConfigurationLazyQuery({
    fetchPolicy: "cache-and-network",
  });

  const [showApplyDialog, setShowApply] = useState(false);

  useEffect(() => {
    if (name) {
      fetchConfig({
        variables: {
          name: `${name}`,
        },
      });
    }
  }, [fetchConfig, name]);

  if (name == null) {
    return <Navigate to="/configurations" />;
  }

  if (data === undefined) {
    return null;
  }

  if (data.configuration == null) {
    enqueueSnackbar(`No config with name ${name} found.`, {
      variant: "error",
    });

    return <Navigate to="/configurations" />;
  }

  function toast(msg: string, variant: "error" | "success") {
    enqueueSnackbar(msg, { variant: variant, autoHideDuration: 3000 });
  }

  function openApplyDialog() {
    setShowApply(true);
  }

  function closeApplyDialog() {
    setShowApply(false);
  }

  function onApplySuccess() {
    toast("Saved config!", "success");
    closeApplyDialog();
  }

  return (
    <>
      <section>
        <ConfigurationDetails configurationName={name} />
      </section>

      <section>
        <EditorSection
          configurationName={name}
          isOtel={!isEmpty(data.configuration.spec.raw)}
          hideRolloutActions={!hasPermission(Role.Admin, role)}
        />
      </section>

      <section>
        <CardContainer>
          <div className={styles["title-button-row"]}>
            <Typography variant="h5">Agents</Typography>
            {!platformIsContainer(
              data.configuration?.metadata?.labels?.platform
            ) && (
              <RBACWrapper requiredRole={Role.User}>
                <Button
                  onClick={openApplyDialog}
                  variant={"contained"}
                  startIcon={<PlusCircleIcon />}
                >
                  Apply config
                </Button>
              </RBACWrapper>
            )}
          </div>

          <AgentsTable
            allowSelection={false}
            selector={selectorString(data.configuration.spec.selector)}
            columnFields={[
              AgentsTableField.NAME,
              AgentsTableField.STATUS,
              AgentsTableField.OPERATING_SYSTEM,
              AgentsTableField.CONFIGURATION_VERSION,
            ]}
            density="standard"
            minHeight="300px"
          />
        </CardContainer>
      </section>

      {showApplyDialog && (
        <ApplyConfigDialog
          configuration={data.configuration}
          maxWidth="lg"
          fullWidth
          open={showApplyDialog}
          onError={() => toast("Failed to apply config.", "error")}
          onSuccess={onApplySuccess}
          onClose={closeApplyDialog}
          onCancel={closeApplyDialog}
        />
      )}
    </>
  );
};

export const ConfigurationPage = withRequireLogin(
  withNavBar(ConfigPageContent)
);
