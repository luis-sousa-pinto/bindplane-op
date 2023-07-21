import { DialogResource, ResourceType } from ".";
import { AdditionalInfo } from "../../graphql/generated";
import { FormValues, ResourceConfigForm } from "../ResourceConfigForm";

interface ConfigureViewProps {
  selected: ResourceType;
  kind: "source" | "destination";
  createNew: boolean;
  clearResource: () => void;
  handleSaveNew: (formValues: FormValues, selected: ResourceType) => void;
  resources: DialogResource[];
  additionalInfo?: AdditionalInfo | null;
}

export const ConfigureView: React.FC<ConfigureViewProps> = ({
  selected,
  kind,
  createNew,
  clearResource,
  handleSaveNew,
  resources,
  additionalInfo,
}) => {
  if (selected === null) {
    return <></>;
  }

  return (
    <ResourceConfigForm
      kind={kind}
      includeNameField={kind === "destination" && createNew}
      includeDisplayNameField={kind === "source"}
      existingResourceNames={resources?.map((r) => r.metadata.name)}
      onBack={clearResource}
      onSave={(fv) => handleSaveNew(fv, selected)}
      resourceTypeDisplayName={selected.metadata.displayName ?? ""}
      description={selected.metadata.description ?? ""}
      parameterDefinitions={selected.spec.parameters ?? []}
      additionalInfo={additionalInfo}
    />
  );
};
