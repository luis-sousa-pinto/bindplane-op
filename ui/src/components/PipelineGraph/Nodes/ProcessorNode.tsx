import { Handle, Position } from "reactflow";
import { CardMeasurementContent } from "../../CardMeasurementContent/CardMeasurementContent";
import { ProcessorCard } from "../../Cards/ProcessorCard";
import { MinimumRequiredConfig } from "../PipelineGraph";

export function ProcessorNode({
  data,
}: {
  data: {
    // ID will have the form source/source0/processors, source/source1/processors, dest-name/destination/processors, etc
    id: string;
    metric: string;
    attributes: Record<string, any>;
    configuration: MinimumRequiredConfig;
  };
}) {
  const { id, metric, configuration, attributes } = data;

  const isSource = isSourceID(id);

  let processorCount = 0;
  let resourceIndex = -1;
  if (isSource) {
    if (typeof attributes["sourceIndex"] === "number") {
      resourceIndex = attributes["sourceIndex"];
    }
    processorCount =
      configuration?.spec?.sources![resourceIndex]?.processors?.length ?? 0;
  } else {
    if (typeof attributes["destinationIndex"] === "number") {
      resourceIndex = attributes["destinationIndex"];
    }

    const destination = configuration?.spec?.destinations![resourceIndex];
    processorCount = destination?.processors?.length ?? 0;
  }
  return (
    <>
      <Handle type="target" position={Position.Left} />
      <ProcessorCard
        processorCount={processorCount}
        resourceType={isSource ? "source" : "destination"}
        resourceIndex={resourceIndex}
      />
      <CardMeasurementContent>{metric}</CardMeasurementContent>
      <Handle type="source" position={Position.Right} />
    </>
  );
}

export function isSourceID(id: string): boolean {
  return id.startsWith("source/");
}

export function getDestinationName(id: string): string {
  // /destination/name/processors
  const REGEX = /^destination\/(?<name>.*)\/processors$/;
  const match = id.match(REGEX);
  return match?.groups?.["name"] ?? "";
}
