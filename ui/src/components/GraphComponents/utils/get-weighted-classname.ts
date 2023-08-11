import { MaxValueMap, CustomEdgeData } from "../CustomEdge";

import styles from "../custom-edge.module.scss";

function determineMaxValue(telmetryType: string, maxValues: MaxValueMap) {
  switch (telmetryType) {
    case "metrics":
      return maxValues.maxMetricValue;
    case "logs":
      return maxValues.maxLogValue;
    case "traces":
      return maxValues.maxTraceValue;
  }

  return null;
}

export type getWeightedClassNameFunc = (
  metrics: CustomEdgeData["metrics"] | undefined,
  maxValues: MaxValueMap,
  telemetryType: string
) => string;

export const getOverviewWeightedClassName: getWeightedClassNameFunc = (
  metrics,
  maxValues,
  telemetryType
) => {
  const maxValue = determineMaxValue(telemetryType, maxValues);

  if (
    maxValue === null ||
    metrics == null ||
    metrics.length === 0 ||
    metrics[0] == null
  ) {
    return styles.inactive;
  }

  const metric = metrics.find((m) => m.rawValue != null);

  if (!metric) {
    return styles.inactive;
  }

  return getWidthClass(metric.rawValue, maxValue);
};

function getWidthClass(rawValue: number, maxValue: number) {
  const ratio = rawValue / maxValue;
  if (ratio >= 1) {
    return styles.w5;
  }
  const scaled = Math.floor(ratio * 5 + 1);
  const widthStyle = `w${scaled}`;
  return styles[widthStyle];
}

export const getWeightedClassName: getWeightedClassNameFunc = (
  metrics: CustomEdgeData["metrics"] | undefined,
  maxValues: MaxValueMap,
  telemetryType: string
) => {
  const maxValue = determineMaxValue(telemetryType, maxValues);

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

  return getWidthClass(metrics[0].rawValue, maxValue);
};
