import { Edge, Node, Position } from "reactflow";
import { TARGET_OFFSET_MULTIPLIER } from "../../components/PipelineGraph/ConfigurationFlow";
import { Graph, GraphMetric } from "../../graphql/generated";
import {
  formatMetric,
  getMetricForNode,
  getNodesAndEdges,
  Page,
  truncateLabel,
  updateMetricData,
} from "./utils";

describe("getMetricForNode", () => {
  const agentID = "01GDNQNEHQ2DFKV84TJFSRBF53";

  it("should return the correct metric", () => {
    const metrics: GraphMetric[] = [
      {
        name: "log_data_size",
        nodeID: "source/source0",
        value: 2840.45,
        unit: "B/s",
        agentID,
        pipelineType: "logs",
      },
      {
        name: "log_data_size",
        nodeID: "source/source0/processors",
        value: 2840.45,
        unit: "B/s",
        agentID,
        pipelineType: "logs",
      },
      {
        name: "log_data_size",
        nodeID: "destination/otlphttp/processors",
        value: 2840.45,
        unit: "B/s",
        agentID,
        pipelineType: "logs",
      },
      {
        name: "log_data_size",
        nodeID: "destination/otlphttp",
        value: 2840.45,
        unit: "B/s",
        agentID,
        pipelineType: "logs",
      },
      {
        name: "metric_data_size",
        nodeID: "source/source0__processor0",
        value: 18.3,
        unit: "B/s",
        agentID,
        pipelineType: "metrics",
      },
      {
        name: "metric_data_size",
        nodeID: "destination/otlphttp/processors",
        value: 18.3,
        unit: "B/s",
        agentID,
        pipelineType: "metrics",
      },
      {
        name: "metric_data_size",
        nodeID: "destination/otlphttp",
        value: 18.3,
        unit: "B/s",
        agentID,
        pipelineType: "metrics",
      },
    ];

    // It finds the metric_data_size metric for the route receiver of source/source0
    expect(
      getMetricForNode("source/source0/processors", metrics, "metrics")
    ).toEqual(metrics[4]);

    // It finds nothing because the nodeID is for a source with no metrics metrics
    expect(
      getMetricForNode("source/source0", metrics, "metrics")
    ).toBeUndefined();

    // It finds the log_data_size metric for source0's processors
    expect(
      getMetricForNode("source/source0/processors", metrics, "logs")
    ).toEqual(metrics[1]);
  });

  it("should handle sources with both logs and metrics", () => {
    const metrics: GraphMetric[] = [
      {
        agentID,
        name: "log_data_size",
        nodeID: "destination/otlphttp",
        pipelineType: "logs",
        value: 14662.07,
        unit: "B/s",
      },
      {
        agentID,
        name: "log_data_size",
        nodeID: "source/source2",
        pipelineType: "logs",
        value: 9369.14,
        unit: "B/s",
      },
      {
        agentID,
        name: "log_data_size",
        nodeID: "source/source2/processors",
        pipelineType: "logs",
        value: 9369.14,
        unit: "B/s",
      },
      {
        agentID,
        name: "log_data_size",
        nodeID: "destination/otlphttp/processors",
        pipelineType: "logs",
        value: 14662.07,
        unit: "B/s",
      },
      {
        agentID,
        name: "metric_data_size",
        nodeID: "destination/otlphttp",
        pipelineType: "metrics",
        value: 4778.68,
        unit: "B/s",
      },
      {
        agentID,
        name: "metric_data_size",
        nodeID: "source/source0__processor0",
        pipelineType: "metrics",
        value: 24.4,
        unit: "B/s",
      },
      {
        agentID,
        name: "metric_data_size",
        nodeID: "source/source2",
        pipelineType: "metrics",
        value: 3611.76,
        unit: "B/s",
      },
      {
        agentID,
        name: "metric_data_size",
        nodeID: "source/source2/processors",
        pipelineType: "metrics",
        value: 3611.76,
        unit: "B/s",
      },
      {
        agentID,
        name: "metric_data_size",
        nodeID: "source/source2__processor0",
        pipelineType: "metrics",
        value: 13.3,
        unit: "B/s",
      },
      {
        agentID,
        name: "metric_data_size",
        nodeID: "destination/otlphttp/processors",
        pipelineType: "metrics",
        value: 4778.68,
        unit: "B/s",
      },
    ];

    // Combine metrics from the source and its route receiver
    expect(
      getMetricForNode("source/source2/processors", metrics, "metrics")
    ).toEqual({
      agentID,
      name: "metric_data_size",
      nodeID: "source/source2/processors",
      pipelineType: "metrics",
      unit: "B/s",
      value: 3625.0600000000004,
    });

    expect(
      getMetricForNode("source/source2/processors", metrics, "logs")
    ).toEqual(metrics[2]);
  });
});

describe("formatMetric", () => {
  it("converts a metric to a human readable string", () => {
    expect(formatMetric({ value: 10, unit: "B/s" }, "/s")).toEqual("10 B/s");
  });

  it("converts to greater units if needed", () => {
    expect(formatMetric({ value: 2048, unit: "KiB/s" }, "/s")).toEqual(
      "2 MiB/s"
    );
  });

  it("converts to the requested rate", () => {
    expect(formatMetric({ value: 1024, unit: "MiB/s" }, "/m")).toEqual(
      "60 GiB/m"
    );

    expect(formatMetric({ value: 1024, unit: "MiB/s" }, "/h")).toEqual(
      "3.5 TiB/h"
    );
  });
});

describe("updateMetricData", () => {
  it("should set a node's metric as blank if there's not a matching metric", () => {
    const node: Node = {
      id: "source/source0",
      type: "source",
      data: {
        metric: "10 B/s",
      },
      position: { x: 0, y: 0 },
    };
    const metrics: GraphMetric[] = [
      {
        name: "log_data_size",
        nodeID: "source/source0",
        value: 2840.45,
        unit: "B/s",
        agentID: "agentID",
        pipelineType: "logs",
      },
    ];

    updateMetricData(Page.Configuration, [node], [], metrics, "1m", "metrics");

    expect(node.data.metric).toEqual("");
  });

  it("should convert any matching metric to the given rate", () => {
    const node: Node = {
      id: "source/source1",
      type: "source",
      data: {},
      position: { x: 0, y: 0 },
    };
    const edge: Edge<any> = {
      id: "source/source1|source/source1/processors",
      source: "source/source1",
      target: "source/source1/processors",
      type: "configurationEdge",
      data: {},
    };

    const metrics: GraphMetric[] = [
      {
        name: "log_data_size",
        nodeID: "source/source1",
        value: 2840.45,
        unit: "B/s",
        agentID: "agentID",
        pipelineType: "logs",
      },
    ];

    updateMetricData(Page.Configuration, [node], [edge], metrics, "1m", "logs");

    expect(edge.data.metrics).toEqual([
      {
        rawValue: 2840.45,
        value: "166.4 KiB/m",
        startOffset: "50%",
        textAnchor: "middle",
      },
      {
        rawValue: 2840.45,
        startOffset: "50%",
        textAnchor: "middle",
        value: "",
      },
    ]);
  });
});

describe("getNodesAndEdges", () => {
  it("just has 'Add' buttons if there are no sources or targets", () => {
    const graph: Graph = {
      sources: [],
      targets: [],
      attributes: [],
      intermediates: [],
      edges: [],
    };
    const setAddDestDialogOpen = () => {};
    const setAddSourceDialogOpen = () => {};
    const targetOffsetMultiplier = TARGET_OFFSET_MULTIPLIER;

    const { nodes, edges } = getNodesAndEdges(
      Page.Configuration,
      graph,
      targetOffsetMultiplier,
      {},
      () => {},
      setAddSourceDialogOpen,
      setAddDestDialogOpen,
      false,
      ""
    );

    expect(nodes).toHaveLength(4);
    expect(edges).toHaveLength(3);
    expect(nodes[0]).toEqual({
      id: "add-source",
      data: {
        id: "add-source",
        buttonText: "Add Source",
        connectedNodesAndEdges: [
          "add-source",
          "add-source-proc",
          "add-source-proc",
          "add-destination-proc",
          "add-destination-proc",
          "add-destination",
          "add-destination-proc_add-destination",
          "add-source-proc_add-destination-proc",
          "add-source_add-source-proc",
        ],
        onClick: setAddSourceDialogOpen,
        handlePosition: Position.Right,
        handleType: "source",
      },
      position: { x: 8, y: 0 },
      type: "uiControlNode",
    });

    expect(nodes[2]).toEqual({
      id: "add-destination",
      data: {
        id: "add-destination",
        buttonText: "Add Destination",
        connectedNodesAndEdges: [
          "add-destination",
          "add-destination-proc",
          "add-destination-proc",
          "add-source-proc",
          "add-source-proc",
          "add-source",
          "add-source_add-source-proc",
          "add-source-proc_add-destination-proc",
          "add-destination-proc_add-destination",
        ],
        onClick: setAddDestDialogOpen,
        handlePosition: Position.Left,
        handleType: "target",
        isButton: false,
      },
      position: { x: 3 * targetOffsetMultiplier, y: 0 },
      type: "uiControlNode",
    });
  });
});

describe("truncateLabel", () => {
  it("won't truncate labels containing space(s)", () => {
    expect(truncateLabel("word1 word2", 5)).toEqual("word1 word2");
  });
  it("truncates with ellipsis", () => {
    expect(truncateLabel("word1word2", 5)).toEqual("word1...");
  });
});
