import { Destination } from "../../graphql/generated";
import { BPBaseResource } from "./base-resource";

export type MinimumDestination = Pick<Destination, "spec" | "metadata">;

export class BPDestination extends BPBaseResource implements Destination {
  __typename?: "Destination" | undefined;

  constructor(d: MinimumDestination) {
    super();

    this.kind = "Destination";
    this.metadata = d.metadata;
    this.spec = d.spec;
  }
}
