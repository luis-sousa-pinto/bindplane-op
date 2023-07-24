import { cloneDeep } from "lodash";
import { Metadata, ParameterizedSpec } from "../../graphql/generated";
import { APIVersion, ResourceStatus } from "../../types/resources";
import { applyResources } from "../rest/apply-resources";

// BaseResource is the base class for Sources, Destinations, and Processors.
export class BPBaseResource {
  apiVersion: string;
  kind: string;
  metadata: Metadata;
  spec: ParameterizedSpec;

  constructor() {
    this.apiVersion = APIVersion.V1;
    this.kind = "Unknown";
    this.metadata = {
      id: "",
      version: 1,
      name: "",
    };

    this.spec = {
      type: "unknown",
      disabled: false,
    };
  }

  name(): string {
    return this.metadata.name;
  }

  // setParamsFromMap sets the spec.parameters from Record<string, any>.
  // If the "name" key is specified it will ignore it.
  setParamsFromMap(values: Record<string, any>) {
    const params: ParameterizedSpec["parameters"] = [];
    for (const [k, v] of Object.entries(values)) {
      switch (k) {
        case "name": // read-only
          break;
        case "processors": // saved in configuration
          break;
        default:
          params.push({
            name: k,
            value: v,
          });
      }
    }

    const newSpec = cloneDeep(this.spec);
    newSpec.parameters = params;
    this.spec = newSpec;
  }

  toggleDisabled() {
    const newSpec = cloneDeep(this.spec);
    newSpec.disabled = !newSpec.disabled;
    this.spec = newSpec;
  }

  async apply(): Promise<ResourceStatus> {
    const { updates } = await applyResources([this]);
    const update = updates.find(
      (u) => u.resource.metadata.name === this.name()
    );
    if (update == null) {
      throw new Error(
        `failed to apply configuration, no update with name ${this.name()}`
      );
    }
    return update;
  }
}
