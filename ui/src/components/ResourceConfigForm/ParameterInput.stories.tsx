import { Meta, StoryFn } from "@storybook/react";
import { ParameterInput, StringParamInput } from "./ParameterInput";
import {
  stringDef,
  stringDefRequired,
  enumDef,
  stringsDef,
  boolDef,
  intDef,
  stringPasswordDef,
} from "./__test__/dummyResources";

export default {
  title: "Parameter Input",
  component: ParameterInput,
} as Meta<typeof StringParamInput>;

const Template: StoryFn<typeof StringParamInput> = (args) => (
  <div style={{ width: 400 }}>
    <ParameterInput {...args} />
  </div>
);

export const String = Template.bind({});
export const StringRequired = Template.bind({});
export const Password = Template.bind({});
export const Strings = Template.bind({});
export const Enum = Template.bind({});
export const Bool = Template.bind({});
export const Int = Template.bind({});

String.args = {
  definition: stringDef,
};

StringRequired.args = {
  definition: stringDefRequired,
};

Password.args = {
  definition: stringPasswordDef,
};

Enum.args = {
  definition: enumDef,
};

Strings.args = {
  definition: stringsDef,
};

Bool.args = {
  definition: boolDef,
};

Int.args = {
  definition: intDef,
};
