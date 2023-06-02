import { DialogActions, SxProps, Theme } from "@mui/material";
import { memo } from "react";

import styles from "./dialog-components.module.scss";

interface Props {
  /**
   * The system prop that allows defining system overrides as well as additional CSS styles.
   */
  sx?: SxProps<Theme>;
}

const ActionSectionComponent: React.FC<React.PropsWithChildren<Props>> = ({
  children,
  sx,
}) => {
  return (
    <DialogActions
      classes={{
        root: styles.actions,
      }}
      sx={sx}
    >
      {children}
    </DialogActions>
  );
};

export const ActionsSection = memo(ActionSectionComponent);
