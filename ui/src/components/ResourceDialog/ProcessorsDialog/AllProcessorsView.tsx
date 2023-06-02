import { Button } from "@mui/material";
import { ResourceConfiguration } from "../../../graphql/generated";
import { ActionsSection } from '../../DialogComponents';
import { InlineProcessorContainer } from "./InlineProcessorContainer";
import { ViewHeading } from './ViewHeading';

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
    <>
      <ViewHeading heading="Processors" />
      <InlineProcessorContainer
        processors={processors}
        onAddProcessor={onAddProcessor}
        onEditProcessor={onEditProcessor}
        onProcessorsChange={onProcessorsChange}
        disableEdit={readOnly}
      />
      {!readOnly && (
        <ActionsSection>
          <Button variant="contained" onClick={onSave}>
            Save
          </Button>
        </ActionsSection>
      )}
    </>
  );
};
