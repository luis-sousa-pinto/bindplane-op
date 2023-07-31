import { createContext } from "react";
import { ProcessorDialog } from "../components/ResourceDialog/ProcessorsDialog";
import { SettingsMenu, SettingsMenuProps } from "../components/SettingsMenu";

interface ComponentContextValue {
  ProcessorDialog: React.FC;
  SettingsMenu: React.FC<SettingsMenuProps>;
}

const defaultValue: ComponentContextValue = {
  ProcessorDialog: ProcessorDialog,
  SettingsMenu: SettingsMenu,
};

export const componentContext = createContext(defaultValue);

export const ComponentContextProvider: React.FC<ComponentContextValue> = ({
  ProcessorDialog: ProcessorDialogProp,
  SettingsMenu: SettingsMenuProp,
  children,
}) => {
  return (
    <componentContext.Provider
      value={{
        ProcessorDialog: ProcessorDialogProp,
        SettingsMenu: SettingsMenuProp,
      }}
    >
      {children}
    </componentContext.Provider>
  );
};
