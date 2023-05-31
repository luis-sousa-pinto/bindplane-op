import { gql } from "@apollo/client";
import {
  Card,
  CardActionArea,
  CardContent,
  Stack,
  Typography,
} from "@mui/material";
import { useSnackbar } from "notistack";
import { memo, useMemo, useState } from "react";
import { ConfirmDeleteResourceDialog } from "../ConfirmDeleteResourceDialog";
import { EditResourceDialog } from "../ResourceDialog/EditResourceDialog";
import { useGetDestinationWithTypeQuery } from "../../graphql/generated";
import { UpdateStatus } from "../../types/resources";
import { BPConfiguration, BPDestination } from "../../utils/classes";
import { FormValues } from "../ResourceConfigForm";
import { classes } from "../../utils/styles";
import { NoMaxWidthTooltip } from "../Custom/NoMaxWidthTooltip";
import { truncateLabel } from "../../utils/graph/utils";
import { usePipelineGraph } from "../PipelineGraph/PipelineGraphContext";
import { trimVersion } from "../../utils/version-helpers";

import styles from "./cards.module.scss";

type onDeleteFunc = () => Promise<void>;

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


const ResourceDestinationCardComponent: React.FC<ResourceDestinationCardProps> =
  ({ name, destinationIndex, disabled }) => {
  const { configuration, refetchConfiguration,  readOnlyGraph } = usePipelineGraph();
  const { enqueueSnackbar } = useSnackbar();
  const [editing, setEditing] = useState(false);
  const [confirmDeleteOpen, setDeleteOpen] = useState(false);

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


  function closeEditDialog() {
    setEditing(false);
  }

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
        console.error("Update: ", update);
        throw new Error(
          `failed to apply destination, got status ${update.status}`
        );
      }

      enqueueSnackbar("Successfully saved destination.", {
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
          console.error("Update: ", update);
          throw new Error(
            `failed to remove destination from configuration, configuration invalid`
          );
        }

        closeEditDialog();
        closeDeleteDialog();
        refetchConfiguration();
        refetchDestination();
      } catch (err) {
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
    const updatedConfig = new BPConfiguration(configuration);
    const updatedDestination = new BPDestination(
      data!.destinationWithType!.destination!
    );
    updatedDestination.toggleDisabled();

    const action = updatedDestination.spec.disabled ? "pause" : "resume";
    if (destinationIndex != null) {
      updatedConfig.replaceDestination(
        {
          name: updatedDestination.name(),
          processors:
            configuration?.spec?.destinations![destinationIndex].processors ??
            [],
          parameters: updatedDestination.spec.parameters,
          type: updatedDestination.spec.type,
          disabled: updatedDestination.spec.disabled,
        },
        destinationIndex
      );

      try {
        const update = await updatedConfig.apply();
        if (update.status === UpdateStatus.INVALID) {
          throw new Error(
            `failed to ${action} destination, got status ${update.status}`
          );
        }
      } catch (err) {
        console.error(err);
        enqueueSnackbar("Failed to update configuration.", {
          variant: "error",
        });
      }
    }

    try {
      const { status, reason } = await updatedDestination.apply();
      if (status === UpdateStatus.INVALID) {
        throw new Error(
          `failed to update configuration, configuration invalid, ${reason}`
        );
      }

      enqueueSnackbar(`Successfully ${action}d destination.`, {
        variant: "success",
      });
      closeEditDialog();
      refetchConfiguration();
      refetchDestination();
    } catch (err) {
      enqueueSnackbar(`Failed to ${action} destination.`, {
        variant: "error",
      });
      console.error(err);
    }
  }

  function closeDeleteDialog() {
    setDeleteOpen(false);
  }

  function openDeleteDialog() {
    setDeleteOpen(true);
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
      <Card
        className={classes([
          styles["resource-card"],
          disabled ? styles.disabled : undefined,
          data.destinationWithType.destination?.spec.disabled
            ? styles.paused
            : undefined,
        ])}
        onClick={() => setEditing(true)}
      >
        <CardActionArea className={styles.action}>
          <NoMaxWidthTooltip title={name.length > 20 ? name : ""}>
            <CardContent>
              <Stack alignItems="center">
                <span
                  className={styles.icon}
                  style={{
                    backgroundImage: `url(${data?.destinationWithType?.destinationType?.metadata.icon})`,
                  }}
                />
                <Typography
                  component="div"
                  fontWeight={600}
                  gutterBottom
                  fontSize={name.length > 15 ? 11 : 16}
                >
                  {truncateLabel(name, 20)}
                </Typography>
                {data.destinationWithType.destination?.spec.disabled && (
                  <Typography
                    component="div"
                    fontWeight={400}
                    fontSize={14}
                    variant="overline"
                  >
                    Paused
                  </Typography>
                )}
              </Stack>
            </CardContent>
          </NoMaxWidthTooltip>
        </CardActionArea>
      </Card>

      <EditResourceDialog
        kind="destination"
        displayName={name}
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
        onClose={closeEditDialog}
        onCancel={closeEditDialog}
        onDelete={onDelete && openDeleteDialog}
        onSave={onSave}
        paused={data.destinationWithType.destination?.spec.disabled ?? false}
        onTogglePause={onTogglePause}
        readOnly={readOnlyGraph}
      />

      {onDelete && (
        <ConfirmDeleteResourceDialog
          open={confirmDeleteOpen}
          onClose={closeDeleteDialog}
          onCancel={closeDeleteDialog}
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
