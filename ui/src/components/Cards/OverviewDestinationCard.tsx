import { useSnackbar } from "notistack";
import { memo } from "react";
import { useLocation, useNavigate } from "react-router-dom";
import React from "react";

import { useGetDestinationWithTypeQuery } from "../../graphql/generated";
import { useOverviewPage } from "../../pages/overview/OverviewPageContext";
import { SquareIcon } from "../Icons";
import { ResourceCard } from "./ResourceCard";

import styles from "./cards.module.scss";

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
    <ResourceCard
      name={label}
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
      icon={data?.destinationWithType?.destinationType?.metadata.icon}
      altIcon={<SquareIcon className={styles["destination-icon"]} />}
      paused={data.destinationWithType.destination?.spec.disabled}
      disabled={disabled}
    />
  );
};

export const OverviewDestinationCard = memo(OverviewDestinationCardComponent);
