import { Button } from "@mui/material";
import { Meta, StoryFn } from "@storybook/react";

export default {
  title: "Button",
  component: Button,
  argTypes: {
    size: {
      options: ["small", "medium", "large"],
      control: { type: "radio" },
    },
    color: {
      options: ["primary", "secondary", "info", "error", "warning", "success"],
      control: { type: "radio" },
    },
  },
} as Meta<typeof Button>;

const Template: StoryFn<typeof Button> = (args) => (
  <Button {...args}>Button</Button>
);

export const Default = Template.bind({});
export const Contained = Template.bind({});
export const Outlined = Template.bind({});

Default.args = {};
Contained.args = {
  variant: "contained",
};
Outlined.args = {
  variant: "outlined",
};
