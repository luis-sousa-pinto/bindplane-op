import { Button, Stack, Typography } from "@mui/material";
import { ResourceConfiguration } from "../../../graphql/generated";
import { PlusCircleIcon } from "../../Icons";
import { ProcessorLabel } from "./ProcessorLabel";
import { DndProvider } from "react-dnd";
import { HTML5Backend } from "react-dnd-html5-backend";
import { useCallback, useState } from "react";
import { ViewOnlyProcessorLabel } from "./ViewOnlyProcessorLabel";

import mixins from "../../../styles/mixins.module.scss";

interface Props {
  processors: ResourceConfiguration[];
  onAddProcessor: () => void;
  onEditProcessor: (index: number) => void;
  onProcessorsChange: (ps: ResourceConfiguration[]) => void;
  disableEdit?: boolean;
}

export const ProcessorsContainer: React.FC<Props> = ({
  processors: processorsProp,
  onProcessorsChange,
  onAddProcessor,
  onEditProcessor,
  disableEdit = false,
}) => {
  // manage state internally
  const [processors, setProcessors] = useState(processorsProp);

  function handleDrop() {
    onProcessorsChange(processors);
  }

  const moveProcessor = useCallback(
    (dragIndex: number, hoverIndex: number) => {
      if (dragIndex === hoverIndex) {
        return;
      }

      const newProcessors = [...processors];

      const dragItem = newProcessors[dragIndex];
      const hoverItem = newProcessors[hoverIndex];

      // Swap places of dragItem and hoverItem in the array
      newProcessors[dragIndex] = hoverItem;
      newProcessors[hoverIndex] = dragItem;

      setProcessors(newProcessors);
    },
    [processors, setProcessors]
  );

  return (
    <Stack className={mixins["flex-grow"]}>
      <DndProvider backend={HTML5Backend}>
        {disableEdit && processors.length === 0 && (
          <Stack
            justifyContent="center"
            alignItems="center"
            width="100%"
            marginBottom={2}
          >
            <Typography>No processors</Typography>
          </Stack>
        )}
        {processors.map((p, ix) => {
          return disableEdit ? (
            <ViewOnlyProcessorLabel
              key={`${p.name}-${ix}`}
              index={ix}
              processor={p}
              onEdit={() => onEditProcessor(ix)}
            />
          ) : (
            <ProcessorLabel
              moveProcessor={moveProcessor}
              key={`${p.name}-${ix}`}
              processor={p}
              onEdit={() => onEditProcessor(ix)}
              index={ix}
              onDrop={handleDrop}
            />
          );
        })}
        {!disableEdit && (
          <Button
            variant="text"
            startIcon={<PlusCircleIcon />}
            classes={{ root: mixins["mb-2"] }}
            onClick={onAddProcessor}
          >
            Add processor
          </Button>
        )}
      </DndProvider>
    </Stack>
  );
};
