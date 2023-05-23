import {
  Card,
  CardActionArea,
  CardContent,
  Stack,
  Typography,
} from "@mui/material";
import React from "react";
import { PlusCircleIcon } from "../Icons";
import { classes } from "../../utils/styles";
import { usePipelineGraph } from "../PipelineGraph/PipelineGraphContext";

import styles from "./cards.module.scss";

interface AddResourceCardProps {
  onClick: () => void;
  buttonText: string;
}

export const AddResourceCard: React.FC<AddResourceCardProps> = ({
  onClick,
  buttonText,
}) => {
  const { readOnlyGraph } = usePipelineGraph();

  const canEdit = !readOnlyGraph;

  const classNames = classes([
    styles["ui-control-card"],
    canEdit ? undefined : styles.noninteractable,
  ]);

  return (
    <Card className={classNames} onClick={onClick}>
      <CardActionArea
        style={{
          cursor: canEdit ? "pointer" : "default",
        }}
      >
        <CardContent>
          <Stack justifyContent="center" alignItems="center" gap={1}>
            <PlusCircleIcon className={styles["ui-control-icon"]} />
            <Typography className={styles["ui-control-text"]}>
              {buttonText}
            </Typography>
          </Stack>
        </CardContent>
      </CardActionArea>
    </Card>
  );
};
