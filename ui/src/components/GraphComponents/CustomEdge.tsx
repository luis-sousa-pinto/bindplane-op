import { memo } from "react";
import { EdgeProps, getBezierPath } from "react-flow-renderer";
import { classes } from "../../utils/styles";
import { isNodeDisabled } from "../PipelineGraph/Nodes/nodeUtils";

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
}) => {
  hoveredSet ||= [];

  const edgePath = getBezierPath({
    sourceX,
    sourceY,
    sourcePosition,
    targetX,
    targetY,
    targetPosition,
  });

  const inactive = isNodeDisabled(telemetryType || "", data?.attributes || {});
  const dimmed = hoveredSet.length > 0 && !hoveredSet.includes(id);
  const highlight = hoveredSet.includes(id);

  const metrics: JSX.Element[] = [];
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
      metrics.push(metric);
    }
  }

  const pathClasses = [styles.edge];
  dimmed && pathClasses.push(styles.dimmed);
  highlight && pathClasses.push(styles.highlight);
  inactive && pathClasses.push(styles.inactive);
  !inactive && pathClasses.push(styles.gradient);
  !inactive &&
    pathClasses.push(
      getWeightedClassName(data?.metrics, maxValues, telemetryType)
    );

  return (
    <>
      {metrics}
      <path id={id} className={classes(pathClasses)} d={edgePath}>
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

export function getWeightedClassName(
  metrics: CustomEdgeData["metrics"] | undefined,
  maxValues: MaxValueMap,
  telemetryType: string
) {
  var maxValue: number | null = null;

  switch (telemetryType) {
    case "metrics":
      maxValue = maxValues.maxMetricValue;
      break;
    case "logs":
      maxValue = maxValues.maxLogValue;
      break;
    case "traces":
      maxValue = maxValues.maxTraceValue;
      break;
  }

  if (
    maxValue === null ||
    metrics == null ||
    metrics.length === 0 ||
    metrics[0] == null
  ) {
    return styles.inactive;
  }
  // If the first metrics is the end of the edge, just discard it -
  // we don't have data for the beginning so it should be inactive
  if (metrics[0].textAnchor === "end") {
    return styles.inactive;
  }

  // Take the ratio of raw value to max value and scale it to the range of [1, 5]
  const ratio = metrics[0].rawValue / maxValue;
  if (ratio >= 1) {
    return styles.w5;
  }

  const scaled = Math.floor(ratio * 5 + 1);

  const widthStyle = `w${scaled}`;

  return styles[widthStyle];
}
