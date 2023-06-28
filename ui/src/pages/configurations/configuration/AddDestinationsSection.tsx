import { memo } from "react";
import {
  DestinationType,
  SourceType,
  useDestinationsAndTypesQuery,
} from "../../../graphql/generated";
import {
  DialogResource,
  NewResourceDialog,
} from "../../../components/ResourceDialog";
import { applyResources } from "../../../utils/rest/apply-resources";
import { useSnackbar } from "notistack";
import { ShowPageConfig } from ".";
import { UpdateStatus } from "../../../types/resources";
import { BPConfiguration, BPDestination } from "../../../utils/classes";

type ResourceType = SourceType | DestinationType;

const AddDestinationsComponent: React.FC<{
  configuration: NonNullable<ShowPageConfig>;
  refetch: () => {};
  setAddDialogOpen: React.Dispatch<React.SetStateAction<boolean>>;
  addDialogOpen: boolean;
}> = ({ configuration, refetch, setAddDialogOpen, addDialogOpen }) => {
  const { data } = useDestinationsAndTypesQuery({
    fetchPolicy: "network-only",
  });
  const { enqueueSnackbar } = useSnackbar();

  async function onNewDestinationSave(
    values: { [key: string]: any },
    destinationType: ResourceType
  ) {
    if (configuration == null) {
      console.error(
        "cannot save destination, current configuration is null or undefined."
      );
      return;
    }

    const destination = new BPDestination({
      metadata: {
        name: values.name,
        id: values.name,
        version: 0,
      },
      spec: {
        parameters: [],
        type: destinationType.metadata.name,
        disabled: false,
      },
    });

    destination.setParamsFromMap(values);

    const updatedConfiguration = new BPConfiguration(configuration);
    updatedConfiguration.addDestination({
      name: destination.name(),
      disabled: destination.spec.disabled,
    });

    try {
      const { updates } = await applyResources([
        destination,
        updatedConfiguration,
      ]);

      const destinationUpdate = updates.find(
        (u) => u.resource.metadata.name === destination.name()
      );

      if (destinationUpdate == null) {
        throw new Error(
          `failed to create destination, no update returned with name ${values.name}`
        );
      }

      if (destinationUpdate.status !== UpdateStatus.CREATED) {
        throw new Error(
          `failed to create destination, got update status ${destinationUpdate.status}`
        );
      }

      const configurationUpdate = updates.find(
        (u) => u.resource.metadata.name === updatedConfiguration.name()
      );

      if (configurationUpdate == null) {
        throw new Error(
          `failed to update configuration, no update returned with name ${values.name}`
        );
      }

      if (configurationUpdate.status !== UpdateStatus.CONFIGURED) {
        throw new Error(
          `failed to update configuration, got update status ${configurationUpdate.status}`
        );
      }

      setAddDialogOpen(false);
      enqueueSnackbar(`Created destination ${destination.name()}!`, {
        variant: "success",
      });
      refetch();
    } catch (err) {
      enqueueSnackbar("Failed to create destination.", { variant: "error" });
      console.error(err);
    }
  }

  async function addExistingDestination(existingDestination: DialogResource) {
    const config = new BPConfiguration(configuration);
    config.addDestination({
      name: existingDestination.metadata.name,
      disabled: existingDestination.spec.disabled ?? false,
    });

    try {
      const update = await config.apply();
      if (update.status === UpdateStatus.INVALID) {
        console.error(update);
        throw new Error(
          `failed to update resource, got status ${update.status}`
        );
      }

      setAddDialogOpen(false);
      refetch();
    } catch (err) {
      enqueueSnackbar("Failed to add destination.", { variant: "error" });
    }
  }

  return (
    <NewResourceDialog
      platform={configuration.metadata.labels?.platform ?? "unknown"}
      kind="destination"
      resources={data?.destinations ?? []}
      resourceTypes={data?.destinationTypes ?? []}
      open={addDialogOpen}
      onSaveNew={onNewDestinationSave}
      onSaveExisting={addExistingDestination}
      onClose={() => setAddDialogOpen(false)}
    />
  );
};

export const AddDestinationsSection = memo(AddDestinationsComponent);
