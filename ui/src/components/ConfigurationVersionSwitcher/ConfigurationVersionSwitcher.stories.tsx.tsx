import { ComponentStory, ComponentMeta } from "@storybook/react";
import { useState } from "react";
import {
  ConfigurationVersionSwitcher,
  ConfigurationVersionSwitcherTab,
} from "./ConfigurationVersionSwitcher";

export default {
  title: "Configuration Version Switcher",
  component: ConfigurationVersionSwitcher,
} as ComponentMeta<typeof ConfigurationVersionSwitcher>;

const Template: ComponentStory<typeof ConfigurationVersionSwitcher> = (
  args
) => {
  const [selectedVersionHistory, setSelectedVersionHistory] = useState(3);
  const [tab, setTab] = useState<ConfigurationVersionSwitcherTab>("current");

  return (
    <div style={{ width: 800 }}>
      <ConfigurationVersionSwitcher
        {...args}
        tab={tab}
        onChange={setTab}
        onSelectedVersionHistoryChange={setSelectedVersionHistory}
        selectedVersionHistory={selectedVersionHistory}
      />
    </div>
  );
};

export const Default = Template.bind({});
Default.args = {
  onChange: (newTab: any) => {},
  onEditNewVersion: () => {},
  versionHistory: [1, 2, 3],
  currentVersion: 4,
  newVersion: 5,
};

export const RollingOut = Template.bind({});
RollingOut.args = {
  onChange: (newTab: any) => {},
  onEditNewVersion: () => {},
  versionHistory: [1, 2, 3],
  currentVersion: 4,
  pendingVersion: 5,
};
