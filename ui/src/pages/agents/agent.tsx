import { gql } from "@apollo/client";
import {
  Dialog,
  DialogContent,
  Grid,
  Stack,
  Typography,
  Alert,
  AlertTitle,
  Button,
  Tooltip,
  Box,
} from "@mui/material";
import { useEffect, useMemo, useState } from "react";
import { useNavigate, useParams } from "react-router-dom";
import {
  BorderlessCardContainer,
  CardContainer,
} from "../../components/CardContainer";
import { ManageConfigForm } from "../../components/ManageConfigForm";
import { AgentTable } from "../../components/Tables/AgentTable";
import {
  useGetAgentAndConfigurationsQuery,
  useGetConfigurationQuery,
} from "../../graphql/generated";
import { useAgentChangesContext } from "../../hooks/useAgentChanges";
import { RawConfigWizard } from "../configurations/wizards/RawConfigWizard";
import { useSnackbar } from "notistack";
import { labelAgents } from "../../utils/rest/label-agents";
import { RawConfigFormValues } from "../../types/forms";
import {
  hasAgentFeature,
  AgentFeatures,
  AgentStatus,
} from "../../types/agents";
import { withRequireLogin } from "../../contexts/RequireLogin";
import { withNavBar } from "../../components/NavBar";
import { AgentChangesProvider } from "../../contexts/AgentChanges";
import { RecentTelemetryDialog } from "../../components/RecentTelemetryDialog/RecentTelemetryDialog";
import { PipelineGraph } from "../../components/PipelineGraph/PipelineGraph";
import { renderAgentStatus } from "../../components/Tables/utils";
import { RawOrTopologyControl } from "../../components/PipelineGraph/RawOrTopologyControl";
import { Config } from "../../components/ManageConfigForm/types";
import { ConfigurationSelect } from "../../components/ManageConfigForm/ConfigurationSelect";
import { classes } from "../../utils/styles";
import { YamlEditor } from "../../components/YamlEditor";
import {
  DEFAULT_PERIOD,
  DEFAULT_TELEMETRY_TYPE,
  MeasurementControlBar,
} from "../../components/MeasurementControlBar";
import { UpgradeError } from "../../components/UpgradeError";

import mixins from "../../styles/mixins.module.scss";

gql`
  query GetAgentAndConfigurations($agentId: ID!) {
    agent(id: $agentId) {
      id
      name
      architecture
      operatingSystem
      labels
      hostName
      platform
      version
      macAddress
      remoteAddress
      home
      status
      connectedAt
      disconnectedAt
      errorMessage
      configuration {
        Collector
      }
      configurationResource {
        metadata {
          id
          version
          name
        }
      }
      upgrade {
        status
        version
        error
      }
      upgradeAvailable
      features
    }
    configurations {
      configurations {
        metadata {
          id
          name
          version
          labels
        }
        spec {
          raw
        }
      }
    }
  }
`;

export const AgentPageContent: React.FC = () => {
  const { id } = useParams();
  const snackbar = useSnackbar();
  const [importOpen, setImportOpen] = useState(false);
  const [recentTelemetryOpen, setRecentTelemetryOpen] = useState(false);
  const [selectedTelemetry, setSelectedTelemetry] = useState(
    DEFAULT_TELEMETRY_TYPE
  );
  const [period, setPeriod] = useState(DEFAULT_PERIOD);

  // AgentChanges subscription to trigger a refetch.
  const agentChanges = useAgentChangesContext();

  const { data, refetch } = useGetAgentAndConfigurationsQuery({
    variables: { agentId: id ?? "" },
    fetchPolicy: "network-only",
  });

  const navigate = useNavigate();

  async function onImportSuccess(values: RawConfigFormValues) {
    if (data?.agent != null) {
      try {
        await labelAgents(
          [data.agent.id],
          { configuration: values.name },
          true
        );
      } catch (err) {
        snackbar.enqueueSnackbar("Failed to apply label to agent.", {
          variant: "error",
        });
      }
    }

    setImportOpen(false);
  }

  useEffect(() => {
    if (agentChanges.length > 0) {
      const thisAgent = agentChanges
        .map((c) => c.agent)
        .find((a) => a.id === id);
      if (thisAgent != null) {
        refetch();
      }
    }
  }, [agentChanges, id, refetch]);

  const currentConfig = useMemo(() => {
    if (data?.agent == null || data?.configurations == null) {
      return null;
    }

    const configName = data.agent.configurationResource?.metadata.name;
    if (configName == null) {
      return null;
    }

    return data.configurations.configurations.find(
      (c) => c.metadata.name === configName
    );
  }, [data?.agent, data?.configurations]);

  const currentConfigName = useMemo(() => {
    if (data?.agent == null || data?.configurations == null) {
      return null;
    }

    const configName = data.agent.configurationResource?.metadata.name;
    if (configName == null) {
      return null;
    }
    return configName;
  }, [data?.agent, data?.configurations]);

  // Get Configuration Data

  const configQuery = useGetConfigurationQuery({
    variables: { name: currentConfigName ?? "" },
    fetchPolicy: "cache-and-network",
  });
  const configGraph = configQuery.data?.configuration;
  const [rawOrTopologyTab, setRawOrTopologyTab] = useState<"topology" | "raw">(
    "topology"
  );
  const [editing, setEditing] = useState(false);

  const [selectedConfig, setSelectedConfig] = useState<Config | undefined>(
    data?.configurations.configurations.find(
      (c) =>
        c.metadata.name === data?.agent?.configurationResource?.metadata.name
    )
  );

  const isRaw =
    configGraph?.spec?.raw != null && configGraph?.spec?.raw.length > 0;

  useEffect(() => {
    setRawOrTopologyTab(isRaw ? "raw" : "topology");
  }, [isRaw]);

  const viewTelemetryButton = useMemo(() => {
    if (currentConfig?.spec.raw !== "") {
      return null;
    }

    let disableReason: string | null = null;

    if (
      data?.agent == null ||
      data.agent?.status === AgentStatus.DISCONNECTED
    ) {
      disableReason = "Cannot view recent telemetry, agent is disconnected.";
    }

    if (
      disableReason == null &&
      !hasAgentFeature(data!.agent!, AgentFeatures.AGENT_SUPPORTS_SNAPSHOTS)
    ) {
      disableReason =
        "Upgrade Agent to v1.8.0 or later to view recent telemetry.";
    }

    if (disableReason == null && data?.agent?.configurationResource == null) {
      disableReason =
        "Cannot view recent telemetry for an agent with an unmanaged configuration.";
    }

    if (disableReason != null) {
      return (
        <Tooltip title={disableReason} disableInteractive>
          <div style={{ display: "inline-block" }}>
            <Button
              variant="contained"
              size="large"
              onClick={() => setRecentTelemetryOpen(true)}
              disabled
            >
              View Recent Telemetry
            </Button>
          </div>
        </Tooltip>
      );
    } else {
      return (
        <Button
          variant="contained"
          size="large"
          onClick={() => setRecentTelemetryOpen(true)}
        >
          View Recent Telemetry
        </Button>
      );
    }
  }, [currentConfig?.spec.raw, data]);

  // Here we use the distinction between graphql returning null vs undefined.
  // If the agent is null then this agent doesn't exist, redirect to agents.
  if (data?.agent === null) {
    navigate("/agents");
    return null;
  }

  // Data is loading, return null for now.
  if (data === undefined || data.agent == null) {
    return null;
  }

  const upgradeError = data.agent?.upgrade?.error;

  const EditConfiguration: React.FC = () => {
    return (
      <>
        <ConfigurationSelect
          agent={data?.agent!}
          setSelectedConfig={setSelectedConfig}
          selectedConfig={selectedConfig}
          configurations={data.configurations?.configurations}
        />
      </>
    );
  };
  return (
    <>
      <BorderlessCardContainer>
        <Grid container spacing={1} alignItems="center">
          <Grid item xs="auto" lg="auto">
            <Typography variant="h5">Agent - {data.agent.name}</Typography>
          </Grid>
          <Grid item xs="auto" lg="auto" alignItems="center">
            {renderAgentStatus(data.agent.status)}
          </Grid>
          <Grid item style={{ flex: 1 }} />
          <Grid item xs="auto" lg="auto">
            {viewTelemetryButton}
          </Grid>
        </Grid>
      </BorderlessCardContainer>

      <CardContainer>
        <Grid container spacing={5}>
          <Grid item xs={12} lg={12}>
            <Box
              sx={{ borderBottom: 1, ml: -3, mr: -3, borderColor: "divider" }}
            >
              <Box sx={{ ml: 3, mr: 3 }}>
                <Typography variant="h6" classes={{ root: mixins["mb-3"] }}>
                  Details
                </Typography>
              </Box>
            </Box>
          </Grid>

          {upgradeError && (
            <Grid item xs={12}>
              <UpgradeError
                agentId={data.agent.id}
                upgradeError={data.agent?.upgrade?.error}
                onClearFailure={() => {
                  snackbar.enqueueSnackbar("Oops! Something went wrong.", {
                    variant: "error",
                    key: "clear-upgrade-error",
                  });
                }}
                onClearSuccess={() => {
                  refetch();
                }}
              />
            </Grid>
          )}
          <Grid item xs={12} lg={12}>
            <AgentTable agent={data.agent} />
          </Grid>
        </Grid>
      </CardContainer>

      <CardContainer>
        <Stack spacing={2}>
          {/* Edit configuration */}
          <Box
            sx={{
              borderBottom: 1,
              ml: -3,
              mr: -3,
              borderColor: "divider",
            }}
          >
            <Box sx={{ ml: 3, mr: 3 }}>
              <ManageConfigForm
                agent={data.agent}
                configurations={data.configurations.configurations ?? []}
                onImport={() => setImportOpen(true)}
                editing={editing}
                setEditing={setEditing}
                selectedConfig={selectedConfig}
                setSelectedConfig={setSelectedConfig}
              />
            </Box>
          </Box>
          {editing && (
            <>
              <EditConfiguration />
              <Box sx={{ minHeight: 400 }}></Box>
            </>
          )}
          {/* Toggle topology/raw */}
          {configGraph && !editing && !isRaw && (
            <RawOrTopologyControl
              rawOrTopology={rawOrTopologyTab}
              setTopologyOrRaw={setRawOrTopologyTab}
            />
          )}
          {data.agent.errorMessage && !editing && (
            <Alert
              severity="error"
              className={classes([mixins["mt-5"], mixins["mb-0"]])}
            >
              <AlertTitle>Error</AlertTitle>
              {data.agent.errorMessage}
            </Alert>
          )}
          {/* Graph or YAML */}
          {configGraph && !editing && (
            <>
              {rawOrTopologyTab === "topology" ? (
                <div>
                  <MeasurementControlBar
                    telemetry={selectedTelemetry}
                    onTelemetryTypeChange={setSelectedTelemetry}
                    period={period}
                    onPeriodChange={setPeriod}
                  />
                  <PipelineGraph
                    agentId={data.agent.id}
                    configurationName={`${data.agent.configurationResource?.metadata.name}:${data.agent.configurationResource?.metadata.version}`}
                    selectedTelemetry={selectedTelemetry}
                    period={period}
                    readOnly
                  />
                </div>
              ) : (
                <YamlEditor
                  value={data.agent.configuration?.Collector || ""}
                  readOnly
                  limitHeight
                />
              )}
            </>
          )}
        </Stack>
      </CardContainer>

      {/** Raw Config wizard for importing an agents config */}
      <Dialog
        open={importOpen}
        onClose={() => setImportOpen(false)}
        PaperComponent={EmptyComponent}
        scroll={"body"}
      >
        <DialogContent>
          <Stack justifyContent="center" alignItems="center" height="100%">
            <RawConfigWizard
              onClose={() => setImportOpen(false)}
              initialValues={{
                name: data.agent.name,
                description: `Imported config from agent ${data.agent.name}.`,
                fileName: "",
                rawConfig: data.agent.configuration?.Collector ?? "",
                platform: configPlatformFromAgentPlatform(data.agent.platform),
                secondaryPlatform: configSecondaryPlatformFromAgentPlatform(
                  data.agent.platform
                ),
              }}
              onSuccess={onImportSuccess}
              fromImport
            />
          </Stack>
        </DialogContent>
      </Dialog>

      {currentConfig?.spec.raw === "" && (
        <RecentTelemetryDialog
          open={recentTelemetryOpen}
          onClose={() => setRecentTelemetryOpen(false)}
          agentID={id!}
        />
      )}
    </>
  );
};

const EmptyComponent: React.FC = ({ children }) => {
  return <>{children}</>;
};

function configPlatformFromAgentPlatform(platform: string | null | undefined) {
  if (platform == null) return "linux";
  if (platform === "darwin") return "macos";
  if (platform.startsWith("kubernetes")) return "kubernetes";
  if (platform.startsWith("openshift")) return "openshift";
  return platform;
}

function configSecondaryPlatformFromAgentPlatform(
  platform: string | null | undefined
) {
  if (platform == null) return "";
  if (platform.startsWith("kubernetes")) return platform;
  if (platform.startsWith("openshift")) return platform;
  return "";
}

export const AgentPage = withRequireLogin(
  withNavBar(() => (
    <AgentChangesProvider>
      <AgentPageContent />
    </AgentChangesProvider>
  ))
);
