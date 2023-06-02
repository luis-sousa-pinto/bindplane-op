import { PipelineGraph } from "../PipelineGraph/PipelineGraph";
import {
  ConfigurationVersionSwitcher,
  ConfigurationVersionSwitcherTab,
} from "../ConfigurationVersionSwitcher";
import {
  DEFAULT_PERIOD,
  DEFAULT_TELEMETRY_TYPE,
  MeasurementControlBar,
} from "../MeasurementControlBar";
import { useState } from "react";
import { gql } from "@apollo/client";
import { useGetConfigurationVersionsQuery } from "../../graphql/generated";
import { useSnackbar } from "notistack";
import { VersionsData } from "./util";
import { RolloutProgress } from "../RolloutProgress";
import { OtelConfigEditor } from "../OtelConfigEditor/OtelConfigEditor";
import { nameAndVersion } from "../../utils/version-helpers";
import { useRefetchOnConfigurationChange } from "../../hooks/useRefetchOnConfigurationChanges";

gql`
  query getConfigurationVersions($name: String!) {
    configurationHistory(name: $name) {
      metadata {
        name
        id
        version
      }

      activeTypes

      status {
        current
        pending
        latest
      }
    }
  }
`;

interface ConfigurationEditorProps {
  configurationName: string;
  isOtel: boolean;
  hideRolloutActions?: boolean;
}

/**
 * ConfigurationEditor is a component used to edit and show information about a
 * BPOP Configuration.  It can show a YAML editor for an oTel configuration
 * or a pipeline graph for a pipeline configuration.
 *
 * @param configurationName should be the non-versioned name of the config
 * @param isOtel is a boolean that determines whether to display a PipelineGraph
 * or OtelConfig component.
 * @param hideRolloutActions is a boolean that determines whether to display the
 * rollout actions (i.e pause, resume, start)
 * @returns
 */
export const ConfigurationEditor: React.FC<ConfigurationEditorProps> = ({
  configurationName,
  isOtel,
  hideRolloutActions,
}) => {
  const { enqueueSnackbar } = useSnackbar();

  const [versionsData, setVersionsData] = useState<VersionsData>();
  const [selectedTelemetry, setSelectedTelemetry] = useState<string>(
    DEFAULT_TELEMETRY_TYPE
  );
  const [period, setPeriod] = useState<string>(DEFAULT_PERIOD);
  const [tab, setTab] = useState<ConfigurationVersionSwitcherTab>();
  const [selectedVersion, setSelectedVersion] = useState<number>();
  const [editingCurrentVersion, setEditingCurrentVersion] =
    useState<boolean>(false);
  const [editingPendingVersion, setEditingPendingVersion] =
    useState<boolean>(false);

  const { refetch } = useGetConfigurationVersionsQuery({
    variables: {
      name: configurationName,
    },
    onError(error) {
      console.error(error);
      enqueueSnackbar("Failed to fetch configuration versions.", {
        variant: "error",
      });
    },
    onCompleted(data) {
      const newVersionsData = new VersionsData(data);
      if (newVersionsData.findNew()) {
        setTab("new");
        setEditingCurrentVersion(false);
        setEditingPendingVersion(false);
      } else if (newVersionsData.findPending()) {
        setTab("pending");
      } else if (newVersionsData.findCurrent()) {
        setTab("current");
      }
      setSelectedVersion(newVersionsData.latestHistoryVersion());
      setVersionsData(newVersionsData);

      setSelectedTelemetry(
        newVersionsData.firstActiveType() ?? DEFAULT_TELEMETRY_TYPE
      );
    },
  });

  useRefetchOnConfigurationChange(configurationName, refetch);

  // TODO(dsvanlani): Add a loading state
  if (tab == null || versionsData == null) {
    return null;
  }

  function handleOnEditCurrentVersion() {
    // If we have a new version, switch user to the new tab
    if (versionsData?.findNew()) {
      setTab("new");
      return;
    } else {
      // Otherwise we will make an editable graph with version latest
      setEditingCurrentVersion(true);
      setTab("new");
    }
  }

  function handleEditPendingVersion() {
    setEditingPendingVersion(true);
    setTab("new");
  }

  const { newVersion, currentVersion, pendingVersion } =
    versionsData.versionMap();

  const EditorComponent = isOtel ? OtelConfigEditor : PipelineGraph;

  var editingVersion: number | undefined;
  if (editingCurrentVersion) {
    editingVersion = currentVersion;
  } else if (editingPendingVersion) {
    editingVersion = pendingVersion;
  } else {
    editingVersion = newVersion;
  }

  return (
    <>
      <ConfigurationVersionSwitcher
        tab={tab}
        onSelectedVersionHistoryChange={setSelectedVersion}
        onChange={setTab}
        onEditNewVersion={handleOnEditCurrentVersion}
        allowEditPendingVersion={versionsData.findNew() == null}
        onEditPendingVersion={handleEditPendingVersion}
        versionHistory={versionsData
          .versionHistory()
          .map((v) => v.metadata.version)}
        selectedVersionHistory={selectedVersion}
        newVersion={editingVersion}
        currentVersion={currentVersion}
        pendingVersion={pendingVersion}
      />
      {!isOtel && (
        <MeasurementControlBar
          telemetry={selectedTelemetry!}
          onTelemetryTypeChange={setSelectedTelemetry}
          period={period}
          onPeriodChange={setPeriod}
        />
      )}
      {tab === "current" && (
        <EditorComponent
          configurationName={nameAndVersion(configurationName, currentVersion)}
          selectedTelemetry={selectedTelemetry!}
          period={period}
          readOnly
        />
      )}

      {tab === "history" && (
        <EditorComponent
          selectedTelemetry={selectedTelemetry!}
          period={period}
          configurationName={nameAndVersion(configurationName, selectedVersion)}
          skipMeasurements
          readOnly
        />
      )}

      {tab === "pending" && (
        <EditorComponent
          selectedTelemetry={selectedTelemetry!}
          period={period}
          configurationName={nameAndVersion(configurationName, pendingVersion)}
          skipMeasurements
          readOnly
        />
      )}

      {tab === "new" && (
        <EditorComponent
          selectedTelemetry={selectedTelemetry!}
          period={period}
          configurationName={nameAndVersion(configurationName, editingVersion)}
          skipMeasurements
        />
      )}

      {tab !== "history" && (
        <RolloutProgress
          configurationName={configurationName}
          configurationVersion={tab === "new" ? "latest" : tab}
          hideActions={hideRolloutActions}
        />
      )}
    </>
  );
};
