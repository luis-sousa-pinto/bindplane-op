import { gql } from "@apollo/client";
import { Dialog, DialogProps, Grid, Stack } from "@mui/material";
import { isEqual } from "lodash";
import { useSnackbar } from "notistack";
import { useEffect, useMemo, useState } from "react";
import {
  GetProcessorTypesQuery,
  PipelineType,
  ResourceConfiguration,
  ResourceTypeKind,
  Role,
  useProcessorDialogDestinationTypeLazyQuery,
  useProcessorDialogSourceTypeLazyQuery,
  useUpdateProcessorsMutation,
} from "../../../graphql/generated";
import { BPResourceConfiguration } from "../../../utils/classes";
import { trimVersion } from "../../../utils/version-helpers";
import { usePipelineGraph } from "../../PipelineGraph/PipelineGraphContext";
import {
  CreateProcessorConfigureView,
  CreateProcessorSelectView,
  EditProcessorView,
  FormValues,
} from "../../ResourceConfigForm";
import { SnapshotConsole } from "../../SnapShotConsole/SnapShotConsole";
import { ResourceDialogContextProvider } from "../ResourceDialogContext";
import { AllProcessorsView } from "./AllProcessorsView";
import {
  SnapshotContextProvider,
  useSnapshot,
} from "../../SnapShotConsole/SnapshotContext";
import { hasPermission } from "../../../utils/has-permission";
import { useRole } from "../../../hooks/useRole";
import { TitleSection } from "../../DialogComponents";
import { BPProcessor } from "../../../utils/classes/processor";
import { applyResources } from "../../../utils/rest/apply-resources";
import { UpdateStatus } from "../../../types/resources";

import styles from "./processor-dialog.module.scss";
import mixins from "../../../styles/mixins.module.scss";

interface ProcessorDialogProps extends DialogProps {
  processors: ResourceConfiguration[];
  readOnly?: boolean;
}

gql`
  query processorDialogSourceType($name: String!) {
    sourceType(name: $name) {
      metadata {
        name
        id
        version
        displayName
        description
      }
      spec {
        telemetryTypes
      }
    }
  }

  query processorDialogDestinationType($name: String!) {
    destinationWithType(name: $name) {
      destinationType {
        metadata {
          id
          name
          version
          displayName
          description
        }
        spec {
          telemetryTypes
        }
      }
    }
  }

  mutation updateProcessors($input: UpdateProcessorsInput!) {
    updateProcessors(input: $input)
  }
`;

enum Page {
  MAIN,
  CREATE_PROCESSOR_SELECT,
  CREATE_PROCESSOR_CONFIGURE,
  EDIT_PROCESSOR,
}

export type ProcessorType = GetProcessorTypesQuery["processorTypes"][0];

export const ProcessorDialog: React.FC = () => {
  const role = useRole();
  const {
    editProcessorsInfo,
    configuration,
    editProcessorsOpen,
    closeProcessorDialog,
    readOnlyGraph,
  } = usePipelineGraph();

  if (editProcessorsInfo === null) {
    return null;
  }

  let processors: ResourceConfiguration[] = [];
  switch (editProcessorsInfo?.resourceType) {
    case "source":
      processors =
        configuration?.spec?.sources?.[editProcessorsInfo.index].processors ??
        [];
      break;
    case "destination":
      processors =
        configuration?.spec?.destinations?.[editProcessorsInfo.index]
          .processors ?? [];
      break;
    default:
      processors = [];
  }

  return (
    <ProcessorDialogComponent
      open={editProcessorsOpen}
      onClose={closeProcessorDialog}
      processors={processors}
      readOnly={readOnlyGraph || !hasPermission(Role.User, role)}
    />
  );
};

export const ProcessorDialogComponent: React.FC<ProcessorDialogProps> = ({
  processors: processorsProp,
  readOnly,
  ...dialogProps
}) => {
  const {
    editProcessorsInfo,
    configuration,
    closeProcessorDialog,
    refetchConfiguration,
    selectedTelemetryType,
  } = usePipelineGraph();

  const [processors, setProcessors] = useState(processorsProp);
  const [view, setView] = useState<Page>(Page.MAIN);
  const [newProcessorType, setNewProcessorType] =
    useState<ProcessorType | null>(null);
  const [editingProcessorIndex, setEditingProcessorIndex] =
    useState<number>(-1);
  const [applyQueue, setApplyQueue] = useState<BPProcessor[]>([]);

  const { enqueueSnackbar } = useSnackbar();

  // Get the type of the source or destination to which we're adding processors
  const resourceTypeName = useMemo(() => {
    try {
      switch (editProcessorsInfo?.resourceType) {
        case "source":
          return configuration?.spec?.sources?.[editProcessorsInfo.index].type;
        case "destination":
          return configuration?.spec?.destinations?.[editProcessorsInfo.index]
            .type;
        default:
          return null;
      }
    } catch (err) {
      return null;
    }
  }, [
    configuration?.spec?.destinations,
    configuration?.spec?.sources,
    editProcessorsInfo?.index,
    editProcessorsInfo?.resourceType,
  ]);

  /* ------------------------ GQL Mutations and Queries ----------------------- */
  const [updateProcessors] = useUpdateProcessorsMutation({});

  const [fetchSourceType, { data: sourceTypeData }] =
    useProcessorDialogSourceTypeLazyQuery({
      variables: { name: resourceTypeName ?? "" },
    });
  const [fetchDestinationType, { data: destinationTypeData }] =
    useProcessorDialogDestinationTypeLazyQuery();

  /* ----------------------------- Event Handlers ----------------------------- */
  useEffect(() => {
    // resetState returns the processor to the first page after close.
    function resetState() {
      setView(Page.MAIN);
      setNewProcessorType(null);
      setEditingProcessorIndex(-1);
    }

    let timeout: ReturnType<typeof setTimeout>;
    // Resets the state on close
    if (dialogProps.open === false) {
      timeout = setTimeout(resetState, 500);
    }

    return () => {
      clearTimeout(timeout);
    };
  }, [dialogProps.open]);

  useEffect(() => {
    function fetchData() {
      if (editProcessorsInfo!.resourceType === "source") {
        fetchSourceType({ variables: { name: resourceTypeName ?? "" } });
      } else {
        const destName =
          configuration?.spec?.destinations?.[editProcessorsInfo!.index].name;
        fetchDestinationType({ variables: { name: destName ?? "" } });
      }
    }

    if (editProcessorsInfo == null) {
      return;
    }

    fetchData();
  }, [
    configuration?.spec?.destinations,
    editProcessorsInfo,
    fetchDestinationType,
    fetchSourceType,
    resourceTypeName,
  ]);

  /* -------------------------------- Functions ------------------------------- */

  // handleSelectNewProcessorType is called when a user selects a processor type
  // in the CreateProcessorSelect view.
  function handleSelectNewProcessorType(type: ProcessorType) {
    setNewProcessorType(type);
    setView(Page.CREATE_PROCESSOR_CONFIGURE);
  }

  function handleReturnToAll() {
    setView(Page.MAIN);
    setNewProcessorType(null);
  }

  // handleClose is called when a user clicks off the dialog or the "X" button
  function handleClose() {
    if (!isEqual(processors, processorsProp)) {
      const ok = window.confirm("Discard changes?");
      if (!ok) {
        return;
      }
      // reset form values if chooses to discard.
      setProcessors(processorsProp);
    }

    closeProcessorDialog();
  }

  // handleAddProcessor adds a new processor to the list of processors
  async function handleAddProcessor(formValues: FormValues) {
    const processorConfig = new BPResourceConfiguration();
    processorConfig.setParamsFromMap(formValues);
    processorConfig.type = newProcessorType!.metadata.name;

    const newProcessors = [...processors, processorConfig];
    setProcessors(newProcessors);
    handleReturnToAll();
  }

  // handleSaveExistingInlineProcessor saves changes to an existing processor in the list
  function handleSaveExistingInlineProcessor(formValues: FormValues) {
    const processorConfig = new BPResourceConfiguration();
    processorConfig.setParamsFromMap(formValues);
    processorConfig.type = processors[editingProcessorIndex].type;

    const newProcessors = [...processors];
    newProcessors[editingProcessorIndex] = processorConfig;
    setProcessors(newProcessors);

    handleReturnToAll();
  }

  // handleSaveExistingResourceProcessor adds a processor to the apply queue
  function handleSaveExistingResourceProcessor(processor: BPProcessor) {
    const foundIndex = applyQueue.findIndex(
      (p) => p.name() === processor.name()
    );
    if (foundIndex !== -1) {
      const newApplyQueue = [...applyQueue];
      newApplyQueue[foundIndex] = processor;
      setApplyQueue(newApplyQueue);
    } else {
      setApplyQueue([...applyQueue, processor]);
    }

    handleReturnToAll();
  }

  // handleRemoveProcessor removes a processor from the list of processors
  async function handleRemoveProcessor(index: number) {
    const newProcessors = [...processors];
    newProcessors.splice(index, 1);
    setProcessors(newProcessors);

    handleReturnToAll();
  }

  // handleEditProcessorClick sets the editing index and switches to the edit page
  function handleEditProcessorClick(index: number) {
    setEditingProcessorIndex(index);
    setView(Page.EDIT_PROCESSOR);
  }

  // handleSave saves the processors to the backend and closes the dialog.
  async function handleSave() {
    const inlineChange = !isEqual(processorsProp, processors);
    const resourceChange = applyQueue.length > 0;
    var closeDialog = true;

    if (!inlineChange && !resourceChange) {
      closeProcessorDialog();
    }

    if (resourceChange) {
      const { updates } = await applyResources(applyQueue);
      if (updates.some((u) => u.status === UpdateStatus.INVALID)) {
        enqueueSnackbar("Failed to save processors", {
          variant: "error",
          key: "save-processors-error",
        });
        closeDialog = false;
      }
    }

    if (inlineChange) {
      await updateProcessors({
        variables: {
          input: {
            configuration: configuration?.metadata?.name!,
            resourceType:
              editProcessorsInfo?.resourceType === "source"
                ? ResourceTypeKind.Source
                : ResourceTypeKind.Destination,
            resourceIndex: editProcessorsInfo?.index!,
            processors: processors,
          },
        },
        onError(error) {
          closeDialog = false;
          console.error(error);
          enqueueSnackbar("Failed to save processors", {
            variant: "error",
            key: "save-processors-error",
          });
        },
      });
    }

    if (closeDialog) {
      closeProcessorDialog();
      refetchConfiguration();
      enqueueSnackbar("Saved Processors! 🎉", { variant: "success" });
    }
  }

  // displayName for sources is the sourceType name, for destinations it's the destination name
  const { resourceName, displayName } = useMemo(() => {
    if (editProcessorsInfo == null) {
      return {
        resourceName: "",
        displayName: "",
      };
    }
    if (editProcessorsInfo.resourceType === "source") {
      return {
        resourceName: `source${editProcessorsInfo.index}`,
        displayName: sourceTypeData?.sourceType?.metadata.displayName,
      };
    }
    const name =
      configuration?.spec?.destinations?.[editProcessorsInfo.index].name;
    if (name) {
      const trimName = trimVersion(name);
      return {
        resourceName: `${trimName}-${editProcessorsInfo.index}`,
        displayName: trimName,
      };
    }
    return {
      resourceName: `dest${editProcessorsInfo.index}`,
      displayName: `dest${editProcessorsInfo.index}`,
    };
  }, [
    configuration?.spec?.destinations,
    editProcessorsInfo,
    sourceTypeData?.sourceType?.metadata.displayName,
  ]);

  let current: JSX.Element;
  switch (view) {
    case Page.MAIN:
      current = (
        <>
          <AllProcessorsView
            processors={processors}
            onAddProcessor={() => setView(Page.CREATE_PROCESSOR_SELECT)}
            onEditProcessor={handleEditProcessorClick}
            onSave={handleSave}
            onProcessorsChange={setProcessors}
            readOnly={Boolean(readOnly)}
          />
        </>
      );
      break;
    case Page.CREATE_PROCESSOR_SELECT:
      current = (
        <CreateProcessorSelectView
          displayName={displayName ?? "unknown"}
          telemetryTypes={
            editProcessorsInfo?.resourceType === "source"
              ? sourceTypeData?.sourceType?.spec?.telemetryTypes
              : destinationTypeData?.destinationWithType.destinationType?.spec
                  ?.telemetryTypes
          }
          onBack={() => setView(Page.MAIN)}
          onSelect={handleSelectNewProcessorType}
        />
      );
      break;
    case Page.CREATE_PROCESSOR_CONFIGURE:
      current = (
        <CreateProcessorConfigureView
          processorType={newProcessorType!}
          onBack={handleReturnToAll}
          onSave={handleAddProcessor}
          onClose={closeProcessorDialog}
        />
      );
      break;
    case Page.EDIT_PROCESSOR:
      current = (
        <EditProcessorView
          processors={processors}
          editingIndex={editingProcessorIndex}
          applyQueue={applyQueue}
          onEditInlineProcessorSave={handleSaveExistingInlineProcessor}
          onEditResourceProcessorSave={handleSaveExistingResourceProcessor}
          onBack={handleReturnToAll}
          onRemove={handleRemoveProcessor}
          readOnly={readOnly}
        />
      );
  }

  const parentDisplayName = displayName ?? "unknown";
  const title = useMemo(() => {
    const kind =
      editProcessorsInfo?.resourceType === "source" ? "Source" : "Destination";
    return `${kind} ${parentDisplayName}: Processors`;
  }, [editProcessorsInfo?.resourceType, parentDisplayName]);

  const description =
    "Processors are run on data after it's received and prior to being sent to a destination. They will be executed in the order they appear below.";

  const snapshotPosition =
    editProcessorsInfo?.resourceType === "source" ? "s0" : "d0";

  return (
    <ResourceDialogContextProvider onClose={handleClose} purpose={"edit"}>
      <SnapshotContextProvider
        pipelineType={convertTelemetryType(selectedTelemetryType)}
        processors={processors}
        position={snapshotPosition}
        resourceName={resourceName}
        showAgentSelector={true}
      >
        <Dialog
          {...dialogProps}
          maxWidth={"xl"}
          fullWidth
          onClose={handleClose}
          classes={{
            root: styles.dialog,
            paper: styles.paper,
          }}
        >
          <Stack height="calc(100vh - 100px)" minHeight="800px">
            <TitleSection
              title={title}
              description={description}
              onClose={handleClose}
            />
            <Stack
              direction="row"
              width="100%"
              height="calc(100vh - 175px)"
              minHeight={"700px"}
              spacing={2}
              padding={2}
            >
              <div
                className={(mixins["flex-grow"], styles["snapshot-container"])}
              >
                <SnapshotSection />
              </div>

              <div className={styles["form-container"]}>{current}</div>
            </Stack>
          </Stack>
        </Dialog>
      </SnapshotContextProvider>
    </ResourceDialogContextProvider>
  );
};

function convertTelemetryType(telemetryType: string): PipelineType {
  switch (telemetryType) {
    case PipelineType.Logs:
      return PipelineType.Logs;
    case PipelineType.Metrics:
      return PipelineType.Metrics;
    case PipelineType.Traces:
      return PipelineType.Traces;
    default:
      return PipelineType.Logs;
  }
}

export const SnapshotSection: React.FC = () => {
  const { logs, metrics, traces, pipelineType } = useSnapshot();
  return (
    <SnapshotConsole
      logs={logs}
      metrics={metrics}
      traces={traces}
      footer={`Showing recent ${pipelineType}`}
    />
  );
};

export const ProcessorsBody: React.FC<{}> = ({ children }) => {
  const { logs, metrics, traces, pipelineType } = useSnapshot();

  const footer = `Showing recent ${pipelineType}`;
  return (
    <>
      <Grid container spacing={2}>
        <Grid item xs={7}>
          <SnapshotConsole
            logs={logs}
            metrics={metrics}
            traces={traces}
            footer={footer}
          />
        </Grid>
        <Grid item xs={5}>
          {children}
        </Grid>
      </Grid>
    </>
  );
};
