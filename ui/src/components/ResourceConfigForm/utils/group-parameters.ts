import { Maybe, ParameterDefinition } from "../../../graphql/generated";

export interface ParameterGroup {
  advanced: boolean;
  parameters: ParameterDefinition[];
  subHeader?: Maybe<string>;
}

export function groupParameters(
  parameters: ParameterDefinition[]
): ParameterGroup[] {
  const groups: ParameterGroup[] = [];
  let group: ParameterGroup | undefined;

  for (const p of parameters) {
    const advanced = p.advancedConfig ?? false;
    if (group == null || advanced !== group.advanced) {
      // start a new group
      group = {
        advanced,
        parameters: [],
      };
      groups.push(group);
    }
    group.parameters.push(p);
  }

  return groups;
}

export function groupBySubHeading(
  parameters: ParameterDefinition[]
): ParameterGroup[] {
  const groups: ParameterGroup[] = [];
  let group: ParameterGroup | undefined;

  for (const p of parameters) {
    const subHeader = p.options.subHeader;
    if (group == null || subHeader !== group.subHeader) {
      // start a new group
      group = {
        advanced: false,
        subHeader,
        parameters: [],
      };
      groups.push(group);
    }
    group.parameters.push(p);
  }

  return groups;
}
