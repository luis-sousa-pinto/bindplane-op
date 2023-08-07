import { gql } from "@apollo/client";
import { useRef } from "react";
import { useDrag, useDrop } from "react-dnd";
import { ResourceConfiguration } from "../../../graphql/generated";
import { ProcessorLabelCard } from "./ProcessorLabelCard";

interface Props {
  index: number;
  processor: ResourceConfiguration;
  onEdit: () => void;
  // Move processor should change the components order state
  moveProcessor: (dragIndex: number, dropIndex: number) => void;

  onDrop: () => void;

  viewOnly?: boolean;
}

gql`
  query getProcessorType($type: String!) {
    processorType(name: $type) {
      metadata {
        displayName
        name
        version
        id
        description
        deprecated
        additionalInfo {
          message
          documentation {
            text
            url
          }
        }
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
          options {
            creatable
            trackUnchecked
            gridColumns
            sectionHeader
            subHeader
            horizontalDivider
            multiline
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
          documentation {
            text
            url
          }
          validValues
        }
      }
    }
  }
`;

type Item = {
  index: number;
};

export const ProcessorLabel: React.FC<Props> = ({
  index,
  processor,
  onEdit,
  moveProcessor,
  onDrop,
}) => {
  const [, dragRef] = useDrag({
    type: "inline-processor",
    item: { index },
    collect: (monitor) => ({
      isDragging: monitor.isDragging(),
    }),
  });

  const [{ isHovered }, dropRef] = useDrop<
    Item,
    unknown,
    { isHovered: boolean }
  >({
    accept: "inline-processor",
    collect: (monitor) => ({
      isHovered: monitor.isOver(),
    }),
    hover: (item, monitor) => {
      if (ref.current == null) {
        return;
      }

      if (monitor == null) {
        return;
      }

      const dragIndex = item.index;
      const hoverIndex = index;

      moveProcessor(dragIndex, hoverIndex);
      item.index = hoverIndex;
    },

    // Save the order on drop
    drop: onDrop,
  });

  // Join the 2 refs together into one (both draggable and can be dropped on)
  const ref = useRef<HTMLDivElement>(null);

  const dragDropRef = dragRef(dropRef(ref)) as any;

  return (
    <ProcessorLabelCard
      index={index}
      processor={processor}
      dragDropRef={dragDropRef}
      isHovered={isHovered}
      onEdit={onEdit}
    />
  );
};
