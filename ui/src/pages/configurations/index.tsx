import { Button } from "@mui/material";
import React, { useState } from "react";
import { Link } from "react-router-dom";
import { CardContainer } from "../../components/CardContainer";
import { ConfigurationsTable } from "../../components/Tables/ConfigurationTable";
import { PlusCircleIcon } from "../../components/Icons";
import { withRequireLogin } from "../../contexts/RequireLogin";
import { withNavBar } from "../../components/NavBar";
import { GridRowSelectionModel } from "@mui/x-data-grid";
import { hasPermission } from "../../utils/has-permission";
import { Role } from "../../graphql/generated";
import { useRole } from "../../hooks/useRole";
import { RBACWrapper } from "../../components/RBACWrapper/RBACWrapper";

import mixins from "../../styles/mixins.module.scss";

export const ConfigurationsPageContent: React.FC = () => {
  // Selected is an array of names of configurations.
  const [selected, setSelected] = useState<GridRowSelectionModel>([]);
  const role = useRole();

  return (
    <CardContainer>
      <RBACWrapper requiredRole={Role.User}>
        <Button
          component={Link}
          to="/configurations/new"
          variant="contained"
          classes={{ root: mixins["float-right"] }}
          startIcon={<PlusCircleIcon />}
        >
          New Configuration
        </Button>
      </RBACWrapper>

      <ConfigurationsTable
        allowSelection={hasPermission(Role.User, role)}
        selected={selected}
        setSelected={setSelected}
        enableDelete={true}
      />
    </CardContainer>
  );
};

export const ConfigurationsPage = withRequireLogin(
  withNavBar(ConfigurationsPageContent)
);
