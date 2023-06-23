import { DialogActions, SxProps, Theme } from "@mui/material";
import { memo } from "react";

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
  return <DialogActions sx={sx}>{children}</DialogActions>;
};

export const ActionsSection = memo(ActionSectionComponent);
