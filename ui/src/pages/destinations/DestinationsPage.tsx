import { gql, QueryHookOptions, QueryResult } from "@apollo/client";
import { Typography, FormControl, Button } from "@mui/material";
import { GridRowSelectionModel } from "@mui/x-data-grid";
import { useSnackbar } from "notistack";
import { useState, useEffect } from "react";
import { CardContainer } from "../../components/CardContainer";
import { ConfirmDeleteResourceDialog } from "../../components/ConfirmDeleteResourceDialog";
import { withNavBar } from "../../components/NavBar";
import {
  DestinationsDataGrid,
  DestinationsTableField,
} from "../../components/Tables/DestinationsTable/DestinationsDataGrid";
import { EditDestinationDialog } from "../../components/Tables/DestinationsTable/EditDestinationDialog";
import { FailedDeleteDialog } from "../../components/Tables/DestinationsTable/FailedDeleteDialog";
import { withRequireLogin } from "../../contexts/RequireLogin";
import {
  DestinationsInConfigsQuery,
  DestinationsInConfigsQueryVariables,
  DestinationsQuery,
  DestinationsQueryVariables,
  Exact,
  Role,
  useDestinationsQuery,
} from "../../graphql/generated";
import { ResourceStatus, ResourceKind } from "../../types/resources";
import {
  deleteResources,
  MinimumDeleteResource,
} from "../../utils/rest/delete-resources";
import { useRole } from "../../hooks/useRole";
import { hasPermission } from "../../utils/has-permission";

import mixins from "../../styles/mixins.module.scss";
import { RBACWrapper } from "../../components/RBACWrapper/RBACWrapper";

gql`
  query Destinations {
    destinations {
      kind
      metadata {
        id
        name
        version
      }
      spec {
        type
      }
    }
  }
`;

export interface DestinationsPageContentProps {
  destinationsPage: boolean;
  // grid selection model
  selected: GridRowSelectionModel;
  // function to set grid selection model
  setSelected: (selected: GridRowSelectionModel) => void;
  columnFields?: DestinationsTableField[];
  minHeight?: string;
  editingDestination: string | null;
  setEditingDestination: (dest: string | null) => void;

  allowSelection: boolean;

  // as function for the graphql query
  destinationsQuery:
    | ((
        baseOptions?: QueryHookOptions<
          DestinationsQuery,
          DestinationsQueryVariables
        >
      ) => QueryResult<
        DestinationsQuery,
        Exact<{
          [key: string]: never;
        }>
      >)
    | ((
        baseOptions?: QueryHookOptions<
          DestinationsInConfigsQuery,
          DestinationsInConfigsQueryVariables
        >
      ) => QueryResult<
        DestinationsInConfigsQuery,
        Exact<{ [key: string]: never }>
      >);
}
export const DestinationsPageSubContent: React.FC<
  DestinationsPageContentProps
> = ({
  destinationsPage,
  selected,
  setSelected,
  columnFields,
  destinationsQuery,
  minHeight,
  editingDestination,
  setEditingDestination,
  allowSelection,
}) => {
  // Used to control the delete confirmation modal.
  const [open, setOpen] = useState<boolean>(false);

  const [failedDeletes, setFailedDeletes] = useState<ResourceStatus[]>([]);
  const [failedDeletesOpen, setFailedDeletesOpen] = useState(false);

  const { enqueueSnackbar } = useSnackbar();

  const { data, loading, refetch, error } = destinationsQuery({
    fetchPolicy: "cache-and-network",
  });

  useEffect(() => {
    if (error != null) {
      enqueueSnackbar("There was an error retrieving data.", {
        variant: "error",
      });
    }
  }, [enqueueSnackbar, error]);

  useEffect(() => {
    if (failedDeletes.length > 0) {
      setFailedDeletesOpen(true);
    }
  }, [failedDeletes, setFailedDeletesOpen]);

  function onAcknowledge() {
    setFailedDeletesOpen(false);
  }

  function handleEditSaveSuccess() {
    refetch();
    setEditingDestination(null);
  }

  async function deleteDestinations() {
    try {
      const items = resourcesFromSelected(selected);
      const { updates } = await deleteResources(items);
      setOpen(false);

      const failures = updates.filter((u) => u.status !== "deleted");
      setFailedDeletes(failures);

      refetch();
    } catch (err) {
      console.error(err);
      enqueueSnackbar("Failed to delete destinations.", { variant: "error" });
    }
  }
  const queryData = data ?? { destinations: [] };
  const rows =
    "destinations" in queryData
      ? [...queryData.destinations]
      : [...queryData.destinationsInConfigs];

  return (
    <>
      <div className={mixins.flex}>
        <Typography variant="h5" className={mixins["mb-5"]}>
          Destinations
        </Typography>
        {destinationsPage && selected.length > 0 && (
          <FormControl classes={{ root: mixins["ml-5"] }}>
            <RBACWrapper requiredRole={Role.User}>
              <Button
                variant="contained"
                color="error"
                onClick={() => setOpen(true)}
              >
                Delete {selected.length} Destination
                {selected.length > 1 && "s"}
              </Button>
            </RBACWrapper>
          </FormControl>
        )}
      </div>
      <DestinationsDataGrid
        loading={loading}
        setSelectionModel={setSelected}
        selectionModel={selected}
        disableRowSelectionOnClick
        checkboxSelection
        onEditDestination={(name: string) => setEditingDestination(name)}
        columnFields={columnFields}
        minHeight={minHeight}
        rows={rows}
        allowSelection={allowSelection}
        classes={
          !destinationsPage && rows.length < 100
            ? { footerContainer: mixins["hidden"] }
            : {}
        }
      />
      <ConfirmDeleteResourceDialog
        open={open}
        onClose={() => setOpen(false)}
        onDelete={deleteDestinations}
        onCancel={() => setOpen(false)}
        action={"delete"}
      >
        <Typography>
          Are you sure you want to delete {selected.length} destination
          {selected.length > 1 && "s"}?
        </Typography>
      </ConfirmDeleteResourceDialog>

      <FailedDeleteDialog
        open={failedDeletesOpen}
        failures={failedDeletes}
        onAcknowledge={onAcknowledge}
        onClose={() => setFailedDeletesOpen(false)}
      />

      {editingDestination && (
        <EditDestinationDialog
          name={editingDestination}
          onCancel={() => setEditingDestination(null)}
          onSaveSuccess={handleEditSaveSuccess}
        />
      )}
    </>
  );
};

export const DestinationsPageContent: React.FC = () => {
  const [selected, setSelected] = useState<GridRowSelectionModel>([]);
  const [editingDestination, setEditingDestination] = useState<string | null>(
    null
  );

  const role = useRole();

  return (
    <CardContainer>
      <DestinationsPageSubContent
        allowSelection={hasPermission(Role.Admin, role)}
        destinationsPage={false}
        selected={selected}
        setSelected={setSelected}
        editingDestination={editingDestination}
        setEditingDestination={setEditingDestination}
        destinationsQuery={useDestinationsQuery}
        minHeight="300px"
      />
    </CardContainer>
  );
};

export const DestinationsPage = withRequireLogin(
  withNavBar(DestinationsPageContent)
);

export function resourcesFromSelected(
  selected: GridRowSelectionModel
): MinimumDeleteResource[] {
  return selected.reduce<MinimumDeleteResource[]>((prev, cur) => {
    if (typeof cur !== "string") {
      console.error(`Unexpected type for GridRowId: ${typeof cur}"`);
      return prev;
    }
    const [kind, name] = cur.split("|");

    if (kind == null || name == null) {
      console.error(`Malformed grid row ID: ${cur}`);
      return prev;
    }

    prev.push({ kind: ResourceKind.DESTINATION, metadata: { name } });
    return prev;
  }, []);
}
