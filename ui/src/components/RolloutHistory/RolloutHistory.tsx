import { gql } from "@apollo/client";
import { Box, CircularProgress, Collapse, Typography } from "@mui/material";
import { Stack } from "@mui/system";
import { useMemo, useState } from "react";
import {
  GetRolloutHistoryQuery,
  useGetRolloutHistoryQuery,
} from "../../graphql/generated";
import { format } from "date-fns";
import { useRefetchOnConfigurationChange } from "../../hooks/useRefetchOnConfigurationChanges";

import styles from "./rollout-history.module.scss";
import { ExpandButton } from "../ExpandButton";

gql`
  query getRolloutHistory($name: String!) {
    configurationHistory(name: $name) {
      metadata {
        name
        id
        version
        dateModified
      }
      status {
        rollout {
          status
          errors
        }
      }
    }
  }
`;

interface RolloutHistoryProps {
  configurationName: string;
}

export const RolloutHistory: React.FC<RolloutHistoryProps> = ({
  configurationName,
}) => {
  const { data, refetch } = useGetRolloutHistoryQuery({
    variables: {
      name: configurationName,
    },
    fetchPolicy: "cache-and-network",
  });

  useRefetchOnConfigurationChange(configurationName, refetch);

  const messages = useMemo(() => {
    if (!data) {
      return [];
    }

    return makeMessages(data);
  }, [data]);

  // state to control whether the accordion is expanded or not
  const [expanded, setExpanded] = useState(false);

  return (
    <Box
      className={styles.box}
      aria-describedby={"rollout-history-loading"}
      aria-busy={data == null}
    >
      <Stack>
        <Typography fontSize={18} fontWeight={600} marginBottom="8px">
          Rollout History
        </Typography>
        {messages.slice(0, 1)}
      </Stack>
      {data == null ? (
        <Stack width="100%" alignItems={"center"} justifyContent="center">
          <CircularProgress
            size={24}
            id="rollout-history-loading"
            data-testid="circular-progress"
            disableShrink
          />
        </Stack>
      ) : (
        <Collapse in={expanded} collapsedSize={0}>
          {messages.slice(1)}
        </Collapse>
      )}

      {
        // only show the expand button if there are more than 1 messages
        messages.length > 1 && (
          <Stack direction="row" justifyContent="center" alignItems="center">
            <ExpandButton
              expanded={expanded}
              onToggleExpanded={() => setExpanded((prev) => !prev)}
            />
          </Stack>
        )
      }
    </Box>
  );
};

/**
 * makeMessages takes in the data from the getRolloutHistory query and returns
 * an array of up to 10 messages to be displayed in the RolloutHistory component.
 *
 * @param data the data from the getRolloutHistory query
 */
function makeMessages(
  data: NonNullable<GetRolloutHistoryQuery>
): JSX.Element[] {
  const messages = data.configurationHistory.map((history) => {
    const { metadata, status } = history;
    const { version, dateModified } = metadata;
    const { rollout } = status;
    const { status: rolloutStatus, errors } = rollout;

    const date = new Date(dateModified);

    const action = rolloutStatusToAction[rolloutStatus];
    const withErrors =
      errors > 0 ? (
        <>
          <Typography component="span"> with</Typography>
          <Typography color="error" component="span">
            {" "}
            {errors} error{errors > 1 && "s"}
          </Typography>
        </>
      ) : null;

    return (
      <Typography key={`${date.toString()}-${version}-${status}`}>
        Version {version} {action}
        {withErrors} on {formatDate(date)} at {formatTime(date)}
      </Typography>
    );
  });

  // only return the first 10 messages
  return messages.slice(0, 10);
}

export function formatDate(date: Date) {
  return format(date, "M/dd/yyyy");
}

export function formatTime(date: Date) {
  return format(date, "HH:mm");
}

const rolloutStatusToAction: Record<number, string> = {
  0: "pending rollout",
  1: "rollout started",
  2: "rollout paused",
  3: "rollout paused",
  4: "completed",
  5: "rollout replaced",
};
