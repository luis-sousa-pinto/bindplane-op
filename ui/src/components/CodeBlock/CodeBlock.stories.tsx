import { Meta, StoryFn } from "@storybook/react";
import { InputType } from "zlib";
import { CodeBlock } from ".";

export default {
  title: "Code Block",
  component: CodeBlock,
  argTypes: {
    value: "something" as InputType,
  },
} as Meta<typeof CodeBlock>;

const Template: StoryFn<typeof CodeBlock> = (args) => (
  <CodeBlock {...args} />
);

export const Default = Template.bind({});

Default.args = {
  value: 'while true; do echo "Hello World"; sleep 1; done;',
};
