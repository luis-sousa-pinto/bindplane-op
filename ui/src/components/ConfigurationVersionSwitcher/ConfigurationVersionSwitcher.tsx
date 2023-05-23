import {
  Box,
  Button,
  MenuItem,
  Select,
  SelectChangeEvent,
  Stack,
  Tab,
  Tabs,
} from "@mui/material";
import { isEmpty } from "lodash";

import styles from "./configuration-version-switcher.module.scss";

export type ConfigurationVersionSwitcherTab =
  | "current"
  | "history"
  | "pending"
  | "new";

interface ConfigurationVersionSwitcherProps {
  tab: ConfigurationVersionSwitcherTab;
  currentVersion?: number;
  pendingVersion?: number;
  newVersion?: number;

  versionHistory: number[];
  selectedVersionHistory?: number;
  onSelectedVersionHistoryChange: (version: number) => void;

  onChange: (newTab: ConfigurationVersionSwitcherTab) => void;
  onEditNewVersion: () => void;
  allowEditPendingVersion?: boolean;
  onEditPendingVersion: () => void;
}

export const ConfigurationVersionSwitcher: React.FC<
  ConfigurationVersionSwitcherProps
> = ({
  tab,
  currentVersion,
  pendingVersion,
  newVersion,
  versionHistory,
  selectedVersionHistory,
  onChange,
  onEditNewVersion,
  onEditPendingVersion,
  allowEditPendingVersion,
  onSelectedVersionHistoryChange,
}) => {
  function handleChange(
    _e: React.SyntheticEvent,
    newTab: ConfigurationVersionSwitcherTab
  ) {
    onChange(newTab);
  }

  function handleEditNewVersionClick() {
    onChange("new");
    onEditNewVersion();
  }

  function handleSelectChange(e: SelectChangeEvent<number>) {
    onSelectedVersionHistoryChange?.(e.target.value as number);
  }

  return (
    <Box className={styles.box}>
      <Stack
        direction="row"
        justifyContent={"space-between"}
        alignItems="center"
      >
        <Tabs
          value={tab}
          onChange={handleChange}
          classes={{ root: styles.tabs }}
        >
          <Tab
            label="Current Version"
            value="current"
            style={{
              display: currentVersion ? "block" : "none",
            }}
          />

          <Tab
            label="Pending Version"
            value="pending"
            style={{
              display: pendingVersion ? "block" : "none",
            }}
          />

          <Tab
            label="New Version"
            value="new"
            style={{
              display: newVersion ? "block" : "none",
            }}
          />

          <Tab
            label="Version History"
            value="history"
            style={{
              display: !isEmpty(versionHistory) ? "block" : "none",
            }}
          />
        </Tabs>

        {tab === "current" && (
          <Button color="action-blue" onClick={handleEditNewVersionClick}>
            Edit New Version
          </Button>
        )}

        {tab === "pending" && allowEditPendingVersion && (
          <Button color="action-blue" onClick={onEditPendingVersion}>
            Edit Pending Version
          </Button>
        )}

        {tab === "history" && !isEmpty(versionHistory) && (
          <Select<number>
            size="small"
            value={selectedVersionHistory ?? versionHistory[0]}
            classes={{ select: styles.select }}
            onChange={handleSelectChange}
          >
            {versionHistory?.map((version) => (
              <MenuItem key={`select-version-${version}`} value={version}>
                Version {version}
              </MenuItem>
            ))}
          </Select>
        )}
      </Stack>
    </Box>
  );
};
