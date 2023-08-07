import { ParameterDefinition, ParameterType } from "../../../graphql/generated";
import { groupParameters } from "./group-parameters";

describe("parameter grouping", () => {
  const parameterDefinition: ParameterDefinition[] = [
    {
      name: "param1",
      label: "Param 1",
      description: "Param 1 description",
      required: true,
      type: ParameterType.String,
      options: {},
    },
    {
      name: "param2",
      label: "Param 2",
      description: "Param 2 description",
      type: ParameterType.String,
      required: false,
      options: {},
    },
    {
      name: "param3",
      label: "Param 3",
      description: "Param 3 description",
      type: ParameterType.String,
      required: false,
      options: {},
    },
    {
      name: "param4",
      label: "Param 4",
      description: "Param 4 description",
      type: ParameterType.String,
      required: false,
      advancedConfig: true,
      options: {},
    },
    {
      name: "param5",
      label: "Param 5",
      description: "Param 5 description",
      type: ParameterType.String,
      required: false,
      advancedConfig: true,
      options: {},
    },
    {
      name: "param6",
      label: "Param 6",
      description: "Param 6 description",
      type: ParameterType.String,
      required: false,
      advancedConfig: true,
      options: {},
    },
  ];

  it("groupParameters", () => {
    const got = groupParameters(parameterDefinition);

    expect(got).toHaveLength(2);
    expect(got[0].parameters[0].name).toEqual("param1");
    expect(got[0].parameters[1].name).toEqual("param2");
    expect(got[0].parameters[2].name).toEqual("param3");
    expect(got[1].parameters[0].name).toEqual("param4");
    expect(got[1].parameters[1].name).toEqual("param5");
    expect(got[1].parameters[2].name).toEqual("param6");
  });
});
