import { gql } from "@apollo/client";
import { IconButton, Stack, ToggleButton, ToggleButtonGroup, Typography } from "@mui/material";
import { useSnackbar } from "notistack";
import React, { useState } from "react";
import { useNavigate, useParams } from "react-router-dom";
import { CardContainer } from "../../../components/CardContainer";
import { PlusCircleIcon } from "../../../components/Icons";
import { withNavBar } from "../../../components/NavBar";
import { PipelineGraph } from "../../../components/PipelineGraph/PipelineGraph";
import { AgentsTable } from "../../../components/Tables/AgentsTable";
import { AgentsTableField } from "../../../components/Tables/AgentsTable/AgentsDataGrid";
import { withRequireLogin } from "../../../contexts/RequireLogin";
import {
  GetConfigurationQuery,
  useGetConfigurationQuery,
} from "../../../graphql/generated";
import { selectorString } from "../../../types/configuration";
import { platformIsKubernetes } from "../../agents/install";
import { AddDestinationsSection } from "./AddDestinationsSection";
import { AddSourcesSection } from "./AddSourcesSection";
import { ApplyConfigDialog } from "./ApplyConfigDialog";
import { ConfigurationPageContextProvider } from "./ConfigurationPageContext";
import { ConfigurationSection } from "./ConfigurationSection";
import { DetailsSection } from "./DetailsSection";
import styles from "./configuration-page.module.scss";

gql`
  query GetConfiguration($name: String!) {
    configuration(name: $name) {
      metadata {
        id
        name
        description
        labels
      }
      spec {
        raw
        sources {
          type
          name
          parameters {
            name
            value
          }
          processors {
            type
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
          parameters {
            name
            value
          }
          processors {
            type
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

const ConfigPageContent: React.FC = () => {
  const { name } = useParams();

  // Get Configuration Data
  const { data, refetch } = useGetConfigurationQuery({
    variables: { name: name ?? "" },
    fetchPolicy: "cache-and-network",
  });

  function toast(msg: string, variant: "error" | "success") {
    enqueueSnackbar(msg, { variant: variant, autoHideDuration: 3000 });
  }

  const [showApplyDialog, setShowApply] = useState(false);
  const [addSourceDialogOpen, setAddSourceDialogOpen] = useState(false);
  const [addDestDialogOpen, setAddDestDialogOpen] = useState(false);

  const navigate = useNavigate();
  const { enqueueSnackbar } = useSnackbar();

  const isRaw = (data?.configuration?.spec?.raw?.length || 0) > 0;

  function openApplyDialog() {
    setShowApply(true);
  }

  function closeApplyDialog() {
    setShowApply(false);
  }

  function onApplySuccess() {
    toast("Saved configuration!", "success");
    closeApplyDialog();
  }

  if (data?.configuration === undefined) {
    return null;
  }

  if (data.configuration === null) {
    enqueueSnackbar(`No configuration with name ${name} found.`, {
      variant: "error",
    });
    navigate("/configurations");
    return null;
  }

  return (
    <ConfigurationPageContextProvider
      configuration={data.configuration!}
      setAddDestDialogOpen={setAddDestDialogOpen}
      setAddSourceDialogOpen={setAddSourceDialogOpen}
      refetchConfiguration={refetch}
    >
      <section>
        <DetailsSection
          configuration={data.configuration}
          refetch={refetch}
          onSaveDescriptionError={() =>
            toast("Failed to save description.", "error")
          }
          onSaveDescriptionSuccess={() =>
            toast("Saved description.", "success")
          }
        />
      </section>

      {isRaw && (
        <section>
          <ConfigurationSection
            configuration={data.configuration}
            refetch={refetch}
            onSaveSuccess={() => toast("Saved configuration!", "success")}
            onSaveError={() => toast("Failed to save configuration.", "error")}
          />
        </section>
      )}

      {!isRaw && (
        <CardContainer>
          <Stack spacing={2}>
            <ToggleButtonGroup
              color="primary"
              value={"topology"}
              sx={{
                display: 'flex',
                justifyContent: 'center',
              }}>
              <ToggleButton
                value="topology"
                sx={{
                  display: "flex",
                  justifyContent: "center",
                  paddingLeft: 15,
                  paddingRight: 15,
                  textTransform: "none",
                }}
                disabled
              >
                Topology
              </ToggleButton>
            </ToggleButtonGroup>
            <PipelineGraph
              configuration={data.configuration}
              refetchConfiguration={refetch}
              agent={""}
              rawOrTopology={"topology"}
              yamlValue={""}
            />
          </Stack>
        </CardContainer>
      )}

      {!isRaw && (
        <section>
          <AddSourcesSection
            configuration={data.configuration}
            refetch={refetch}
            setAddDialogOpen={setAddSourceDialogOpen}
            addDialogOpen={addSourceDialogOpen}
          />
        </section>
      )}

      {!isRaw && (
        <section>
          <AddDestinationsSection
            configuration={data.configuration}
            destinations={data.configuration.spec.destinations ?? []}
            refetch={refetch}
            setAddDialogOpen={setAddDestDialogOpen}
            addDialogOpen={addDestDialogOpen}
          />
        </section>
      )}

      <section>
        <CardContainer>
          <div className={styles["title-button-row"]}>
            <Typography variant="h5">Agents</Typography>
            {!platformIsKubernetes(data.configuration?.metadata?.labels?.platform) && (
              <IconButton onClick={openApplyDialog} color="primary">
                <PlusCircleIcon />
            </IconButton>
            )}
          </div>

          <AgentsTable
            selector={selectorString(data.configuration.spec.selector)}
            columnFields={[
              AgentsTableField.NAME,
              AgentsTableField.STATUS,
              AgentsTableField.OPERATING_SYSTEM,
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
          onError={() => toast("Failed to apply configuration.", "error")}
          onSuccess={onApplySuccess}
          onClose={closeApplyDialog}
          onCancel={closeApplyDialog}
        />
      )}
    </ConfigurationPageContextProvider>
  );
};

export const ViewConfiguration = withRequireLogin(
  withNavBar(ConfigPageContent)
);
