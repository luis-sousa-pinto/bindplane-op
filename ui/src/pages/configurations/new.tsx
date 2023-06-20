import React from "react";
import { withNavBar } from "../../components/NavBar";
import { withRequireLogin } from "../../contexts/RequireLogin";
import { AssistedConfigWizard } from "./wizards/AssistedConfigWizard";

export const NewConfigurationPageContent: React.FC = () => {
  return <AssistedConfigWizard />;
};

export const NewConfigurationPage: React.FC = withRequireLogin(
  withNavBar(NewConfigurationPageContent)
);
