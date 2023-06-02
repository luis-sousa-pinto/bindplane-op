import { createContext, useContext, useState } from "react";
import { DEFAULT_TELEMETRY_TYPE } from "../MeasurementControlBar/MeasurementControlBar";
import { MinimumRequiredConfig } from "./PipelineGraph";
import { MaxValueMap } from "../GraphComponents";

export interface PipelineGraphContextValue {
  configuration: MinimumRequiredConfig;
  refetchConfiguration: () => void;

  selectedTelemetryType: string;
  hoveredSet: string[];
  setHoveredNodeAndEdgeSet: React.Dispatch<React.SetStateAction<string[]>>;

  editProcessorsInfo: EditProcessorsInfo | null;
  // editProcessor opens up the editing dialog for a source or destination
  editProcessors: (
    resourceType: "source" | "destination",
    index: number
  ) => void;
  closeProcessorDialog(): void;
  editProcessorsOpen: boolean;
  maxValues: MaxValueMap;

  readOnlyGraph?: boolean;

  addSourceOpen: boolean;
  addDestinationOpen: boolean;
  setAddSourceOpen: React.Dispatch<React.SetStateAction<boolean>>;
  setAddDestinationOpen: React.Dispatch<React.SetStateAction<boolean>>;
}
export interface PipelineGraphProviderProps {
  configuration: MinimumRequiredConfig;
  refetchConfiguration: () => void;
  selectedTelemetryType: string;
  readOnly?: boolean;
  addSourceOpen: boolean;
  addDestinationOpen: boolean;
  setAddSourceOpen: React.Dispatch<React.SetStateAction<boolean>>;
  setAddDestinationOpen: React.Dispatch<React.SetStateAction<boolean>>;
  maxValues: MaxValueMap;
}

interface EditProcessorsInfo {
  resourceType: "source" | "destination";
  index: number;
}

const defaultValue: PipelineGraphContextValue = {
  configuration: {
    __typename: undefined,
    metadata: {
      __typename: undefined,
      id: "",
      name: "",
      description: undefined,
      labels: undefined,
      version: 0,
    },
    spec: {
      __typename: undefined,
      raw: undefined,
      sources: undefined,
      destinations: undefined,
      selector: undefined,
    },
    graph: undefined,
  },
  refetchConfiguration: () => {},
  selectedTelemetryType: DEFAULT_TELEMETRY_TYPE,
  hoveredSet: [],
  setHoveredNodeAndEdgeSet: () => {},
  editProcessors: () => {},
  closeProcessorDialog: () => {},
  editProcessorsInfo: null,
  editProcessorsOpen: false,
  maxValues: {
    maxMetricValue: 0,
    maxLogValue: 0,
    maxTraceValue: 0,
  },
  setAddSourceOpen: () => {},
  setAddDestinationOpen: () => {},
  addSourceOpen: false,
  addDestinationOpen: false,
};

export const PipelineContext = createContext(defaultValue);

export const PipelineGraphProvider: React.FC<PipelineGraphProviderProps> = ({
  children,
  selectedTelemetryType,
  configuration,
  refetchConfiguration,
  setAddSourceOpen,
  setAddDestinationOpen,
  addSourceOpen,
  addDestinationOpen,
  readOnly,
  maxValues,
}) => {
  const [hoveredSet, setHoveredNodeAndEdgeSet] = useState<string[]>([]);

  const [editProcessorsInfo, setEditingProcessors] = useState<{
    resourceType: "source" | "destination";
    index: number;
  } | null>(null);

  const [editProcessorsOpen, setEditProcessorsOpen] = useState(false);

  function editProcessors(
    resourceType: "source" | "destination",
    index: number
  ) {
    setEditingProcessors({ resourceType, index });
    setEditProcessorsOpen(true);
  }
  function closeProcessorDialog() {
    setEditProcessorsOpen(false);
    // Reset the editing processors on a timeout to avoid a flash of empty state.
    setTimeout(() => {
      setEditingProcessors(null);
    }, 300);
  }
  return (
    <PipelineContext.Provider
      value={{
        configuration,
        refetchConfiguration,
        setHoveredNodeAndEdgeSet,
        hoveredSet,
        selectedTelemetryType,
        editProcessors,
        closeProcessorDialog,
        editProcessorsInfo,
        editProcessorsOpen,
        readOnlyGraph: readOnly,
        setAddSourceOpen,
        setAddDestinationOpen,
        addSourceOpen,
        addDestinationOpen,
        maxValues,
      }}
    >
      {children}
    </PipelineContext.Provider>
  );
};

export function usePipelineGraph(): PipelineGraphContextValue {
  return useContext(PipelineContext);
}
