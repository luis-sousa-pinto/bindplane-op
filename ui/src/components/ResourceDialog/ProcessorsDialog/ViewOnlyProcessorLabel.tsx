import { ResourceConfiguration } from "../../../graphql/generated";
import { ProcessorLabelCard } from "./ProcessorLabelCard";

interface Props {
  index: number;
  processor: ResourceConfiguration;
  onEdit: () => void;
}

export const ViewOnlyProcessorLabel: React.FC<Props> = ({
  index,
  processor,
  onEdit,
}) => {
  return (
    <ProcessorLabelCard index={index} processor={processor} onEdit={onEdit} />
  );
};
