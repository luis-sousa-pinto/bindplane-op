import mixins from "../../../styles/mixins.module.scss";
import { gql } from "@apollo/client";
import { Button, Stack, Typography } from "@mui/material";
import { GridRowSelectionModel } from "@mui/x-data-grid";
import { debounce } from "lodash";
import React, { useEffect, useMemo, useState } from "react";
import {
  ConfigurationChangesDocument,
  ConfigurationChangesSubscription,
  EventType,
  GetConfigurationTableQuery,
  Role,
  useConfigurationTableMetricsSubscription,
  useGetConfigurationTableQuery,
} from "../../../graphql/generated";
import { SearchBar } from "../../SearchBar";
import {
  ConfigurationsDataGrid,
  ConfigurationsTableField,
} from "./ConfigurationsDataGrid";
import { DeleteDialog } from "./DeleteDialog";
import { Link } from "react-router-dom";
import { PlusCircleIcon } from "../../Icons";
import { RBACWrapper } from "../../RBACWrapper/RBACWrapper";

gql`
  query GetConfigurationTable(
    $selector: String
    $query: String
    $onlyDeployedConfigurations: Boolean
  ) {
    configurations(
      selector: $selector
      query: $query
      onlyDeployedConfigurations: $onlyDeployedConfigurations
    ) {
      configurations {
        metadata {
          id
          version
          name
          labels
          description
        }
        agentCount
      }
      query
      suggestions {
        query
        label
      }
    }
  }

  subscription ConfigurationChanges($selector: String, $query: String) {
    configurationChanges(selector: $selector, query: $query) {
      configuration {
        metadata {
          id
          version
          name
          description
          labels
        }
        agentCount
      }
      eventType
    }
  }

  subscription ConfigurationTableMetrics($period: String!) {
    overviewMetrics(period: $period) {
      metrics {
        name
        nodeID
        pipelineType
        value
        unit
      }
    }
  }
`;

type TableConfig =
  GetConfigurationTableQuery["configurations"]["configurations"][0];

function mergeConfigs(
  currentConfigs: TableConfig[],
  configurationUpdates:
    | ConfigurationChangesSubscription["configurationChanges"]
    | undefined
): TableConfig[] {
  const newConfigs: TableConfig[] = [...currentConfigs];

  for (const update of configurationUpdates || []) {
    const config = update.configuration;
    const configIndex = currentConfigs.findIndex(
      (c) => c.metadata.name === config.metadata.name
    );
    if (update.eventType === EventType.Remove) {
      // remove the agent if it exists
      if (configIndex !== -1) {
        newConfigs.splice(configIndex, 0);
      }
    } else if (configIndex === -1) {
      newConfigs.push(config);
    } else {
      newConfigs[configIndex] = config;
    }
  }
  return newConfigs;
}

interface ConfigurationTableProps {
  selector?: string;
  initQuery?: string;
  columns?: ConfigurationsTableField[];
  setSelected: (selected: GridRowSelectionModel) => void;
  selected: GridRowSelectionModel;
  enableDelete?: boolean;
  enableNew?: boolean;
  allowSelection: boolean;
  minHeight?: string;
  overviewPage?: boolean;
}

export const ConfigurationsTable: React.FC<ConfigurationTableProps> = ({
  initQuery = "",
  selector,
  setSelected,
  selected,
  columns,
  enableDelete = true,
  enableNew = true,
  allowSelection,
  minHeight,
  overviewPage = false,
  ...dataGridProps
}) => {
  const { data, loading, refetch, subscribeToMore } =
    useGetConfigurationTableQuery({
      variables: {
        selector,
        query: initQuery,
        onlyDeployedConfigurations: overviewPage,
      },
      fetchPolicy: overviewPage ? "cache-and-network" : "network-only",
      nextFetchPolicy: "cache-only",
    });

  const { data: configurationMetrics } =
    useConfigurationTableMetricsSubscription({
      variables: { period: "1m" },
    });

  // Used to control the delete confirmation modal.
  const [open, setOpen] = useState<boolean>(false);

  const [subQuery, setSubQuery] = useState<string>(initQuery);
  const debouncedRefetch = useMemo(() => debounce(refetch, 100), [refetch]);

  useEffect(() => {
    subscribeToMore({
      document: ConfigurationChangesDocument,
      variables: { query: subQuery, selector },
      updateQuery: (prev, { subscriptionData, variables }) => {
        if (
          subscriptionData == null ||
          variables?.query !== subQuery ||
          variables.selector !== selector
        ) {
          return prev;
        }
        const { data } = subscriptionData as unknown as {
          data: ConfigurationChangesSubscription;
        };
        return {
          configurations: {
            __typename: "Configurations",
            suggestions: prev.configurations?.suggestions ?? [],
            query: prev.configurations?.query ?? "",
            configurations: mergeConfigs(
              prev.configurations?.configurations ?? [],
              data.configurationChanges
            ),
          },
        };
      },
    });
  }, [selector, subQuery, subscribeToMore]);

  function onQueryChange(query: string) {
    debouncedRefetch({ selector, query });
    setSubQuery(query);
  }

  function openModal() {
    setOpen(true);
  }

  function closeModal() {
    setOpen(false);
  }

  return (
    <>
      <Stack
        direction="row"
        alignItems="center"
        justifyContent="space-between"
        marginBottom={3}
      >
        <Typography variant="h5">Configs</Typography>
        {selected.length > 0 && enableDelete && (
          <RBACWrapper requiredRole={Role.User}>
            <Button variant="contained" color="error" onClick={openModal}>
              Delete {selected.length} Config
              {selected.length > 1 && "s"}
            </Button>
          </RBACWrapper>
        )}

        {selected.length === 0 && enableNew && (
          <RBACWrapper requiredRole={Role.User}>
            <Button
              component={Link}
              to="/configurations/new"
              variant="contained"
              classes={{ root: mixins["float-right"] }}
              startIcon={<PlusCircleIcon />}
            >
              Create Config
            </Button>
          </RBACWrapper>
        )}
      </Stack>

      <Stack spacing={1}>
        <SearchBar
          suggestions={data?.configurations.suggestions}
          onQueryChange={onQueryChange}
          suggestionQuery={data?.configurations.query}
          initialQuery={initQuery}
        />

        <ConfigurationsDataGrid
          {...dataGridProps}
          allowSelection={allowSelection}
          setSelectionModel={setSelected}
          loading={loading}
          configurations={data?.configurations.configurations ?? []}
          configurationMetrics={configurationMetrics}
          columnFields={columns}
          selectionModel={selected}
          minHeight={minHeight}
          classes={
            overviewPage &&
            (data?.configurations.configurations.length ?? 0) < 100
              ? { footerContainer: mixins["hidden"] }
              : undefined
          }
        />
      </Stack>
      <DeleteDialog
        onClose={closeModal}
        selected={selected}
        open={open}
        onDeleteSuccess={refetch}
      />
    </>
  );
};
