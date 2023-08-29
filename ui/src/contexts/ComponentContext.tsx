import { createContext } from "react";
import { ProcessorDialog } from "../components/ResourceDialog/ProcessorsDialog";
import { SettingsMenu, SettingsMenuProps } from "../components/SettingsMenu";
import {
  AgentsTable,
  AgentsTableProps,
} from "../components/Tables/AgentsTable";

interface ComponentContextValue {
  ProcessorDialog: React.FC;
  SettingsMenu: React.FC<SettingsMenuProps>;
  AgentsTable: React.FC<AgentsTableProps>;
}

const defaultValue: ComponentContextValue = {
  ProcessorDialog: ProcessorDialog,
  SettingsMenu: SettingsMenu,
  AgentsTable: AgentsTable,
};

export const componentContext = createContext(defaultValue);

export const ComponentContextProvider: React.FC<ComponentContextValue> = ({
  ProcessorDialog: ProcessorDialogProp,
  SettingsMenu: SettingsMenuProp,
  AgentsTable: AgentsTableProp,
  children,
}) => {
  return (
    <componentContext.Provider
      value={{
        ProcessorDialog: ProcessorDialogProp,
        SettingsMenu: SettingsMenuProp,
        AgentsTable: AgentsTableProp,
      }}
    >
      {children}
    </componentContext.Provider>
  );
};
