import {
  IconButton,
  ListItemIcon,
  ListItemText,
  Menu,
  MenuItem,
  MenuList,
  Typography,
  colors,
} from "@mui/material";
import { CopyIcon, MenuIcon, SettingsIcon, TrashIcon } from "../Icons";
import { useState } from "react";
import { DuplicateConfigDialog } from "../../pages/configurations/configuration/DuplicateConfigDialog";
import { ConfirmDeleteResourceDialog } from "../ConfirmDeleteResourceDialog";
import { deleteResources } from "../../utils/rest/delete-resources";
import { ResourceKind } from "../../types/resources";
import { useSnackbar } from "notistack";
import { useNavigate } from "react-router-dom";
import { asCurrentVersion } from "../../utils/version-helpers";
import { AdvancedConfigDialog } from "../../pages/configurations/configuration/AdvancedConfigDialog";

interface ConfigDetailsMenuProps {
  configName: string;
}

export const ConfigDetailsMenu: React.FC<ConfigDetailsMenuProps> = ({
  configName,
}) => {
  const [anchorEl, setAnchorEl] = useState<null | HTMLElement>(null);
  const menuOpen = Boolean(anchorEl);
  const [duplicateDialogOpen, setDuplicateDialogOpen] = useState(false);
  const [deleteDialogOpen, setDeleteDialogOpen] = useState(false);
  const [advancedDialogOpen, setAdvancedDialogOpen] = useState(false);

  const { enqueueSnackbar } = useSnackbar();
  const navigate = useNavigate();

  async function handleClick(e: React.MouseEvent<HTMLButtonElement>) {
    setAnchorEl(e.currentTarget);
  }

  function handleClose() {
    setAnchorEl(null);
  }

  async function handleDelete() {
    try {
      await deleteResources([
        {
          kind: ResourceKind.CONFIGURATION,
          metadata: {
            name: configName,
          },
        },
      ]);
      enqueueSnackbar(`Deleted configuration ${configName}`, {
        variant: "success",
      });
      setDeleteDialogOpen(false);
      navigate("/configurations");
    } catch (err) {
      console.error(err);
      enqueueSnackbar("Failed to delete configuration", { variant: "error" });
    }
  }

  return (
    <>
      <IconButton
        data-testid="config-menu-button"
        id="config-menu-button"
        aria-controls={menuOpen ? "config-menu" : undefined}
        aria-haspopup="true"
        aria-expanded={menuOpen ? "true" : undefined}
        onClick={handleClick}
      >
        <MenuIcon />
      </IconButton>
      <Menu
        id="config-menu"
        anchorEl={anchorEl}
        open={menuOpen}
        onClose={handleClose}
        MenuListProps={{
          "aria-labelledby": "config-menu-button",
        }}
        anchorOrigin={{
          vertical: "bottom",
          horizontal: "left",
        }}
        transformOrigin={{
          vertical: "top",
          horizontal: "center",
        }}
      >
        <MenuList>
          <MenuItem onClick={() => setDuplicateDialogOpen(true)}>
            <ListItemIcon>
              <CopyIcon width="20px" />
            </ListItemIcon>
            <ListItemText>Duplicate current version</ListItemText>
          </MenuItem>
          <MenuItem onClick={() => setAdvancedDialogOpen(true)}>
            <ListItemIcon>
              <SettingsIcon width="20px" />
            </ListItemIcon>
            <ListItemText>Advanced Configuration</ListItemText>
          </MenuItem>
          <MenuItem onClick={() => setDeleteDialogOpen(true)}>
            <ListItemIcon>
              <TrashIcon width="20px" stroke={colors.red[700]} />
            </ListItemIcon>
            <Typography color="error">Delete</Typography>
          </MenuItem>
        </MenuList>
      </Menu>

      <DuplicateConfigDialog
        currentConfigName={asCurrentVersion(configName)}
        open={duplicateDialogOpen}
        onClose={() => setDuplicateDialogOpen(false)}
        onSuccess={() => {
          setDuplicateDialogOpen(false);
          setAnchorEl(null);
        }}
      />

      <AdvancedConfigDialog
        configName={configName}
        open={advancedDialogOpen}
        onClose={() => setAdvancedDialogOpen(false)}
        onSuccess={() => {
          setAdvancedDialogOpen(false);
          setAnchorEl(null);
        }}
      />

      <ConfirmDeleteResourceDialog
        onDelete={handleDelete}
        onCancel={() => setDeleteDialogOpen(false)}
        action={"delete"}
        open={deleteDialogOpen}
      >
        <Typography>
          Are you sure you want to delete this configuration?
        </Typography>
      </ConfirmDeleteResourceDialog>
    </>
  );
};
