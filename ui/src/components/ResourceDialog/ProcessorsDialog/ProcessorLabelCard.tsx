import { Card, Stack, Typography, IconButton } from "@mui/material";
import { MenuIcon, EditIcon } from "../../Icons";
import { ApolloError } from "@apollo/client";
import { useSnackbar } from "notistack";
import { useState } from "react";
import {
  ResourceConfiguration,
  useGetProcessorTypeQuery,
  useGetProcessorWithTypeQuery,
} from "../../../graphql/generated";
import { trimVersion } from "../../../utils/version-helpers";
import { BPResourceConfiguration } from "../../../utils/classes";

import styles from "./processor-label-card.module.scss";

interface LabelCardProps {
  index: number;
  processor: ResourceConfiguration;
  dragDropRef?: React.RefObject<HTMLDivElement>;
  isHovered?: boolean;
  onEdit: () => void;
}

export const ProcessorLabelCard: React.FC<LabelCardProps> = ({
  index,
  processor,
  dragDropRef,
  isHovered,
  onEdit,
}) => {
  const resourceConfig = new BPResourceConfiguration(processor);
  const { enqueueSnackbar } = useSnackbar();
  const [resourceTypeDisplayName, setResourceTypeDisplayName] =
    useState<string>("");
  const [processorDisplayName, setProcessorDisplayName] = useState<string>();

  function onError(error: ApolloError) {
    console.error(error);
    enqueueSnackbar("Error retrieving Processor Type", {
      variant: "error",
      key: "Error retrieving Processor Type",
    });
  }

  useGetProcessorTypeQuery({
    variables: { type: resourceConfig.type! },
    skip: !resourceConfig.isInline(),
    onError,
    onCompleted(data) {
      setResourceTypeDisplayName(data.processorType!.metadata!.displayName!);
      setProcessorDisplayName(processor.displayName ?? "");
    },
  });

  useGetProcessorWithTypeQuery({
    variables: { name: resourceConfig.name! },
    skip: resourceConfig.isInline(),
    onError,
    onCompleted(data) {
      setResourceTypeDisplayName(
        data.processorWithType!.processorType!.metadata!.displayName!
      );
      setProcessorDisplayName(trimVersion(resourceConfig.name!));
    },
  });

  return (
    <Card
      variant="outlined"
      ref={dragDropRef}
      style={{
        border: isHovered ? "1px solid #4abaeb" : undefined,
      }}
      classes={{ root: styles.card }}
    >
      <Stack
        direction="row"
        alignItems={"center"}
        spacing={1}
        justifyContent={"space-between"}
      >
        <Stack direction={"row"} spacing={1}>
          <MenuIcon className={styles["hover-icon"]} />
          <Typography fontWeight={600}>
            {resourceTypeDisplayName}
            {processorDisplayName && ":"}
          </Typography>
          {processorDisplayName && (
            <Typography>{processorDisplayName}</Typography>
          )}
        </Stack>

        <IconButton onClick={onEdit} data-testid={`edit-processor-${index}`}>
          <EditIcon width={15} height={15} style={{ float: "right" }} />
        </IconButton>
      </Stack>
    </Card>
  );
};
