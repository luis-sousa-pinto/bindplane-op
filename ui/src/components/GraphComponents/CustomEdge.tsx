import { memo } from "react";
import { EdgeProps, getBezierPath } from "reactflow";
import { classes } from "../../utils/styles";
import { isNodeDisabled } from "../PipelineGraph/Nodes/nodeUtils";
import { getWeightedClassNameFunc } from "./utils/get-weighted-classname";

import styles from "./custom-edge.module.scss";

export interface CustomEdgeData {
  connectedNodesAndEdges: string[];
  metrics: {
    rawValue: number;
    value: string;
    startOffset: string;
    textAnchor: string;
  }[];
  active: boolean;
  attributes: Record<string, any>;
}

export interface MaxValueMap {
  maxMetricValue: number;
  maxLogValue: number;
  maxTraceValue: number;
}

export interface CustomEdgeProps extends EdgeProps<CustomEdgeData> {
  hoveredSet?: string[];
  className?: string;
  telemetryType: string;
  maxValues: MaxValueMap;
  getWeightedClassFunc: getWeightedClassNameFunc;
}

export const CustomEdge: React.FC<CustomEdgeProps> = ({
  id,
  sourceX,
  sourceY,
  targetX,
  targetY,
  sourcePosition,
  targetPosition,
  data,
  hoveredSet,
  className,
  telemetryType,
  maxValues,
  getWeightedClassFunc,
}) => {
  hoveredSet ||= [];

  const [path] = getBezierPath({
    sourceX,
    sourceY,
    sourcePosition,
    targetX,
    targetY,
    targetPosition,
  });

  const inactive = isNodeDisabled(telemetryType || "", data?.attributes || {});
  const dimmed = hoveredSet.length > 0 && !hoveredSet.includes(id);

  const metricLabels: JSX.Element[] = [];
  if (data?.metrics) {
    for (var i = 0; i < data.metrics.length; i++) {
      const m = data.metrics[i];
      const metric = (
        // transform moves the metric a few pixels off the line
        <g transform={`translate(0 -15)`} key={`metric${i}`}>
          <text>
            <textPath
              className={classes([
                styles["metric"],
                className ? styles[className] : undefined,
              ])}
              href={`#${id}`}
              startOffset={m.startOffset}
              textAnchor={m.textAnchor}
              spacing="auto"
            >
              {m.value}
            </textPath>
          </text>
        </g>
      );
      metricLabels.push(metric);
    }
  }

  const pathClasses = [styles.edge];
  dimmed && pathClasses.push(styles.dimmed);
  inactive && pathClasses.push(styles.inactive);
  !inactive && pathClasses.push(styles.gradient);
  !inactive &&
    pathClasses.push(
      getWeightedClassFunc(data?.metrics, maxValues, telemetryType)
    );

  return (
    <>
      {metricLabels}
      <path id={id} className={classes(pathClasses)} d={path}>
        <animate
          attributeName="stroke-dashoffset"
          from="0"
          to="-130"
          dur="10s"
          repeatCount="indefinite"
        />
      </path>
    </>
  );
};

export default memo(CustomEdge);
