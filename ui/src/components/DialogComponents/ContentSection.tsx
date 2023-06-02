import { DialogContent } from "@mui/material";
import { memo } from "react";

import styles from "./dialog-components.module.scss";

const ContentSectionComponent: React.FC<React.PropsWithChildren<{ dividers?: boolean }>> = ({
  dividers,
  children,
}) => {
  return (
    <DialogContent dividers={dividers} classes={{ root: styles.content }}>{children}</DialogContent>
  );
};

export const ContentSection = memo(ContentSectionComponent);
