import {
  Box,
  Card,
  CardActionArea,
  CardContent,
  Stack,
  Typography,
} from "@mui/material";
import { classes } from "../../utils/styles";

import styles from "./cards.module.scss";
import { NoMaxWidthTooltip } from "../Custom/NoMaxWidthTooltip";
import { truncateLabel } from "../../utils/graph/utils";

interface ResourceCardProps {
  name: string;
  displayName?: string;
  icon?: string | null;
  paused?: boolean;
  disabled?: boolean;
  onClick: () => void;
  dataTestID?: string;
}

export const ResourceCard: React.FC<ResourceCardProps> = ({
  name,
  displayName,
  icon,
  paused,
  disabled,
  onClick,
  dataTestID,
}) => {
  return (
    <Card
      className={classes([
        styles["resource-card"],
        disabled ? styles.disabled : undefined,
        paused ? styles.paused : undefined,
      ])}
      onClick={onClick}
    >
      <CardActionArea className={styles.action}>
        <NoMaxWidthTooltip title={name.length > 20 ? name : ""}>
          <CardContent data-testid={dataTestID}>
            <Stack spacing={0} alignItems="left">
              <Stack
                direction={"row"}
                alignItems="center"
                justifyContent="space-between"
                spacing={1}
              >
                <Box
                  className={styles.icon}
                  style={{
                    backgroundImage: `url(${icon})`,
                  }}
                />
                <Stack flexGrow={1} alignItems={"center"} width="100%">
                  <Typography
                    fontWeight={600}
                    fontSize={name.length > 15 ? 11 : 16}
                  >
                    {truncateLabel(name, 20)}
                  </Typography>
                  {displayName && (
                    <NoMaxWidthTooltip
                      title={
                        displayName === truncateLabel(displayName, 20)
                          ? ""
                          : displayName
                      }
                    >
                      <Typography fontWeight={600} fontSize={12}>
                        {truncateLabel(displayName, 20)}
                      </Typography>
                    </NoMaxWidthTooltip>
                  )}
                </Stack>
              </Stack>
              {paused && (
                <Typography
                  align="center"
                  fontWeight={400}
                  fontSize={14}
                  variant="overline"
                >
                  Paused
                </Typography>
              )}
            </Stack>
          </CardContent>
        </NoMaxWidthTooltip>
      </CardActionArea>
    </Card>
  );
};
