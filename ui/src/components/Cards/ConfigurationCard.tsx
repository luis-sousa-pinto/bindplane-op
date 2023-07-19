import { Chip, Stack } from "@mui/material";
import { useLocation, useNavigate } from "react-router-dom";

import { CardMeasurementContent } from "../CardMeasurementContent/CardMeasurementContent";
import { SlidersIcon } from "../Icons";
import { ResourceCard } from "./ResourceCard";

import styles from "./cards.module.scss";

interface ConfigurationCardProps {
  id: string;
  label: string;
  attributes: Record<string, any>;
  metric: string;
  disabled?: boolean;
}

export const ConfigurationCard: React.FC<ConfigurationCardProps> = ({
  id,
  label,
  attributes,
  metric,
  disabled,
}) => {
  const navigate = useNavigate();
  const location = useLocation();
  const isEverything = id === "everything/configuration";
  const configurationURL = isEverything
    ? "/configurations"
    : `/configurations/${id.split("/").pop()}`;
  const agentCount = attributes["agentCount"] ?? 0;

  return (
    <div className={disabled ? styles.disabled : undefined}>
      <ResourceCard
        name={label}
        disabled={disabled}
        altIcon={
          isEverything ? (
            <Stack spacing={1}>
              <Stack direction="row" spacing={1}>
                <SlidersIcon height="20px" width="20px" />
                <SlidersIcon height="20px" width="20px" />
              </Stack>
              <Stack direction="row" spacing={1}>
                <SlidersIcon height="20px" width="20px" />
                <SlidersIcon height="20px" width="20px" />
              </Stack>
            </Stack>
          ) : (
            <SlidersIcon height="40px" width="40px" />
          )
        }
        onClick={() =>
          navigate({ pathname: configurationURL, search: location.search })
        }
      />
      <Chip
        classes={{
          root: styles["overview-count-chip"],
          label: styles["overview-count-chip-label"],
        }}
        size="small"
        label={formatAgentCount(agentCount)}
      />
      <CardMeasurementContent>{metric}</CardMeasurementContent>
    </div>
  );
};
export function formatAgentCount(agentCount: number): string {
  switch (agentCount) {
    case 0:
      return "";
    case 1:
      return "1 agent";
    default:
      return `${agentCount} agents`;
  }
}
