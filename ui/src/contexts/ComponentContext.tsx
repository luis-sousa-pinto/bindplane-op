import { createContext } from "react";
import { ProcessorDialog } from "../components/ResourceDialog/ProcessorsDialog";

interface ComponentContextValue {
  ProcessorDialog: React.FC;
}

const defaultValue: ComponentContextValue = {
  ProcessorDialog: ProcessorDialog,
};

export const componentContext = createContext(defaultValue);

export const ComponentContextProvider: React.FC<ComponentContextValue> = ({
  ProcessorDialog: ProcessorDialogProp,
  children,
}) => {
  return (
    <componentContext.Provider value={{ ProcessorDialog: ProcessorDialogProp }}>
      {children}
    </componentContext.Provider>
  );
};
