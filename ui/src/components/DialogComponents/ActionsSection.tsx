import { DialogActions } from "@mui/material";
import { memo } from "react";

import styles from "./dialog-components.module.scss";

interface Props {}

const ActionSectionComponent: React.FC<React.PropsWithChildren<Props>> = ({
  children,
}) => {
  return (
    <DialogActions
      classes={{
        root: styles.actions,
      }}
    >
      {children}
    </DialogActions>
  );
};

export const ActionsSection = memo(ActionSectionComponent);
