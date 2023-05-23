import { Card, IconButton, Stack, Typography } from "@mui/material";
import { useSnackbar } from "notistack";
import { useEffect } from "react";
import {
  ResourceConfiguration,
  useGetProcessorTypeQuery,
} from "../../../graphql/generated";
import { MenuIcon, EditIcon } from "../../Icons";

import styles from "./inline-processor-label.module.scss";

interface Props {
  index: number;
  processor: ResourceConfiguration;
  onEdit: () => void;
}

export const ViewOnlyProcessorLabel: React.FC<Props> = ({
  index,
  processor,
  onEdit,
}) => {
  const { data, error } = useGetProcessorTypeQuery({
    variables: { type: processor.type! },
  });

  const { enqueueSnackbar } = useSnackbar();

  useEffect(() => {
    if (error != null) {
      console.error(error);
      enqueueSnackbar("Error retrieving Processor Type", {
        variant: "error",
        key: "Error retrieving Processor Type",
      });
    }
  }, [enqueueSnackbar, error]);

  return (
    <Card variant="outlined" classes={{ root: styles.card }}>
      <Stack
        direction="row"
        alignItems={"center"}
        spacing={1}
        justifyContent={"space-between"}
      >
        <Stack direction={"row"} spacing={1}>
          <MenuIcon />
          <Typography>{data?.processorType?.metadata.displayName}</Typography>
        </Stack>

        <IconButton onClick={onEdit} data-testid={`edit-processor-${index}`}>
          <EditIcon width={15} height={15} style={{ float: "right" }} />
        </IconButton>
      </Stack>
    </Card>
  );
};
