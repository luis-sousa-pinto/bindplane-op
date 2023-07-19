import { gql } from "@apollo/client";
import { Typography, Button, Stack } from "@mui/material";
import { DataGridProps, GridRowSelectionModel } from "@mui/x-data-grid";
import { useSnackbar } from "notistack";
import { useState, useEffect } from "react";
import { debounce } from "lodash";

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
import { Role, useDestinationsQuery } from "../../graphql/generated";
import { ResourceStatus, ResourceKind } from "../../types/resources";
import {
  deleteResources,
  MinimumDeleteResource,
} from "../../utils/rest/delete-resources";
import { useRole } from "../../hooks/useRole";
import { hasPermission } from "../../utils/has-permission";
import { RBACWrapper } from "../../components/RBACWrapper/RBACWrapper";
import { useLocation } from "react-router-dom";
import { SearchBar } from "../../components/SearchBar";

import mixins from "../../styles/mixins.module.scss";

gql`
  query Destinations($query: String, $filterUnused: Boolean) {
    destinations(query: $query, filterUnused: $filterUnused) {
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

export interface DestinationsPageContentProps
  extends Omit<DataGridProps, "rows" | "columns"> {
  destinationsPage: boolean;
  // grid selection model
  selected: GridRowSelectionModel;
  // function to set grid selection model
  setSelected: (selected: GridRowSelectionModel) => void;
  columnFields?: DestinationsTableField[];
  minHeight?: string;
  maxHeight?: string;
  editingDestination: string | null;
  setEditingDestination: (dest: string | null) => void;
  allowSelection: boolean;
}

export const DestinationsPageSubContent: React.FC<
  DestinationsPageContentProps
> = ({
  destinationsPage,
  selected,
  setSelected,
  columnFields,
  minHeight,
  maxHeight,
  editingDestination,
  setEditingDestination,
  allowSelection,
  ...dataGridProps
}) => {
  // Used to control the delete confirmation modal.
  const [open, setOpen] = useState<boolean>(false);

  const [failedDeletes, setFailedDeletes] = useState<ResourceStatus[]>([]);
  const [failedDeletesOpen, setFailedDeletesOpen] = useState(false);

  const { enqueueSnackbar } = useSnackbar();

  const { data, refetch, error } = useDestinationsQuery({
    variables: {
      filterUnused: !destinationsPage,
    },
    fetchPolicy: "cache-and-network",
    refetchWritePolicy: "merge",
  });

  const debouncedRefetch = debounce((query: string) => {
    refetch({
      filterUnused: !destinationsPage,
      query: query,
    });
  }, 100);

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

  return (
    <>
      <Stack
        direction="row"
        justifyContent="space-between"
        alignItems="center"
        height="48px"
        marginBottom={2}
      >
        <Typography variant="h5">Destinations</Typography>
        {destinationsPage && selected.length > 0 && (
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
        )}
      </Stack>
      <Stack spacing={1}>
        <SearchBar
          suggestions={[]}
          onQueryChange={debouncedRefetch}
          suggestionQuery={""}
          initialQuery={""}
          placeholder={"Filter by destination name"}
        />
        <DestinationsDataGrid
          {...dataGridProps}
          loading={data == null}
          setSelectionModel={setSelected}
          selectionModel={selected}
          disableRowSelectionOnClick
          checkboxSelection
          onEditDestination={(name: string) => setEditingDestination(name)}
          columnFields={columnFields}
          minHeight={minHeight}
          maxHeight={maxHeight}
          rows={data?.destinations ?? []}
          allowSelection={allowSelection}
          classes={
            !destinationsPage && (data?.destinations ?? []).length < 100
              ? { footerContainer: mixins["hidden"] }
              : {}
          }
        />
      </Stack>
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
  const location = useLocation();
  const role = useRole();

  const isDestinationsPage = location.pathname.includes("destinations");

  return (
    <CardContainer>
      <DestinationsPageSubContent
        allowSelection={hasPermission(Role.Admin, role)}
        destinationsPage={isDestinationsPage}
        selected={selected}
        setSelected={setSelected}
        editingDestination={editingDestination}
        setEditingDestination={setEditingDestination}
        maxHeight="70vh"
        minHeight="70vh"
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
