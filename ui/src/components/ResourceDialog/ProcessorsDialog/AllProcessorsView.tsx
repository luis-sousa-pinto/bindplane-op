import { Button, Stack } from "@mui/material";
import { ResourceConfiguration, Role } from "../../../graphql/generated";
import { ActionsSection } from "../../DialogComponents";
import { ProcessorsContainer } from "./ProcessorContainer";
import { ViewHeading } from "./ViewHeading";
import { RBACWrapper } from "../../RBACWrapper/RBACWrapper";

import mixins from "../../../styles/mixins.module.scss";

interface AllProcessorsProps {
  processors: ResourceConfiguration[];
  readOnly: boolean;
  onAddProcessor: () => void;
  onEditProcessor: (index: number) => void;
  onSave: () => void;
  onProcessorsChange: (pt: ResourceConfiguration[]) => void;
}

/**
 * AllProcessorsView shows the initial view of the processors dialog, which is a list of all processors,
 * with the ability to add a new processor, reorder processors, and select a processor to edit or delete.
 */
export const AllProcessorsView: React.FC<AllProcessorsProps> = ({
  processors,
  readOnly,
  onAddProcessor,
  onEditProcessor,
  onSave,
  onProcessorsChange,
}) => {
  return (
    <Stack className={mixins["flex-grow"]}>
      <ViewHeading heading="Processors" />
      <ProcessorsContainer
        processors={processors}
        onAddProcessor={onAddProcessor}
        onEditProcessor={onEditProcessor}
        onProcessorsChange={onProcessorsChange}
        disableEdit={readOnly}
      />
      {!readOnly && (
        <ActionsSection>
          <RBACWrapper requiredRole={Role.User}>
            <Button variant="contained" onClick={onSave}>
              Save
            </Button>
          </RBACWrapper>
        </ActionsSection>
      )}
    </Stack>
  );
};
