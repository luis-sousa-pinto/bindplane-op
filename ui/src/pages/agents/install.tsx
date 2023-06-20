import { gql } from "@apollo/client";
import {
  Box,
  Divider,
  FormControl,
  InputLabel,
  Link,
  MenuItem,
  Select,
  SelectChangeEvent,
  Stack,
  Typography,
} from "@mui/material";
import React, { useEffect, useState } from "react";
import { CardContainer } from "../../components/CardContainer";
import { CodeBlock } from "../../components/CodeBlock";
import {
  GetConfigurationNamesQuery,
  useGetConfigurationNamesQuery,
} from "../../graphql/generated";
import { InstallCommandResponse } from "../../types/rest";
import { withRequireLogin } from "../../contexts/RequireLogin";
import { withNavBar } from "../../components/NavBar";
import { PlatformSelect } from "../../components/PlatformSelect";

import mixins from "../../styles/mixins.module.scss";
import { isEmpty } from "lodash";

gql`
  query GetConfigurationNames {
    configurations {
      configurations {
        metadata {
          id
          name
          version
          labels
        }
      }
    }
  }
`;

/**
 * Platforms that agents can be installed on
 */
export enum Platform {
  KubernetesDaemonset = "kubernetes-daemonset",
  KubernetesDeployment = "kubernetes-deployment",
  Linux = "linux",
  macOS = "macos",
  OpenShiftDaemonset = "openshift-daemonset",
  OpenShiftDeployment = "openshift-deployment",
  Windows = "windows",
}

export const InstallPageContent: React.FC = () => {
  const [platform, setPlatform] = useState<string>(Platform.Linux);
  const [secondaryPlatform, setSecondaryPlatform] = useState<string>("");
  const [secondaryPlatformRequired, setSecondaryPlatformRequired] =
    useState(false);
  const [installCommand, setCommand] = useState("");
  const [configs, setConfigs] = useState<string[]>([]);
  const [selectedConfig, setSelectedConfig] = useState<string>("");
  const { data } = useGetConfigurationNamesQuery({
    fetchPolicy: "network-only",
  });

  // Don't show the command if the platform is k8s and no config is selected

  function shouldShowCommand() {
    // If the platform has a secondary platform, don't show the command until both are selected
    if (secondaryPlatformRequired && isEmpty(secondaryPlatform)) {
      return false;
    }

    // If the platform is containerized, don't show the command until a config is selected
    if (
      platformIsContainer(
        isEmpty(secondaryPlatform) ? platform : secondaryPlatform
      ) &&
      isEmpty(selectedConfig)
    ) {
      return false;
    }
    return true;
  }

  useEffect(() => {
    if (data) {
      if (secondaryPlatformRequired && isEmpty(secondaryPlatform)) {
        setConfigs([]);
        return;
      }
      // First filter the configs to match the platform
      const filtered = filterConfigurationsByPlatform(
        data.configurations.configurations,
        isEmpty(secondaryPlatform) ? platform : secondaryPlatform
      );

      const configNames = filtered.map((c) => c.metadata.name);

      setConfigs(configNames);
    }
  }, [
    data,
    platform,
    secondaryPlatform,
    secondaryPlatformRequired,
    setConfigs,
  ]);

  useEffect(() => {
    async function fetchInstallText() {
      const actualPlatform = isEmpty(secondaryPlatform)
        ? platform
        : secondaryPlatform;
      // If the platform is k8s, don't show the command until a config is selected
      if (
        (secondaryPlatformRequired && isEmpty(secondaryPlatform)) ||
        (platformIsContainer(actualPlatform) && isEmpty(selectedConfig))
      ) {
        setCommand("");
        return;
      }

      const url = installCommandUrl({
        platform: actualPlatform,
        configuration: selectedConfig,
      });
      const resp = await fetch(url);
      const { command } = (await resp.json()) as InstallCommandResponse;
      if (resp.status === 200) {
        setCommand(command);
      }
    }

    fetchInstallText();
  }, [platform, secondaryPlatform, secondaryPlatformRequired, selectedConfig]);

  return (
    <CardContainer>
      <Typography variant="h5" classes={{ root: mixins["mb-5"] }}>
        Agent Installation
      </Typography>

      <Box
        component="form"
        className={`${mixins["form-width"]} ${mixins["mb-3"]}`}
      >
        <PlatformSelect
          platformValue={platform}
          secondaryPlatformValue={secondaryPlatform}
          helperText="Select the platform the agent will run on."
          onPlatformSelected={(v) => {
            setPlatform(v);
            setSecondaryPlatform("");
            setSelectedConfig("");
          }}
          onSecondaryPlatformSelected={(v) => {
            setSecondaryPlatform(v);
            setSelectedConfig("");
          }}
          setSecondaryPlatformRequired={setSecondaryPlatformRequired}
        />
        <ConfigurationSelect
          configs={configs}
          platform={!isEmpty(secondaryPlatform) ? secondaryPlatform : platform}
          selectedConfig={selectedConfig}
          setSelectedConfig={setSelectedConfig}
        />
      </Box>

      {platformIsKubernetes(secondaryPlatform) && shouldShowCommand() && (
        <Typography fontSize="18px" fontWeight="bold">
          To deploy the agent to Kubernetes:<br></br>
          <Typography fontSize="16px">
            1. Copy the YAML below to a file<br></br>
            2. Apply with kubectl:{" "}
            <code>kubectl apply -f &lt;filename&gt;</code>
          </Typography>
          <br></br>
        </Typography>
      )}

      {platformIsOpenShift(secondaryPlatform) && shouldShowCommand() && (
        <Typography fontSize="18px" fontWeight="bold">
          To deploy the agent to OpenShift:<br></br>
          <Typography fontSize="16px">
            1. Copy the YAML below to a file<br></br>
            2. Apply with oc: <code>oc apply -f &lt;filename&gt;</code>
          </Typography>
          <br></br>
        </Typography>
      )}
      {shouldShowCommand() && <CodeBlock value={installCommand} />}
    </CardContainer>
  );
};

interface configurationSelectProps {
  platform: string;
  configs: string[];
  selectedConfig: string;
  setSelectedConfig: (config: string) => void;
}

/**
 * Renders a select box for selecting a configuration depending on the platform
 * k8s and openshift require a configuration, others do not
 *
 * @param configs - The list of configurations to display
 * @param platform - The platform to filter the configurations by
 * @param selectedConfig - The currently selected configuration
 * @param setSelectedConfig - The function to call when the configuration is changed
 */
const ConfigurationSelect: React.FC<configurationSelectProps> = ({
  configs,
  platform,
  selectedConfig,
  setSelectedConfig,
}: configurationSelectProps) => {
  const configRequired = platformIsContainer(platform);
  const label = configRequired
    ? "Select Configuration"
    : "Select Configuration (optional)";

  return (
    <Stack spacing={2}>
      {configs.length > 0 && (
        <FormControl fullWidth margin="normal">
          <InputLabel id="config-label">{label}</InputLabel>

          <Select
            inputProps={{ "data-testid": "config-select" }}
            labelId="config-label"
            id="configuration"
            label={label}
            onChange={(e: SelectChangeEvent<string>) => {
              setSelectedConfig(e.target.value);
            }}
            value={selectedConfig}
          >
            {!configRequired && (
              <MenuItem value="">
                <em>None</em>
              </MenuItem>
            )}
            {configs.map((c) => (
              <MenuItem key={c} value={c} data-testid={`config-${c}`}>
                {c}
              </MenuItem>
            ))}
          </Select>
        </FormControl>
      )}
      {configs.length === 0 && configRequired && (
        <>
          <Divider />
          <Typography>
            No configurations found for this platform. Please create a{" "}
            <Link href="/configurations/new">configuration</Link> before
            deploying the agent.
          </Typography>
        </>
      )}
    </Stack>
  );
};

function installCommandUrl(params: {
  platform: string;
  configuration?: string;
}): string {
  const url = new URL(window.location.href);
  url.pathname = "/v1/agent-versions/latest/install-command";

  const searchParams: { platform: string; labels?: string } = {
    platform: params.platform,
  };

  if (params.configuration) {
    searchParams.labels = encodeURI(`configuration=${params.configuration}`);
  }

  url.search = new URLSearchParams(searchParams).toString();
  return url.href;
}

function filterConfigurationsByPlatform(
  configs: GetConfigurationNamesQuery["configurations"]["configurations"],
  platform: string
): GetConfigurationNamesQuery["configurations"]["configurations"] {
  switch (platform) {
    case Platform.KubernetesDaemonset:
      return configs.filter(
        (c) => c.metadata.labels?.platform === Platform.KubernetesDaemonset
      );
    case Platform.KubernetesDeployment:
      return configs.filter(
        (c) => c.metadata.labels?.platform === Platform.KubernetesDeployment
      );
    case Platform.Linux:
      return configs.filter((c) => c.metadata.labels?.platform === "linux");
    case Platform.macOS:
      return configs.filter((c) => c.metadata.labels?.platform === "macos");
    case Platform.OpenShiftDaemonset:
      return configs.filter(
        (c) => c.metadata.labels?.platform === Platform.OpenShiftDaemonset
      );
    case Platform.OpenShiftDeployment:
      return configs.filter(
        (c) => c.metadata.labels?.platform === Platform.OpenShiftDeployment
      );
    case Platform.Windows:
      return configs.filter((c) => c.metadata.labels?.platform === "windows");
    default:
      return configs;
  }
}

/**
 * Check if the platform is a k8s platform
 * @param platform Reported platform
 */
export function platformIsKubernetes(platform: string): boolean {
  return (
    platform === Platform.KubernetesDaemonset ||
    platform === Platform.KubernetesDeployment
  );
}

/**
 * Check if the platform is an OpenShift platform
 * @param platform Reported platform
 */
export function platformIsOpenShift(platform: string): boolean {
  return (
    platform === Platform.OpenShiftDaemonset ||
    platform === Platform.OpenShiftDeployment
  );
}

/**
 * Check if the platform is a container platform
 * @param platform Reported platform
 */
export function platformIsContainer(platform: string): boolean {
  return platformIsKubernetes(platform) || platformIsOpenShift(platform);
}

export const InstallPage = withRequireLogin(withNavBar(InstallPageContent));
