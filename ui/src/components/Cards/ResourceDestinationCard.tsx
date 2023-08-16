import { gql } from "@apollo/client";
import { Typography } from "@mui/material";
import { useSnackbar } from "notistack";
import { memo, useMemo, useState } from "react";
import { ConfirmDeleteResourceDialog } from "../ConfirmDeleteResourceDialog";
import { EditResourceDialog } from "../ResourceDialog/EditResourceDialog";
import { Role, useGetDestinationWithTypeQuery } from "../../graphql/generated";
import { UpdateStatus } from "../../types/resources";
import { BPConfiguration, BPDestination } from "../../utils/classes";
import { FormValues } from "../ResourceConfigForm";
import { classes } from "../../utils/styles";
import { usePipelineGraph } from "../PipelineGraph/PipelineGraphContext";
import { trimVersion } from "../../utils/version-helpers";
import { ResourceCard } from "./ResourceCard";
import { hasPermission } from "../../utils/has-permission";
import { useRole } from "../../hooks/useRole";
import { onDeleteFunc } from "./types";

import styles from "./cards.module.scss";

gql`
  query getDestinationWithType($name: String!) {
    destinationWithType(name: $name) {
      destination {
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
      destinationType {
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
              subHeader
              horizontalDivider
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

interface ResourceDestinationCardProps {
  name: string;
  destinationIndex: number;
  // disabled indicates that the card is not active and should be greyed out
  disabled?: boolean;
}

const ResourceDestinationCardComponent: React.FC<
  ResourceDestinationCardProps
> = ({ name, destinationIndex, disabled }) => {
  const { configuration, refetchConfiguration, readOnlyGraph } =
    usePipelineGraph();
  const { enqueueSnackbar } = useSnackbar();
  const [editing, setEditing] = useState(false);
  const [confirmDeleteOpen, setDeleteOpen] = useState(false);
  const role = useRole();

  // Use the version name of the destination specified in the configuration
  const versionedName = useMemo(() => {
    const destination = configuration?.spec?.destinations?.find((d) => {
      return d.name && trimVersion(d.name) === name;
    });

    return destination?.name;
  }, [configuration?.spec?.destinations, name]);

  const { data, refetch: refetchDestination } = useGetDestinationWithTypeQuery({
    variables: { name: versionedName || name },
    fetchPolicy: "cache-and-network",
  });

  async function onSave(formValues: FormValues) {
    const updatedDestination = new BPDestination(
      data!.destinationWithType!.destination!
    );

    updatedDestination.setParamsFromMap(formValues);
    // TODO(cpheps): Validate this is ok in the long run.
    // Commenting this out now as we don't need to update the processor on save.
    // The formValues.processors is always empty and wipes out the destination processors
    // const updatedProcessors = formValues.processors;

    // // Assign processors to the configuration if we have an
    // // index for the destination, implying that we are editing this
    // // destination on a particular config and that processors are enabled.
    // if (destinationIndex != null) {
    //   const updatedConfig = new BPConfiguration(configuration);
    //   updatedConfig.replaceDestination(
    //     {
    //       name: updatedDestination.name(),
    //       processors: updatedProcessors,
    //       disabled: updatedDestination.spec.disabled,
    //     },
    //     destinationIndex
    //   );

    //   try {
    //     const update = await updatedConfig.apply();
    //     if (update.status === UpdateStatus.INVALID) {
    //       throw new Error(
    //         `failed to apply configuration, got status ${update.status}`
    //       );
    //     }
    //   } catch (err) {
    //     console.error(err);
    //     enqueueSnackbar("Failed to update configuration.", {
    //       variant: "error",
    //     });
    //   }
    // }

    try {
      const update = await updatedDestination.apply();
      if (update.status === UpdateStatus.INVALID) {
        console.error("Invalid Update: ", update);
        throw new Error(
          `failed to apply destination, got status ${update.status}`
        );
      }

      enqueueSnackbar("Saved Destination! 🎉", {
        variant: "success",
      });
      setEditing(false);
      refetchConfiguration();
      refetchDestination();
    } catch (err) {
      console.error(err);
      enqueueSnackbar("Failed to update destination.", { variant: "error" });
    }
  }

  const onDelete: onDeleteFunc | undefined = useMemo(() => {
    if (destinationIndex == null) {
      return undefined;
    }
    return async function onDelete() {
      const updatedConfig = new BPConfiguration(configuration);
      updatedConfig.removeDestination(destinationIndex);

      try {
        const update = await updatedConfig.apply();
        if (update.status === UpdateStatus.INVALID) {
          console.error("Invalid Update: ", update);
          throw new Error(
            `failed to remove destination from configuration, configuration invalid`
          );
        }

        setEditing(false);
        setDeleteOpen(false);
        refetchConfiguration();
        refetchDestination();
      } catch (err) {
        console.error(err);
        enqueueSnackbar("Failed to remove destination.", {
          variant: "error",
        });
      }
    };
  }, [
    configuration,
    destinationIndex,
    enqueueSnackbar,
    refetchConfiguration,
    refetchDestination,
  ]);

  /**
   * Toggle `disabled` on the destination spec, replace it in the configuration, and save
   */
  async function onTogglePause() {
    if (data?.destinationWithType?.destination == null) {
      enqueueSnackbar("Oops! Something went wrong.", { variant: "error" });
      console.error(
        "could not toggle destination disabled, no destination returned in data"
      );
      return;
    }

    const updatedDestination = new BPDestination(
      data.destinationWithType.destination
    );
    updatedDestination.toggleDisabled();

    const action = updatedDestination.spec.disabled ? "pause" : "resume";

    try {
      const { status, reason } = await updatedDestination.apply();
      if (status === UpdateStatus.INVALID) {
        throw new Error(
          `failed to update configuration, configuration invalid, ${reason}`
        );
      }

      enqueueSnackbar(`Destination ${action}d! 🎉`, {
        variant: "success",
      });

      setEditing(false);
      refetchConfiguration();
      refetchDestination();
    } catch (err) {
      enqueueSnackbar(`Failed to ${action} destination.`, {
        variant: "error",
      });
      console.error(err);
    }
  }
  // Loading
  if (data === undefined) {
    return null;
  }

  if (data.destinationWithType.destination == null) {
    enqueueSnackbar(`Could not retrieve destination ${name}.`, {
      variant: "error",
    });
    return null;
  }

  if (data.destinationWithType.destinationType == null) {
    enqueueSnackbar(
      `Could not retrieve destination type for destination ${name}.`,
      { variant: "error" }
    );
    return null;
  }

  return (
    <div
      className={classes([
        disabled ? styles.disabled : undefined,
        data.destinationWithType.destination?.spec.disabled
          ? styles.paused
          : undefined,
      ])}
    >
      <ResourceCard
        name={name}
        icon={data.destinationWithType?.destinationType?.metadata.icon}
        paused={data.destinationWithType.destination?.spec.disabled}
        disabled={disabled}
        onClick={() => setEditing(true)}
      />
      <EditResourceDialog
        kind="destination"
        resourceTypeDisplayName={name}
        description={
          data.destinationWithType.destinationType.metadata.description ?? ""
        }
        fullWidth
        maxWidth="sm"
        parameters={data.destinationWithType.destination.spec.parameters ?? []}
        parameterDefinitions={
          data.destinationWithType.destinationType.spec.parameters
        }
        open={editing}
        onClose={() => setEditing(false)}
        onCancel={() => setEditing(false)}
        onDelete={onDelete && (() => setDeleteOpen(true))}
        onSave={onSave}
        paused={data.destinationWithType.destination?.spec.disabled ?? false}
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
          <Typography>
            Are you sure you want to remove this destination?
          </Typography>
        </ConfirmDeleteResourceDialog>
      )}
    </div>
  );
};

export const ResourceDestinationCard = memo(ResourceDestinationCardComponent);
