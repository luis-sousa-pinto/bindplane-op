import { Button, Stack } from "@mui/material";
import { DialogResource, ResourceType } from ".";
import { ResourceTypeButton } from "../ResourceTypeButton";
import { useResourceDialog } from "./ResourceDialogContext";
import {
  TitleSection,
  ContentSection,
  ActionsSection,
} from "../DialogComponents";
import { filterResourcesByType } from "./utils";

interface ChooseViewProps {
  resources: DialogResource[];
  selected: ResourceType;
  kind: "source" | "destination";
  clearResource: () => void;
  handleSaveExisting: (r: DialogResource) => void;
  setCreateNew: (b: boolean) => void;
}

export const ChooseView: React.FC<ChooseViewProps> = ({
  resources,
  selected,
  handleSaveExisting,
  setCreateNew,
  clearResource,
}) => {
  const { onClose } = useResourceDialog();

  const matchingResources = filterResourcesByType(resources, selected);

  return (
    <>
      <TitleSection title={"Choose Existing or Create New"} onClose={onClose} />

      <ContentSection>
        <Stack spacing={1}>
          {matchingResources?.map((resource) => {
            return (
              <ResourceTypeButton
                key={resource.metadata.name}
                icon={selected?.metadata.icon!}
                displayName={resource.metadata.name}
                onSelect={() => handleSaveExisting(resource)}
              />
            );
          })}
          <Button
            variant="contained"
            color="primary"
            onClick={() => setCreateNew(true)}
          >
            Create New
          </Button>
        </Stack>
      </ContentSection>

      <ActionsSection>
        <Button
          variant="outlined"
          color="secondary"
          onClick={() => clearResource()}
        >
          Back
        </Button>
      </ActionsSection>
    </>
  );
};
