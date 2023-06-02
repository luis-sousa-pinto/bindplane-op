import { Box, Button, Collapse, Grid, Stack, Typography } from "@mui/material";
import React, { useState } from "react";
import { GetAgentAndConfigurationsQuery } from "../../../graphql/generated";
import { AgentStatus } from "../../../types/agents";
import { renderAgentDate, renderAgentLabels } from "../utils";
import { ArrowUpIcon } from "../../Icons";
import { upgradeAgent } from "../../../utils/rest/upgrade-agent";
import { isEmpty } from "lodash";
import { ExpandButton } from "../../ExpandButton";

import styles from "./agent-table.module.scss";
import mixins from "../../../styles/mixins.module.scss";

type AgentTableAgent = NonNullable<GetAgentAndConfigurationsQuery["agent"]>;
interface AgentTableProps {
  agent: AgentTableAgent;
}

export const AgentTable: React.FC<AgentTableProps> = ({ agent }) => {
  const [expanded, setExpanded] = useState(false);
  function renderTable(agent: AgentTableAgent): JSX.Element {
    const { status, labels, connectedAt, disconnectedAt } = agent;

    const labelsEl = renderAgentLabels(labels);

    function renderConnectedAtRow(): JSX.Element {
      if (status === AgentStatus.CONNECTED) {
        const connectedEl = renderAgentDate(connectedAt);
        return renderRow("Connected", connectedEl);
      }

      const disconnectedEl = renderAgentDate(disconnectedAt);
      return renderRow("Disconnected", disconnectedEl);
    }

    function fillIn(s: any): string {
      return isEmpty(s) ? "-" : s;
    }

    return (
      <>
        <Grid container columnSpacing={5} rowSpacing={5} alignItems="center">
          {renderVersionRow("Version", agent)}
          {renderConnectedAtRow()}
          {renderRow("Platform", <>{fillIn(agent.platform)}</>)}
          {renderRow("Agent ID", <>{fillIn(agent.id)}</>)}
          {renderLabelsRow("Labels", labelsEl)}
          <Grid item></Grid>
        </Grid>
        <Collapse in={expanded} collapsedSize={0}>
          <Grid container columnSpacing={5} rowSpacing={5} alignItems="center">
            {renderRow("Host Name", <>{fillIn(agent.hostName)}</>)}
            {renderRow("Remote Address", <>{fillIn(agent.remoteAddress)}</>)}
            {renderRow("MAC Address", <>{fillIn(agent.macAddress)}</>)}
            {renderRow(
              "Operating System",
              <>{fillIn(agent.operatingSystem)}</>
            )}
            {renderRow("Architecture", <>{fillIn(agent.architecture)}</>)}
            {renderRow("Home", <>{fillIn(agent.home)}</>)}
          </Grid>
        </Collapse>
        <Box
          sx={{
            borderTop: 1,
            paddingTop: 3,
            ml: -3,
            mr: -3,
            borderColor: "divider",
          }}
        >
          <Box sx={{ ml: 3, mr: 3 }}>
            <Grid
              container
              spacing={10}
              direction="column"
              alignItems="center"
              justifyContent="center"
            >
              <Grid item flex={1} xs={12} sm={12}>
                <ExpandButton
                  expanded={expanded}
                  onToggleExpanded={() => setExpanded((prev) => !prev)}
                />
              </Grid>
            </Grid>
          </Box>
        </Box>
      </>
    );
  }
  return <>{agent == null ? null : renderTable(agent)}</>;
};

function renderVersionRow(key: string, agent: AgentTableAgent): JSX.Element {
  async function handleUpgrade() {
    if (!agent.upgradeAvailable) {
      return;
    }

    try {
      await upgradeAgent(agent.id, agent.upgradeAvailable);
    } catch (err) {
      console.error(err);
    }
  }

  return (
    <>
      <Grid item xs={6} lg={3}>
        <Stack>
          <Typography classes={{ root: styles["key-column"] }}>
            {key}
          </Typography>
          <Stack direction="row" spacing={2} alignContent="center">
            <Typography classes={{ root: mixins["mt-1"] }}>
              {agent.version}
            </Typography>
            {agent.upgradeAvailable &&
              agent.status !== AgentStatus.DISCONNECTED && (
                <Button
                  endIcon={<ArrowUpIcon />}
                  size="small"
                  classes={{ root: mixins["ml-2"] }}
                  variant="outlined"
                  disabled={agent.status === AgentStatus.UPGRADING}
                  onClick={() => handleUpgrade()}
                >
                  Upgrade to {agent.upgradeAvailable}
                </Button>
              )}
          </Stack>
        </Stack>
      </Grid>
    </>
  );
}

function renderRow(key: string, value: JSX.Element): JSX.Element {
  return (
    <Grid item xs={6} lg={3}>
      <Stack sx={{ width: 200 }}>
        <Typography classes={{ root: styles["key-column"] }}>{key}</Typography>
        <Typography classes={{ root: mixins["mt-1"] }}>{value}</Typography>
      </Stack>
    </Grid>
  );
}

function renderLabelsRow(key: string, value: JSX.Element): JSX.Element {
  return (
    <Grid item xs={12} lg={12}>
      <Stack sx={{ width: 200 }}>
        <Typography classes={{ root: styles["key-column"] }}>{key}</Typography>
        <Stack classes={{ root: mixins["mt-1"] }} direction="row" spacing={1}>
          {value}
        </Stack>
      </Stack>
    </Grid>
  );
}
