import React, { useState } from "react";
import { CardContainer } from "../../components/CardContainer";
import { ConfigurationsTable } from "../../components/Tables/ConfigurationTable";
import { withRequireLogin } from "../../contexts/RequireLogin";
import { withNavBar } from "../../components/NavBar";
import { GridRowSelectionModel } from "@mui/x-data-grid";
import { hasPermission } from "../../utils/has-permission";
import { Role } from "../../graphql/generated";
import { useRole } from "../../hooks/useRole";

export const ConfigurationsPageContent: React.FC = () => {
  // Selected is an array of names of configurations.
  const [selected, setSelected] = useState<GridRowSelectionModel>([]);
  const role = useRole();

  return (
    <CardContainer>
      <ConfigurationsTable
        allowSelection={hasPermission(Role.User, role)}
        selected={selected}
        setSelected={setSelected}
        enableDelete={true}
        maxHeight="65vh"
        minHeight="65vh"
      />
    </CardContainer>
  );
};

export const ConfigurationsPage = withRequireLogin(
  withNavBar(ConfigurationsPageContent)
);
