import { CustomEdgeData, getWeightedClassName } from "./CustomEdge";

import styles from "./edge-styles.module.scss";

describe("getWeightedClassName", () => {
  function metricWithRawValue(rawValue: number): CustomEdgeData["metrics"] {
    return [
      {
        rawValue: rawValue,
        value: "",
        startOffset: "",
        textAnchor: "",
      },
    ];
  }

  const maxValues = {
    maxMetricValue: 100,
    maxLogValue: 10,
    maxTraceValue: 10,
  };

  it("returns the correct class", () => {
    expect(
      getWeightedClassName(metricWithRawValue(0), maxValues, "metrics")
    ).toEqual(styles["w1"]);
    expect(
      getWeightedClassName(metricWithRawValue(15), maxValues, "metrics")
    ).toEqual(styles["w1"]);
    expect(
      getWeightedClassName(metricWithRawValue(20), maxValues, "metrics")
    ).toEqual(styles["w2"]);
    expect(
      getWeightedClassName(metricWithRawValue(79), maxValues, "metrics")
    ).toEqual(styles["w4"]);
    expect(
      getWeightedClassName(metricWithRawValue(80), maxValues, "metrics")
    ).toEqual(styles["w5"]);
    expect(
      getWeightedClassName(metricWithRawValue(100), maxValues, "metrics")
    ).toEqual(styles["w5"]);
    expect(
      getWeightedClassName(metricWithRawValue(105), maxValues, "metrics")
    ).toEqual(styles["w5"]);
  });
});
