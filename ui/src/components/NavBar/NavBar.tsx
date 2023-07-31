import { AppBar, IconButton, Toolbar } from "@mui/material";
import React, { SyntheticEvent, useEffect, useState } from "react";
import { Link, NavLink, useLocation, useNavigate } from "react-router-dom";
import {
  AgentGridIcon,
  EmailIcon,
  GridIcon,
  HelpCircleIcon,
  SettingsIcon,
  SlackIcon,
  SlidersIcon,
  SquareIcon,
} from "../Icons";
import { BindPlaneOPLogo } from "../Logos";
import { classes } from "../../utils/styles";
import { useComponents } from "../../hooks/useComponents";

import styles from "./nav-bar.module.scss";

export const NavBar: React.FC = () => {
  const [settingsAnchorEl, setAnchorEl] = useState<Element | null>(null);
  const settingsOpen = Boolean(settingsAnchorEl);
  const { SettingsMenu } = useComponents();

  const navigate = useNavigate();

  // make navigate available to the global window
  // to let us use it outside of components.
  useEffect(() => {
    window.navigate = navigate;
  }, [navigate]);

  const handleSettingsClick = (event: SyntheticEvent) => {
    setAnchorEl(event.currentTarget);
  };

  const handleSettingsClose = () => {
    setAnchorEl(null);
  };

  const location = useLocation();
  return (
    <>
      <AppBar position="static" classes={{ root: styles["app-bar-root"] }}>
        <Toolbar classes={{ root: styles.toolbar }}>
          <Link to="/">
            <BindPlaneOPLogo
              className={styles.logo}
              aria-label="bindplane-logo"
            />
          </Link>

          <div className={styles["main-nav"]}>
            <NavLink
              className={({ isActive }) =>
                isActive
                  ? classes([styles["nav-link"], styles["active"]])
                  : styles["nav-link"]
              }
              to={{ pathname: "/overview", search: location.search }}
            >
              {({ isActive }) => {
                const className = isActive
                  ? classes([styles["icon"], styles["active"]])
                  : styles["icon"];
                return (
                  <>
                    <GridIcon className={className} />
                    Overview
                  </>
                );
              }}
            </NavLink>
            <NavLink
              className={({ isActive }) =>
                isActive
                  ? classes([styles["nav-link"], styles["active"]])
                  : styles["nav-link"]
              }
              to={{ pathname: "/agents", search: location.search }}
            >
              {({ isActive }) => {
                const className = isActive
                  ? classes([styles["icon"], styles["active"]])
                  : styles["icon"];
                return (
                  <>
                    <AgentGridIcon className={className} />
                    Agents
                  </>
                );
              }}
            </NavLink>

            <NavLink
              className={({ isActive }) =>
                isActive
                  ? classes([styles["nav-link"], styles["active"]])
                  : styles["nav-link"]
              }
              to={{ pathname: "/configurations", search: location.search }}
            >
              {({ isActive }) => {
                const className = isActive
                  ? classes([styles["icon"], styles["active"]])
                  : styles["icon"];
                return (
                  <>
                    <SlidersIcon className={className} />
                    Configurations
                  </>
                );
              }}
            </NavLink>

            <NavLink
              className={({ isActive }) =>
                isActive
                  ? classes([styles["nav-link"], styles["active"]])
                  : styles["nav-link"]
              }
              to={{ pathname: "/destinations", search: location.search }}
            >
              {({ isActive }) => {
                const className = isActive
                  ? classes([
                      styles["icon"],
                      styles["destination-icon"],
                      styles["active"],
                    ])
                  : classes([styles["icon"], styles["destination-icon"]]);
                return (
                  <>
                    <SquareIcon className={className} />
                    Destinations
                  </>
                );
              }}
            </NavLink>
          </div>

          <div className={styles["sub-nav"]}>
            <IconButton
              className={styles.button}
              target="_blank"
              color="inherit"
              data-testid="doc-link"
              href="https://docs.bindplane.observiq.com/docs"
            >
              <HelpCircleIcon className={styles.icon} />
            </IconButton>
            <IconButton
              className={styles.button}
              target="_blank"
              color="inherit"
              data-testid="support-link"
              href="mailto:support.bindplaneop@observiq.com"
            >
              <EmailIcon className={styles.icon} />
            </IconButton>
            <IconButton
              className={styles.button}
              target="_blank"
              color="inherit"
              data-testid="slack-link"
              href="https://observiq.com/support-bindplaneop/"
            >
              <SlackIcon className={styles.icon} />
            </IconButton>
            <IconButton
              className={styles.button}
              aria-controls={settingsOpen ? "settings-menu" : undefined}
              aria-haspopup="true"
              aria-expanded={settingsOpen ? "true" : undefined}
              color="inherit"
              data-testid="settings-button"
              onClick={handleSettingsClick}
            >
              <SettingsIcon className={styles.icon} />
            </IconButton>
            <SettingsMenu
              anchorEl={settingsAnchorEl}
              onClose={handleSettingsClose}
              open={settingsOpen}
            />
          </div>
        </Toolbar>
      </AppBar>
    </>
  );
};

export function withNavBar(FC: React.FC): React.FC {
  return () => (
    <>
      <NavBar />
      <div className="content">
        <FC />
      </div>
    </>
  );
}
