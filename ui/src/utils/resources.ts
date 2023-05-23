import { Configuration, ConfigurationSpec } from "../graphql/generated";
import { APIVersion, ResourceKind } from "../types/resources";

export function newConfiguration({
  name,
  description,
  spec,
  labels,
}: {
  name: string;
  description: string;
  spec: ConfigurationSpec;
  labels?: { [key: string]: string };
}): Pick<Configuration, "apiVersion" | "kind" | "metadata" | "spec"> {
  return {
    apiVersion: APIVersion.V1,
    kind: ResourceKind.CONFIGURATION,
    metadata: { name, description, labels, id: "", version: 1 },
    spec,
  };
}
