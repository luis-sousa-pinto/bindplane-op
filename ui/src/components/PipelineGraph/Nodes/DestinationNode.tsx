import { memo } from "react";
import { Handle, Position } from "reactflow";
import { CardMeasurementContent } from "../../CardMeasurementContent/CardMeasurementContent";
import { ResourceDestinationCard } from "../../Cards/ResourceDestinationCard";
import { usePipelineGraph } from "../PipelineGraphContext";
import { isNodeDisabled } from "./nodeUtils";

function DestinationNode(params: {
  data: {
    pipelineType: string;
    id: string;
    label: string;
    attributes: Record<string, any>;
    metric: string;
    connectedNodesAndEdges: string[];
  };
}): JSX.Element {
  const { id, attributes, metric } = params.data;
  const destinationIndex =
    typeof attributes["destinationIndex"] === "number"
      ? attributes["destinationIndex"]
      : -1;
  const { hoveredSet, setHoveredNodeAndEdgeSet, selectedTelemetryType } =
    usePipelineGraph();
  const isDisabled = isNodeDisabled(selectedTelemetryType, attributes);
  const isNotInHoverSet =
    hoveredSet.length > 0 &&
    !hoveredSet.find((elem) => elem === params.data.id);
  return (
    <div
      onMouseEnter={() => {
        setHoveredNodeAndEdgeSet(params.data.connectedNodesAndEdges);
      }}
      onMouseLeave={() => setHoveredNodeAndEdgeSet([])}
    >
      <ResourceDestinationCard
        key={id}
        destinationIndex={destinationIndex}
        name={attributes.resourceId}
        disabled={isDisabled || isNotInHoverSet}
      />
      <CardMeasurementContent>{metric}</CardMeasurementContent>
      <Handle type="target" position={Position.Left} />
    </div>
  );
}

export default memo(DestinationNode);
