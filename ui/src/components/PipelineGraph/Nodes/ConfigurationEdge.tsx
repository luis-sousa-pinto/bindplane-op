import { memo } from "react";
import { EdgeProps } from "reactflow";
import { usePipelineGraph } from "../PipelineGraphContext";
import {
  CustomEdge,
  CustomEdgeData,
  getWeightedClassName,
} from "../../GraphComponents";

const ConfigurationEdge: React.FC<EdgeProps<CustomEdgeData>> = (props) => {
  const { selectedTelemetryType, hoveredSet, maxValues } = usePipelineGraph();

  return (
    <CustomEdge
      {...props}
      hoveredSet={hoveredSet}
      telemetryType={selectedTelemetryType}
      maxValues={maxValues}
      getWeightedClassFunc={getWeightedClassName}
    />
  );
};

export default memo(ConfigurationEdge);
