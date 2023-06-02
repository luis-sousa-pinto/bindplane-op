import { memo } from "react";
import { EdgeProps } from "reactflow";
import { usePipelineGraph } from "../PipelineGraphContext";

import { CustomEdge, CustomEdgeData } from "../../GraphComponents";

const ConfigurationEdge: React.FC<EdgeProps<CustomEdgeData>> = (props) => {
  const { selectedTelemetryType, hoveredSet, maxValues } = usePipelineGraph();

  return (
    <CustomEdge
      {...props}
      hoveredSet={hoveredSet}
      telemetryType={selectedTelemetryType}
      maxValues={maxValues}
    />
  );
};

export default memo(ConfigurationEdge);
