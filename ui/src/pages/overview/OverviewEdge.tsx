import React, { memo } from "react";
import { EdgeProps } from "reactflow";
import { useOverviewPage } from "./OverviewPageContext";
import { DEFAULT_TELEMETRY_TYPE } from "../../components/MeasurementControlBar/MeasurementControlBar";
import { CustomEdge, CustomEdgeData } from "../../components/GraphComponents";

const OverviewEdge: React.FC<EdgeProps<CustomEdgeData>> = (props) => {
  const { hoveredSet, selectedTelemetry, maxValues } = useOverviewPage();

  return (
    <CustomEdge
      {...props}
      hoveredSet={hoveredSet}
      telemetryType={selectedTelemetry ?? DEFAULT_TELEMETRY_TYPE}
      maxValues={maxValues}
    />
  );
};

export default memo(OverviewEdge);
