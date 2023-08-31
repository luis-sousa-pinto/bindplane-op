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
import {
  useGetConfigurationVersionsQuery,
  useGetLatestMeasurementIntervalQuery,
} from "../../graphql/generated";
import { useSnackbar } from "notistack";
import { VersionsData } from "./versions-data";
import { RolloutProgress } from "../RolloutProgress";
import { OtelConfigEditor } from "../OtelConfigEditor/OtelConfigEditor";
import { asCurrentVersion, nameAndVersion } from "../../utils/version-helpers";
import { useRefetchOnConfigurationChange } from "../../hooks/useRefetchOnConfigurationChanges";
import { DiffDialog } from "../DiffDialog/DiffDialog";

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

  query getLatestMeasurementInterval($name: String!) {
    configuration(name: $name) {
      metadata {
        name
        id
        version
      }

      spec {
        measurementInterval
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
  const [period, setPeriod] = useState<string>();
  const [measurementPeriods, setMeasurementPeriods] = useState<string[]>();
  const [tab, setTab] = useState<ConfigurationVersionSwitcherTab>();
  const [selectedVersion, setSelectedVersion] = useState<number>();
  const [editingCurrentVersion, setEditingCurrentVersion] =
    useState<boolean>(false);
  const [editingPendingVersion, setEditingPendingVersion] =
    useState<boolean>(false);
  const [diffDialogOpen, setDiffDialogOpen] = useState<boolean>(false);

  const [showCompareVersions, setShowCompareVersions] =
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

  const { refetch: refetchMI } = useGetLatestMeasurementIntervalQuery({
    variables: {
      name: asCurrentVersion(configurationName),
    },
    onCompleted(data) {
      if (data.configuration?.spec?.measurementInterval != null) {
        switch (data.configuration.spec.measurementInterval) {
          case "1m":
            setMeasurementPeriods(["1m", "5m", "1h", "24h"]);
            setPeriod("1m");
            break;
          case "5m":
            setMeasurementPeriods(["5m", "1h", "24h"]);
            setPeriod("5m");
            break;
          case "1h":
            setMeasurementPeriods(["1h", "24h"]);
            setPeriod("1h");
            break;
          case "24h":
            setMeasurementPeriods(["24h"]);
            setPeriod("24h");
            break;
          default:
            setMeasurementPeriods(["10s", "1m", "5m", "1h", "24h"]);
            setPeriod(DEFAULT_PERIOD);
        }
      }
    },
  });

  useRefetchOnConfigurationChange(configurationName, () => {
    refetch();
    refetchMI();
  });

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
        onCompareVersions={() => setDiffDialogOpen(true)}
        showCompareVersions={showCompareVersions}
      />
      {!isOtel && (
        <MeasurementControlBar
          telemetry={selectedTelemetry!}
          onTelemetryTypeChange={setSelectedTelemetry}
          period={period ?? DEFAULT_PERIOD}
          onPeriodChange={setPeriod}
          periods={measurementPeriods}
        />
      )}
      {tab === "current" && (
        <EditorComponent
          configurationName={nameAndVersion(configurationName, currentVersion)}
          selectedTelemetry={selectedTelemetry!}
          period={period ?? DEFAULT_PERIOD}
          readOnly
        />
      )}

      {tab === "history" && (
        <EditorComponent
          selectedTelemetry={selectedTelemetry!}
          period={period ?? DEFAULT_PERIOD}
          configurationName={nameAndVersion(configurationName, selectedVersion)}
          skipMeasurements
          readOnly
        />
      )}

      {tab === "pending" && (
        <EditorComponent
          selectedTelemetry={selectedTelemetry!}
          period={period ?? DEFAULT_PERIOD}
          configurationName={nameAndVersion(configurationName, pendingVersion)}
          skipMeasurements
          readOnly
        />
      )}

      {tab === "new" && (
        <EditorComponent
          selectedTelemetry={selectedTelemetry!}
          period={period ?? DEFAULT_PERIOD}
          configurationName={nameAndVersion(configurationName, editingVersion)}
          skipMeasurements
        />
      )}

      {tab !== "history" && (
        <RolloutProgress
          configurationName={configurationName}
          configurationVersion={tab === "new" ? "latest" : tab}
          hideActions={hideRolloutActions}
          setShowCompareVersions={setShowCompareVersions}
        />
      )}

      <DiffDialog
        onClose={() => setDiffDialogOpen(false)}
        configurationName={configurationName}
        open={diffDialogOpen}
      />
    </>
  );
};
