import { Processor } from "../../graphql/generated";
import { BPBaseResource } from "./base-resource";

export type MinimumProcessor = Pick<Processor, "spec" | "metadata">;

export class BPProcessor extends BPBaseResource implements Processor {
  __typename?: "Processor" | undefined;

  constructor(d: MinimumProcessor) {
    super();

    this.kind = "Processor";
    this.metadata = d.metadata;
    this.spec = d.spec;
  }
}
