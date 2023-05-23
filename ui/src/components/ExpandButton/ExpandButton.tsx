import { Button } from "@mui/material";
import { ChevronDown, ChevronUp } from "../Icons";

import styles from "./expand-button.module.scss";

interface ExpandButtonProps {
  expanded: boolean;
  onToggleExpanded: () => void;
}

/**
 * ExpandButton is a button that shows "Show More" or "Show Less" depending on the
 * expanded prop.
 */
export const ExpandButton: React.FC<ExpandButtonProps> = ({
  expanded,
  onToggleExpanded,
}) => {
  return (
    <Button
      className={styles.button}
      variant="contained"
      color="secondary"
      size="small"
      onClick={onToggleExpanded}
      endIcon={expanded ? <ChevronUp /> : <ChevronDown />}
    >
      {expanded ? "Show less" : "Show more"}
    </Button>
  );
};
