import { Menu, MenuItem, MenuProps } from "@mui/material";
import { LogoutIcon } from "../Icons";
import { useNavigate } from "react-router-dom";

import styles from "./settings-menu.module.scss";

export type SettingsMenuProps = Pick<
  MenuProps,
  "anchorEl" | "open" | "onClose"
>;

export const SettingsMenu: React.FC<SettingsMenuProps> = ({
  anchorEl,
  open,
  onClose,
}) => {
  const navigate = useNavigate();

  async function handleLogout() {
    await fetch("/logout", {
      method: "PUT",
    });

    localStorage.removeItem("user");
    navigate("/login");
  }

  return (
    <Menu
      anchorEl={anchorEl}
      open={open}
      onClose={onClose}
      anchorOrigin={{
        vertical: "bottom",
        horizontal: "center",
      }}
      transformOrigin={{
        vertical: "top",
        horizontal: "right",
      }}
      MenuListProps={{
        "aria-labelledby": "settings-button   ",
      }}
    >
      <MenuItem onClick={handleLogout}>
        <LogoutIcon className={styles["settings-icon"]} />
        Logout
      </MenuItem>
    </Menu>
  );
};
