import { isEmpty } from "lodash";
import {
  Maybe,
  Parameter,
  ResourceConfiguration,
} from "../../graphql/generated";

export class BPResourceConfiguration implements ResourceConfiguration {
  id?: Maybe<string> | undefined;
  name?: Maybe<string> | undefined;
  displayName: Maybe<string> | undefined;
  type?: Maybe<string> | undefined;
  parameters?: Maybe<Parameter[]> | undefined;
  processors?: Maybe<ResourceConfiguration[]> | undefined;
  disabled: boolean;
  constructor(rc?: ResourceConfiguration) {
    this.id = rc?.id;
    this.name = rc?.name;
    this.displayName = rc?.displayName;
    this.type = rc?.type;
    this.parameters = rc?.parameters;
    this.processors = rc?.processors;
    this.disabled = rc?.disabled ?? false;
  }

  isInline(): boolean {
    return isEmpty(this.name);
  }

  hasConfigurationParameters(): boolean {
    return this.parameters != null && this.parameters.length > 0;
  }

  // setParamsFromMap will set the parameters from Record<string, any>.
  // If the "name" key is specified it will set the name field of the ResourceConfiguration.
  // If the "processors" key is specified it will set the processors value.
  // It will not set undefined or null values to parameters.
  setParamsFromMap(map: Record<string, any>) {
    // Set name field if present
    if (map.name != null && map.name !== "") {
      this.name = map.name;
      delete map.name;
    }

    // Set displayName field if present
    if (map.displayName != null) {
      this.displayName = map.displayName;
      delete map.displayName;
    }

    // Set processors field if present
    if (map.processors != null) {
      this.processors = map.processors;
      delete map.processors;
    }

    // Set the parameters only if their values are not nullish.
    const parameters = Object.entries(map).reduce<Parameter[]>(
      (params, [name, value]) => {
        if (value != null) {
          params.push({ name, value });
        }
        return params;
      },
      []
    );

    this.parameters = parameters;
  }
}
