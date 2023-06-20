import React from "react";
import { useNavigate } from "react-router-dom";
import { withNavBar } from "../../components/NavBar";
import { withRequireLogin } from "../../contexts/RequireLogin";
import { RawConfigWizard } from "./wizards/RawConfigWizard";

export const NewRawConfigurationPageContent: React.FC = () => {
  const navigate = useNavigate();

  return <RawConfigWizard onSuccess={() => navigate("/configurations")} />;
};

export const NewRawConfigurationPage: React.FC = withRequireLogin(
  withNavBar(NewRawConfigurationPageContent)
);
