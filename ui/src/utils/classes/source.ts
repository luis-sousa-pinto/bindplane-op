import { Source } from "../../graphql/generated";
import { BPBaseResource } from "./base-resource";

type MinimumSource = Pick<Source, "spec" | "metadata">;

export class BPSource extends BPBaseResource implements Source {
  __typename?: "Source" | undefined;

  constructor(s: MinimumSource) {
    super();
    this.kind = "Source";
    this.metadata = s.metadata;
    this.spec = s.spec;
  }
}
