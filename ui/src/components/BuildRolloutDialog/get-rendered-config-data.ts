import { GetRenderedConfigQuery } from "../../graphql/generated";

export class RenderedConfigData implements GetRenderedConfigQuery {
  configuration: NonNullable<GetRenderedConfigQuery["configuration"]>;
  constructor(config: NonNullable<GetRenderedConfigQuery["configuration"]>) {
    this.configuration = config;
  }

  // title returns "Version <version>"
  title() {
    return `Version ${this.configuration.metadata.version}`;
  }

  value() {
    return this.configuration.rendered!.replaceAll("    ", "  ");
  }
}
