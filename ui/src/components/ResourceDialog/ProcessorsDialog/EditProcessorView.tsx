import { useSnackbar } from "notistack";
import { useState } from "react";
import { FormValues, ResourceConfigForm } from "../../ResourceConfigForm";
import {
  GetProcessorTypeQuery,
  Parameter,
  ResourceConfiguration,
  useGetProcessorTypeQuery,
  useGetProcessorWithTypeQuery,
} from "../../../graphql/generated";
import { BPResourceConfiguration } from "../../../utils/classes";
import { ApolloError, gql } from "@apollo/client";
import { CircularProgress, Stack } from "@mui/material";
import { BPProcessor } from "../../../utils/classes/processor";
import { trimVersion } from "../../../utils/version-helpers";

gql`
  query getProcessorWithType($name: String!) {
    processorWithType(name: $name) {
      processor {
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
      processorType {
        metadata {
          id
          name
          version
          description
          displayName
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

interface EditProcessorViewProps {
  processors: ResourceConfiguration[];
  applyQueue: BPProcessor[];
  editingIndex: number;
  readOnly?: boolean;
  onEditInlineProcessorSave: (values: FormValues) => void;
  onEditResourceProcessorSave: (processor: BPProcessor) => void;
  onBack: () => void;
  onRemove: (removeIndex: number) => void;
}

export const EditProcessorView: React.FC<EditProcessorViewProps> = ({
  processors,
  editingIndex,
  readOnly,
  applyQueue,
  onEditInlineProcessorSave,
  onEditResourceProcessorSave,
  onBack,
  onRemove,
}) => {
  const resourceConfig = new BPResourceConfiguration(processors[editingIndex]);

  const { enqueueSnackbar } = useSnackbar();

  function onError(error: ApolloError) {
    console.error(error.message);
    enqueueSnackbar("Oops! Something went wrong.");
    onBack();
  }

  const [parameters, setParameters] = useState<Parameter[]>();
  const [processor, setProcessor] = useState<BPProcessor>();
  const [processorType, setProcessorType] =
    useState<GetProcessorTypeQuery["processorType"]>();

  useGetProcessorTypeQuery({
    variables: { type: resourceConfig.type! },
    skip: !resourceConfig.isInline(),
    onCompleted(data) {
      setProcessorType(data.processorType);
      setParameters(resourceConfig.parameters ?? []);
    },
    onError,
    fetchPolicy: "network-only",
  });

  useGetProcessorWithTypeQuery({
    variables: { name: resourceConfig.name! },
    skip: resourceConfig.isInline(),
    onCompleted(data) {
      setProcessorType(data.processorWithType.processorType);

      // Use an existing processor editor if it exists
      const existingEdit = applyQueue.find(
        (p) => p.name() === trimVersion(resourceConfig.name ?? "")
      );
      if (existingEdit) {
        setProcessor(existingEdit);
        setParameters(existingEdit.spec.parameters ?? []);
        return;
      }

      setProcessor(new BPProcessor(data.processorWithType.processor!));
      setParameters(data.processorWithType.processor?.spec?.parameters ?? []);
    },
    onError,
    fetchPolicy: "network-only",
  });

  function handleFormSave(values: FormValues) {
    if (resourceConfig.isInline()) {
      onEditInlineProcessorSave(values);
      return;
    }

    if (processor == null) {
      console.error(
        `Cannot save resource processor, no processor found with name: ${resourceConfig.name}.`
      );
      enqueueSnackbar("Oops! Something went wrong.", { variant: "error" });
      return;
    }

    const newProcessor = new BPProcessor(processor);
    newProcessor.setParamsFromMap(values);
    onEditResourceProcessorSave(newProcessor);
  }

  if (parameters == null || processorType == null) {
    return (
      <Stack
        width="100%"
        height="100%"
        justifyContent="center"
        alignItems="center"
      >
        <CircularProgress />
      </Stack>
    );
  }

  return (
    <>
      <ResourceConfigForm
        resourceTypeDisplayName={
          processorType.metadata.displayName ?? processorType.metadata.name
        }
        displayName={resourceConfig.displayName ?? resourceConfig.name ?? ""}
        description={processorType.metadata.description ?? ""}
        kind={"processor"}
        parameterDefinitions={processorType.spec.parameters}
        includeDisplayNameField={resourceConfig.isInline()}
        parameters={parameters}
        onSave={handleFormSave}
        saveButtonLabel="Done"
        onBack={onBack}
        onDelete={() => onRemove(editingIndex)}
        readOnly={readOnly}
        embedded={true}
      />
    </>
  );
};
