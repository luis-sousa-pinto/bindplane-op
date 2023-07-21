import { useSnackbar } from "notistack";
import { usePipelineGraph } from "../PipelineGraph/PipelineGraphContext";
import { useMemo, useState } from "react";
import { useRole } from "../../hooks/useRole";
import { gql } from "@apollo/client";
import { Role, useGetSourceWithTypeQuery } from "../../graphql/generated";
import { FormValues } from "../ResourceConfigForm";
import { BPSource } from "../../utils/classes/source";
import { UpdateStatus } from "../../types/resources";
import { onDeleteFunc } from "./types";
import { BPConfiguration } from "../../utils/classes";
import { ResourceCard } from "./ResourceCard";
import { EditResourceDialog } from "../ResourceDialog/EditResourceDialog";
import { ConfirmDeleteResourceDialog } from "../ConfirmDeleteResourceDialog";
import { Typography } from "@mui/material";
import { hasPermission } from "../../utils/has-permission";
import { classes } from "../../utils/styles";

import styles from "./cards.module.scss";
import { trimVersion } from "../../utils/version-helpers";

gql`
  query getSourceWithType($name: String!) {
    sourceWithType(name: $name) {
      source {
        metadata {
          name
          version
          id
          labels
          version
        }
        spec {
          type
          parameters {
            name
            value
          }
          disabled
        }
      }
      sourceType {
        metadata {
          id
          name
          version
          icon
          description
        }
        spec {
          parameters {
            label
            name
            description
            required
            type
            default
            relevantIf {
              name
              operator
              value
            }
            documentation {
              text
              url
            }
            advancedConfig
            validValues
            options {
              multiline
              creatable
              trackUnchecked
              sectionHeader
              gridColumns
              labels
              metricCategories {
                label
                column
                metrics {
                  name
                  description
                  kpi
                }
              }
              password
              sensitive
            }
          }
        }
      }
    }
  }
`;

export const ResourceSourceCard: React.FC<{
  name: string;
  sourceIndex: number;
  disabled?: boolean;
}> = ({ name, sourceIndex, disabled }) => {
  const { configuration, refetchConfiguration, readOnlyGraph } =
    usePipelineGraph();
  const { enqueueSnackbar } = useSnackbar();
  const [editing, setEditing] = useState(false);
  const [confirmDeleteOpen, setDeleteOpen] = useState(false);
  const role = useRole();

  // Use the version name of the source specified in the configuration
  const versionedName = useMemo(() => {
    const sources = configuration?.spec?.sources;
    const source = sources?.[sourceIndex];

    return source?.name;
  }, [configuration?.spec?.sources, sourceIndex]);

  const { data, refetch: refetchSource } = useGetSourceWithTypeQuery({
    variables: { name: versionedName || name },
    fetchPolicy: "cache-and-network",
  });

  async function onSave(formValues: FormValues) {
    if (data?.sourceWithType?.source == null) {
      enqueueSnackbar("Cannot save source.", { variant: "error" });
      console.error("no source found when requestion source and type");
      return;
    }

    const updatedSource = new BPSource(data.sourceWithType.source);
    updatedSource.setParamsFromMap(formValues);

    try {
      const update = await updatedSource.apply();
      if (update.status === UpdateStatus.INVALID) {
        console.error("Invalid Update: ", update);
        throw new Error(`failed to apply source, got status ${update.status}`);
      }

      enqueueSnackbar("Saved Source! 🎉", { variant: "success" });

      refetchConfiguration();
      refetchSource();
      setEditing(false);
    } catch (err) {
      console.error(err);
      enqueueSnackbar("Failed to update source.", { variant: "error" });
    }
  }

  const onDelete: onDeleteFunc | undefined = useMemo(() => {
    if (sourceIndex == null) {
      return undefined;
    }

    return async function onDelete() {
      const updatedConfig = new BPConfiguration(configuration);
      updatedConfig.removeSource(sourceIndex);

      try {
        const update = await updatedConfig.apply();
        if (update.status === UpdateStatus.INVALID) {
          console.error("Invalid Update: ", update);
          throw new Error(
            `failed to apply configuration, got status ${update.status}`
          );
        }

        setEditing(false);
        setDeleteOpen(false);
        refetchConfiguration();
        refetchSource();
      } catch (err) {
        console.error(err);
        enqueueSnackbar("Failed to remove source.", { variant: "error" });
      }
    };
  }, [
    configuration,
    enqueueSnackbar,
    refetchConfiguration,
    refetchSource,
    sourceIndex,
  ]);

  async function onTogglePause() {
    if (data?.sourceWithType?.source == null) {
      enqueueSnackbar("Cannot save source.", { variant: "error" });
      console.error("no source found when requestion source and type");
      return;
    }

    if (sourceIndex == null) {
      enqueueSnackbar("Cannot save source.", { variant: "error" });
      console.error("no source index found");
      return;
    }

    const updatedSource = new BPSource(data.sourceWithType.source);

    updatedSource.toggleDisabled();

    const action = updatedSource.spec.disabled ? "pause" : "resume";

    try {
      const update = await updatedSource.apply();
      if (update.status === UpdateStatus.INVALID) {
        console.error("Invalid Update: ", update);
        throw new Error(
          `failed to apply configuration, got status ${update.status}`
        );
      }

      enqueueSnackbar(`Source ${action}d! 🎉`, { variant: "success" });

      setEditing(false);
      refetchConfiguration();
      refetchSource();
    } catch (err) {
      console.error(err);
      enqueueSnackbar(`Failed to ${action} source.`, { variant: "error" });
    }
  }

  // Loading
  if (data === undefined) {
    return null;
  }

  if (
    data.sourceWithType.source == null ||
    data.sourceWithType.source.spec.type == null
  ) {
    enqueueSnackbar(`Failed to find Source or Type for Source ${name}`, {
      variant: "error",
    });

    return null;
  }

  return (
    <div
      className={classes([
        disabled ? styles.disabled : undefined,
        data.sourceWithType.source?.spec.disabled ? styles.paused : undefined,
      ])}
    >
      <ResourceCard
        name={trimVersion(name)}
        icon={data.sourceWithType?.sourceType?.metadata.icon}
        paused={data.sourceWithType.source?.spec.disabled}
        disabled={data.sourceWithType.source?.spec.disabled}
        onClick={() => setEditing(true)}
      />
      <EditResourceDialog
        kind="source"
        resourceTypeDisplayName={name}
        description={data.sourceWithType.sourceType?.metadata.description ?? ""}
        fullWidth
        maxWidth="sm"
        parameters={data.sourceWithType.source.spec.parameters ?? []}
        parameterDefinitions={
          data.sourceWithType.sourceType?.spec.parameters ?? []
        }
        open={editing}
        onClose={() => setEditing(false)}
        onCancel={() => setEditing(false)}
        onDelete={onDelete && (() => setDeleteOpen(true))}
        onSave={onSave}
        paused={data.sourceWithType.source?.spec.disabled}
        onTogglePause={onTogglePause}
        readOnly={readOnlyGraph || !hasPermission(Role.User, role)}
      />

      {onDelete && (
        <ConfirmDeleteResourceDialog
          open={confirmDeleteOpen}
          onClose={() => setDeleteOpen(false)}
          onCancel={() => setDeleteOpen(false)}
          onDelete={onDelete}
          action={"remove"}
        >
          <Typography>Are you sure you want to remove this source?</Typography>
        </ConfirmDeleteResourceDialog>
      )}
    </div>
  );
};
