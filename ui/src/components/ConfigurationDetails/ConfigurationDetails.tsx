import {
  Box,
  Card,
  CardContent,
  CardHeader,
  CircularProgress,
  ClickAwayListener,
  Divider,
  IconButton,
  Stack,
  TextField,
  Typography,
} from "@mui/material";
import { useRef, useState } from "react";
import { PencilIcon } from "../Icons";
import { ApolloError, gql } from "@apollo/client";
import { useSnackbar } from "notistack";
import {
  useEditConfigDescriptionMutation,
  useGetCurrentConfigVersionQuery,
  useGetLatestConfigDescriptionQuery,
} from "../../graphql/generated";
import { asCurrentVersion, asLatestVersion } from "../../utils/version-helpers";
import { useRefetchOnConfigurationChange } from "../../hooks/useRefetchOnConfigurationChanges";
import { ConfigDetailsMenu } from "../ConfigDetailsMenu/ConfigDetailsMenu";

import styles from "./configuration-details.module.scss";

gql`
  query getCurrentConfigVersion($configurationName: String!) {
    configuration(name: $configurationName) {
      metadata {
        id
        name
        version
        labels
      }
      agentCount
    }
  }

  query getLatestConfigDescription($configurationName: String!) {
    configuration(name: $configurationName) {
      metadata {
        id
        name
        version
        description
      }
    }
  }

  mutation editConfigDescription($input: EditConfigurationDescriptionInput!) {
    editConfigurationDescription(input: $input)
  }
`;

interface ConfigurationDetailsProps {
  configurationName: string;
  disableEdit?: boolean;
}

/**
 * ConfigurationDetails shows some details about the configuration and allows
 * a user to edit the description.
 *
 * @param configurationName should be the non-versioned name of the configuration
 * @param disableDescriptionEdit if true, the description will not be editable
 */
export const ConfigurationDetails: React.FC<ConfigurationDetailsProps> = ({
  configurationName,
  disableEdit,
}) => {
  const { enqueueSnackbar } = useSnackbar();
  function onError(error: ApolloError) {
    enqueueSnackbar(error.message, { variant: "error" });
  }

  const { data: currentVersionData, refetch: refetchCurrent } =
    useGetCurrentConfigVersionQuery({
      variables: {
        configurationName: asCurrentVersion(configurationName),
      },
      onError,
      fetchPolicy: "cache-and-network",
    });
  const {
    data: latestVersionData,
    refetch: refetchLatest,
    loading: loadingLatest,
  } = useGetLatestConfigDescriptionQuery({
    variables: {
      configurationName: asLatestVersion(configurationName),
    },
    onError,
    fetchPolicy: "cache-and-network",
  });

  useRefetchOnConfigurationChange(configurationName, () => {
    refetchCurrent();
    refetchLatest();
  });

  const [editConfigDescription, { loading: editLoading }] =
    useEditConfigDescriptionMutation();

  async function handleEditDescriptionSave(description: string) {
    await editConfigDescription({
      variables: {
        input: {
          name: configurationName,
          description,
        },
      },
    });

    await refetchLatest();
  }

  const details: DetailProps[] = [
    {
      label: "Current Version",
      value: currentVersionData?.configuration?.metadata.version,
      loading: !currentVersionData,
    },
    {
      label: "Platform",
      value: currentVersionData?.configuration?.metadata.labels.platform,
      loading: !currentVersionData,
    },
    {
      label: "Number of Agents",
      value: currentVersionData?.configuration?.agentCount ?? "",
      loading: !currentVersionData,
    },
    {
      label: "Description",
      value: latestVersionData?.configuration?.metadata.description ?? "",
      onChange: disableEdit ? undefined : handleEditDescriptionSave,
      loading: !latestVersionData || editLoading || loadingLatest,
      flexGrow: 4,
    },
  ];

  return (
    <Card classes={{ root: styles.card }}>
      <CardHeader
        action={
          disableEdit ? null : (
            <ConfigDetailsMenu configName={configurationName} />
          )
        }
        title={configurationName}
        titleTypographyProps={{ fontWeight: 600 }}
        classes={{ root: styles.padding }}
      />
      <Divider />
      <CardContent className={styles.padding}>
        <Stack direction="row" width="100%">
          {details.map((detail) => (
            <Detail key={`config-detail-${detail.label}`} {...detail} />
          ))}
        </Stack>
      </CardContent>
    </Card>
  );
};

interface DetailProps {
  label: string;
  value?: string | number;
  // onChange will render an edit icon and allow the user to edit the value
  // Only intended for use with the description field.
  onChange?: (value: string) => Promise<void>;
  loading?: boolean;
  flexGrow?: number;
}

const Detail: React.FC<DetailProps> = ({
  label,
  value,
  loading,
  onChange,
  flexGrow = 1,
}) => {
  const [editing, setEditing] = useState(false);
  const textboxRef = useRef<HTMLInputElement | null>(null);

  function handleEdit() {
    setEditing(true);
  }

  async function handleClickAway() {
    const newDescription = textboxRef.current?.value;
    if (newDescription != null) {
      await onChange?.(newDescription);
    }
    setEditing(false);
  }

  return (
    <Stack flexGrow={flexGrow} maxWidth={500} minWidth={200}>
      <Stack direction="row" alignItems="center" height="24px" spacing={1}>
        <Typography fontWeight={600}>{label}</Typography>{" "}
        {onChange && !editing && (
          <IconButton
            size="small"
            onClick={handleEdit}
            data-testid="edit-description-button"
          >
            <PencilIcon width={14} />
          </IconButton>
        )}
      </Stack>
      {onChange && editing ? (
        <ClickAwayListener onClickAway={handleClickAway}>
          {loading ? (
            <Stack
              width={"100%"}
              height={"100%"}
              alignItems="center"
              justifyContent="center"
            >
              <CircularProgress size={20} disableShrink />
            </Stack>
          ) : (
            <TextField
              multiline
              inputRef={textboxRef}
              defaultValue={value}
              fullWidth
            />
          )}
        </ClickAwayListener>
      ) : loading ? (
        <Box marginTop={"4px"} marginLeft={"4px"}>
          <CircularProgress size={14} disableShrink />
        </Box>
      ) : (
        <Typography className={styles.value}>{value}</Typography>
      )}
    </Stack>
  );
};
