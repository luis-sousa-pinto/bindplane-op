import { Button, Stack, Typography } from "@mui/material";
import { GridRowSelectionModel } from "@mui/x-data-grid";
import React, { useRef, useState } from "react";
import { Link } from "react-router-dom";
import { CardContainer } from "../../components/CardContainer";
import { PlusCircleIcon } from "../../components/Icons";
import { classes } from "../../utils/styles";
import { deleteAgents } from "../../utils/rest/delete-agents";
import { useSnackbar } from "notistack";
import { ConfirmDeleteResourceDialog } from "../../components/ConfirmDeleteResourceDialog";
import { withRequireLogin } from "../../contexts/RequireLogin";
import { withNavBar } from "../../components/NavBar";
import { isFunction } from "lodash";
import { upgradeAgents } from "../../utils/rest/upgrade-agent";
import { Role } from "../../graphql/generated";
import { hasPermission } from "../../utils/has-permission";
import { useRole } from "../../hooks/useRole";
import { useComponents } from "../../hooks/useComponents";
import { RBACWrapper } from "../../components/RBACWrapper/RBACWrapper";

import mixins from "../../styles/mixins.module.scss";

export const AgentsPageContent: React.FC = () => {
  const [updatable, setUpdatable] = useState<GridRowSelectionModel>([]);
  const [deletable, setDeletable] = useState<GridRowSelectionModel>([]);
  const [deleteConfirmOpen, setDeleteConfirmOpen] = useState(false);

  const clearSelectionModelFnRef = useRef<(() => void) | null>(null);

  const { enqueueSnackbar } = useSnackbar();
  const role = useRole();
  const { AgentsTable } = useComponents();

  function handleSelectUpdatable(agentIds: GridRowSelectionModel) {
    setUpdatable(agentIds);
  }
  function handleSelectDeletable(agentIds: GridRowSelectionModel) {
    setDeletable(agentIds);
  }

  async function handleDeleteAgents() {
    try {
      await deleteAgents(deletable as string[]);
      setDeletable([]);
      setDeleteConfirmOpen(false);
    } catch (err) {
      console.error(err);
      enqueueSnackbar("Failed to delete agents.", { variant: "error" });
    }
  }

  async function handleUpgradeAgents() {
    try {
      const errors = await upgradeAgents(updatable as string[]);

      if (isFunction(clearSelectionModelFnRef.current)) {
        clearSelectionModelFnRef.current();
      }

      setUpdatable([]);

      if (errors.length > 0) {
        console.error("Upgrade errors.", { errors });
      }
    } catch (err) {
      enqueueSnackbar("Failed to send upgrade request.", {
        variant: "error",
        key: "Failed to send upgrade request.",
      });
    }
  }

  return (
    <>
      {/* --------------------- Delete Button and Confirmation --------------------- */}
      <ConfirmDeleteResourceDialog
        onDelete={handleDeleteAgents}
        onCancel={() => setDeleteConfirmOpen(false)}
        action={"delete"}
        open={deleteConfirmOpen}
        title={`Delete ${deletable.length} Disconnected Agent${
          deletable.length > 1 ? "s" : ""
        }?`}
      >
        <>
          <Typography>
            Agents will reappear in BindPlane OP if reconnected.
          </Typography>
        </>
      </ConfirmDeleteResourceDialog>
      <CardContainer>
        {deletable.length > 0 ? (
          <Button
            variant="contained"
            color="error"
            classes={{ root: mixins["float-right"] }}
            onClick={() => setDeleteConfirmOpen(true)}
          >
            Delete {deletable.length} Disconnected Agent
            {deletable.length > 1 && "s"}
          </Button>
        ) : (
          <RBACWrapper requiredRole={Role.User}>
            <Button
              component={Link}
              variant={"contained"}
              classes={{ root: mixins["float-right"] }}
              to="/agents/install"
              startIcon={<PlusCircleIcon />}
            >
              Install Agent
            </Button>
          </RBACWrapper>
        )}

        {/* --------------------- Update Button and Confirmation ---------------------  */}

        {updatable.length > 0 && (
          <Button
            variant="outlined"
            color="primary"
            classes={{ root: classes([mixins["float-right"], mixins["mr-3"]]) }}
            onClick={handleUpgradeAgents}
          >
            Upgrade {updatable.length} Outdated Agent
            {updatable.length > 1 && "s"}
          </Button>
        )}
        <Stack
          direction="row"
          justifyContent="space-between"
          alignItems="center"
          height="48px"
          marginBottom={3}
        >
          <Typography variant="h5">Agents</Typography>
        </Stack>
        <AgentsTable
          allowSelection={hasPermission(Role.User, role)}
          onDeletableAgentsSelected={handleSelectDeletable}
          onUpdatableAgentsSelected={handleSelectUpdatable}
          clearSelectionModelFnRef={clearSelectionModelFnRef}
        />
      </CardContainer>
    </>
  );
};

export const AgentsPage = withRequireLogin(withNavBar(AgentsPageContent));
