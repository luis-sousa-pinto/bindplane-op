import { Card, Paper } from "@mui/material";
import { ReactFlowProvider } from "reactflow";
import { ShowPageConfig } from "../../pages/configurations/configuration";
import { DEFAULT_PERIOD } from "../MeasurementControlBar/MeasurementControlBar";
import { ConfigurationFlow } from "./ConfigurationFlow";
import { PipelineGraphProvider } from "./PipelineGraphContext";
import { ProcessorDialog } from "../ResourceDialog/ProcessorsDialog";
import {
  useConfigurationMetricsSubscription,
  useGetConfigurationQuery,
} from "../../graphql/generated";
import { useState } from "react";
import { AddSourcesSection } from "../../pages/configurations/configuration/AddSourcesSection";
import { AddDestinationsSection } from "../../pages/configurations/configuration/AddDestinationsSection";
import { trimVersion } from "../../utils/version-helpers";
import { ApolloError, gql } from "@apollo/client";
import { useSnackbar } from "notistack";
import { Page } from "../../utils/graph/utils";
import { GraphGradient, MaxValueMap } from "../GraphComponents";

import styles from "./pipeline-graph.module.scss";

gql`
  subscription ConfigurationMetrics(
    $period: String!
    $name: String!
    $agent: String
  ) {
    configurationMetrics(period: $period, name: $name, agent: $agent) {
      metrics {
        name
        nodeID
        pipelineType
        value
        unit
      }
      maxMetricValue
      maxLogValue
      maxTraceValue
    }
  }
`;

export type MinimumRequiredConfig = Partial<ShowPageConfig>;

interface PipelineGraphProps {
  selectedTelemetry: string;
  period: string;
  // configurationName is the versioned configuration name
  configurationName: string;
  // agentId can be specified to show the pipeline/telemetry for an agent
  agentId?: string;
  // readOnly will set edit dialogs to be read only
  readOnly?: boolean;

  // skipMeasurements will skip the subscription for pipeline measurements
  skipMeasurements?: boolean;
}

export const PipelineGraph: React.FC<PipelineGraphProps> = ({
  configurationName,
  agentId,
  selectedTelemetry,
  period,
  readOnly,
  skipMeasurements,
}) => {
  const { enqueueSnackbar } = useSnackbar();

  const [addSourceOpen, setAddSourceOpen] = useState(false);
  const [addDestinationOpen, setAddDestinationOpen] = useState(false);
  const [maxValues, setMaxValues] = useState<MaxValueMap>({
    maxMetricValue: 0,
    maxLogValue: 0,
    maxTraceValue: 0,
  });

  function onError(err: ApolloError) {
    console.error(err);
    enqueueSnackbar(err.message, { variant: "error" });
  }

  const { data, refetch: refetchConfiguration } = useGetConfigurationQuery({
    variables: {
      name: configurationName,
    },
    fetchPolicy: "cache-and-network",
    onError,
  });

  const { data: measurementData } = useConfigurationMetricsSubscription({
    variables: {
      period,
      name: trimVersion(configurationName),
      agent: agentId,
    },
    onError,
    onData({ data }) {
      if (data.data?.configurationMetrics) {
        setMaxValues({
          maxMetricValue: data.data.configurationMetrics.maxMetricValue,
          maxLogValue: data.data.configurationMetrics.maxLogValue,
          maxTraceValue: data.data.configurationMetrics.maxTraceValue,
        });
      }
    },
    skip: skipMeasurements,
  });

  return (
    <PipelineGraphProvider
      selectedTelemetryType={selectedTelemetry || DEFAULT_PERIOD}
      configuration={data?.configuration}
      refetchConfiguration={refetchConfiguration}
      addSourceOpen={addSourceOpen}
      setAddSourceOpen={setAddSourceOpen}
      addDestinationOpen={addDestinationOpen}
      setAddDestinationOpen={setAddDestinationOpen}
      readOnly={readOnly}
      maxValues={maxValues}
    >
      <GraphContainer>
        <Card className={styles.card}>
          <ReactFlowProvider>
            <ConfigurationFlow
              period={period}
              selectedTelemetry={selectedTelemetry}
              page={agentId ? Page.Agent : Page.Configuration}
              loading={data?.configuration == null}
              measurementData={measurementData}
            />
          </ReactFlowProvider>
        </Card>
      </GraphContainer>
      <ProcessorDialog />

      {!readOnly && data?.configuration && (
        <>
          <AddSourcesSection
            configuration={data.configuration}
            refetch={refetchConfiguration}
            setAddDialogOpen={setAddSourceOpen}
            addDialogOpen={addSourceOpen}
          />
          <AddDestinationsSection
            configuration={data.configuration}
            refetch={refetchConfiguration}
            setAddDialogOpen={setAddDestinationOpen}
            addDialogOpen={addDestinationOpen}
          />
        </>
      )}

      <GraphGradient />
    </PipelineGraphProvider>
  );
};

const GraphContainer: React.FC = ({ children }) => {
  return (
    <Paper classes={{ root: styles.container }} elevation={1}>
      {children}
    </Paper>
  );
};
