import {
  Card,
  CardActionArea,
  CardContent,
  Stack,
  Typography,
} from "@mui/material";
import { useSnackbar } from "notistack";
import { memo } from "react";

import { useGetDestinationWithTypeQuery } from "../../graphql/generated";
import { useOverviewPage } from "../../pages/overview/OverviewPageContext";
import { useLocation, useNavigate } from "react-router-dom";
import React from "react";
import { SquareIcon } from "../Icons";

import { classes } from "../../utils/styles";
import styles from "./cards.module.scss";
import { NoMaxWidthTooltip } from "../Custom/NoMaxWidthTooltip";
import { truncateLabel } from "../../utils/graph/utils";

interface ResourceDestinationCardProps {
  id: string;
  label: string;
  // disabled indicates that the card is not active and should be greyed out
  disabled?: boolean;
}

const OverviewDestinationCardComponent: React.FC<
  ResourceDestinationCardProps
> = ({ id, label, disabled }) => {
  const { enqueueSnackbar } = useSnackbar();

  const isEverythingDestination = id === "everything/destination";

  const { data } = useGetDestinationWithTypeQuery({
    variables: { name: id },
    fetchPolicy: "cache-and-network",
  });

  const navigate = useNavigate();
  const location = useLocation();
  const { setEditingDestination } = useOverviewPage();

  // Loading
  if (data === undefined) {
    return null;
  }
  if (
    !isEverythingDestination &&
    data.destinationWithType.destination == null
  ) {
    enqueueSnackbar(`Could not retrieve destination ${id}.`, {
      variant: "error",
    });
    return null;
  }

  if (
    !isEverythingDestination &&
    data.destinationWithType.destinationType == null
  ) {
    enqueueSnackbar(
      `Could not retrieve destination type for destination ${id}.`,
      { variant: "error" }
    );
    return null;
  }

  return (
    <div
      className={classes([
        disabled ? styles.disabled : undefined,
        data.destinationWithType.destination?.spec.disabled
          ? styles.paused
          : undefined,
      ])}
    >
      <Card
        className={classes([
          styles["resource-card"],
          disabled ? styles.disabled : undefined,
          data.destinationWithType.destination?.spec.disabled
            ? styles.paused
            : undefined,
        ])}
        onClick={() => {
          if (isEverythingDestination) {
            navigate({
              pathname: "/destinations",
              search: location.search,
            });
          } else {
            setEditingDestination(id);
          }
        }}
      >
        <CardActionArea className={styles.action}>
          <NoMaxWidthTooltip title={label.length > 20 ? label : ""}>
            <CardContent>
              <Stack alignItems="center" spacing={1}>
                {isEverythingDestination ? (
                  <SquareIcon className={styles["destination-icon"]} />
                ) : (
                  <span
                    className={styles.icon}
                    style={{
                      backgroundImage: `url(${data?.destinationWithType?.destinationType?.metadata.icon})`,
                    }}
                  />
                )}
                <Typography
                  align="center"
                  component="div"
                  fontWeight={600}
                  gutterBottom
                  fontSize={label.length > 15 ? 11 : 16}
                >
                  {truncateLabel(label, 20)}
                </Typography>
                {data.destinationWithType.destination?.spec.disabled && (
                  <Typography
                    component="div"
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
    </div>
  );
};

export const OverviewDestinationCard = memo(OverviewDestinationCardComponent);
