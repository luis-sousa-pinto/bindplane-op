import { Configuration, Destination, Source } from "../graphql/generated";

export type Resource =
  | Pick<Configuration, "metadata" | "kind" | "spec" | "apiVersion">
  | Pick<Source, "metadata" | "kind" | "spec" | "apiVersion">
  | Pick<Destination, "metadata" | "kind" | "spec" | "apiVersion">;

/** ResourceStatus contains a resource and its UpdateStatus after a change */
export interface ResourceStatus {
  resource: Resource;
  status: UpdateStatus;
  reason?: string;
}

export enum APIVersion {
  V1 = "bindplane.observiq.com/v1",
}

export enum ResourceKind {
  CONFIGURATION = "Configuration",
  DESTINATION = "Destination",
  SOURCE = "Source",
  DESTINATION_TYPE = "DestinationType",
  SOURCE_TYPE = "SourceType",
}

export enum UpdateStatus {
  CREATED = "created",
  CONFIGURED = "configured",
  UNCHANGED = "unchanged",
  DELETED = "deleted",
  INVALID = "invalid",
}
