import { useMemo, useState } from "react";
import { DialogResource, ResourceType } from ".";
import { metadataSatisfiesSubstring } from "../../utils/metadata-satisfies-substring";
import {
  ResourceTypeButton,
  ResourceTypeButtonContainer,
} from "../ResourceTypeButton";
import { useResourceDialog } from "./ResourceDialogContext";
import { TitleSection, ContentSection } from "../DialogComponents";
import { someResourceOfType } from "./utils";

interface SelectViewProps {
  resourceTypes: ResourceType[];
  resources: DialogResource[];
  setSelected: (t: ResourceType) => void;
  setCreateNew: (b: boolean) => void;
  kind: "source" | "destination";
  platform: string;
}

export const SelectView: React.FC<SelectViewProps> = ({
  platform,
  resourceTypes,
  resources,
  setSelected,
  setCreateNew,
  kind,
}) => {
  const [resourceSearchValue, setResourceSearch] = useState("");
  const { onClose } = useResourceDialog();

  const sortedResourceTypes = useMemo(() => {
    const copy = resourceTypes.slice();
    return copy.sort((a, b) =>
      a.metadata
        .displayName!.toLowerCase()
        .localeCompare(b.metadata.displayName!.toLowerCase())
    );
  }, [resourceTypes]);

  return (
    <>
      <TitleSection
        title={kind === "destination" ? "Add Destination" : "Add Source"}
        onClose={onClose}
      />

      <ContentSection>
        <ResourceTypeButtonContainer
          onSearchChange={(v: string) => setResourceSearch(v)}
        >
          {sortedResourceTypes
            .filter((rt) => filterByPlatform(platform, kind, rt))
            // Filter resource types by the resourceSearchValue
            .filter((rt) => metadataSatisfiesSubstring(rt, resourceSearchValue))
            // map the results to resource buttons
            .map((resourceType) => {
              const matchingResourcesExist = someResourceOfType(
                resources,
                resourceType
              );

              // Either we send the directly to the form if there are no existing resources
              // of that type, or we send them to the Choose View by just setting the selected.
              function onSelect() {
                setSelected(resourceType);
                if (!matchingResourcesExist) {
                  setCreateNew(true);
                }
              }
              return (
                <ResourceTypeButton
                  key={resourceType.metadata.name}
                  icon={resourceType.metadata.icon!}
                  displayName={resourceType.metadata.displayName!}
                  onSelect={onSelect}
                  telemetryTypes={resourceType.spec.telemetryTypes}
                />
              );
            })}
        </ResourceTypeButtonContainer>
      </ContentSection>
    </>
  );
};

function filterByPlatform(
  platform: string,
  kind: "source" | "destination",
  resourceType: ResourceType
) {
  if (kind !== "source") {
    return true;
  }

  if (platform === "unknown") {
    return true;
  }

  return resourceType.spec.supportedPlatforms.some((p) => p === platform);
}
