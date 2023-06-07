import { gql } from '@apollo/client';
import * as Apollo from '@apollo/client';
export type Maybe<T> = T | null;
export type InputMaybe<T> = Maybe<T>;
export type Exact<T extends { [key: string]: unknown }> = { [K in keyof T]: T[K] };
export type MakeOptional<T, K extends keyof T> = Omit<T, K> & { [SubKey in K]?: Maybe<T[SubKey]> };
export type MakeMaybe<T, K extends keyof T> = Omit<T, K> & { [SubKey in K]: Maybe<T[SubKey]> };
export type MakeEmpty<T extends { [key: string]: unknown }, K extends keyof T> = { [_ in K]?: never };
export type Incremental<T> = T | { [P in keyof T]?: P extends ' $fragmentName' | '__typename' ? T[P] : never };
const defaultOptions = {} as const;
/** All built-in and custom scalars, mapped to their actual values */
export type Scalars = {
  ID: { input: string | number; output: string; }
  String: { input: string; output: string; }
  Boolean: { input: boolean; output: boolean; }
  Int: { input: number; output: number; }
  Float: { input: number; output: number; }
  Any: { input: any; output: any; }
  Map: { input: any; output: any; }
  RolloutStatus: { input: number; output: number; }
  Time: { input: any; output: any; }
  Version: { input: number; output: number; }
};

export type Agent = {
  __typename?: 'Agent';
  architecture?: Maybe<Scalars['String']['output']>;
  configuration?: Maybe<AgentConfiguration>;
  configurationResource?: Maybe<Configuration>;
  connectedAt?: Maybe<Scalars['Time']['output']>;
  disconnectedAt?: Maybe<Scalars['Time']['output']>;
  errorMessage?: Maybe<Scalars['String']['output']>;
  features: Scalars['Int']['output'];
  home?: Maybe<Scalars['String']['output']>;
  hostName?: Maybe<Scalars['String']['output']>;
  id: Scalars['ID']['output'];
  labels?: Maybe<Scalars['Map']['output']>;
  macAddress?: Maybe<Scalars['String']['output']>;
  name: Scalars['String']['output'];
  operatingSystem?: Maybe<Scalars['String']['output']>;
  platform?: Maybe<Scalars['String']['output']>;
  remoteAddress?: Maybe<Scalars['String']['output']>;
  status: Scalars['Int']['output'];
  type?: Maybe<Scalars['String']['output']>;
  upgrade?: Maybe<AgentUpgrade>;
  upgradeAvailable?: Maybe<Scalars['String']['output']>;
  version?: Maybe<Scalars['String']['output']>;
};

export type AgentChange = {
  __typename?: 'AgentChange';
  agent: Agent;
  changeType: AgentChangeType;
};

export enum AgentChangeType {
  Insert = 'INSERT',
  Remove = 'REMOVE',
  Update = 'UPDATE'
}

export type AgentConfiguration = {
  __typename?: 'AgentConfiguration';
  Collector?: Maybe<Scalars['String']['output']>;
  Logging?: Maybe<Scalars['String']['output']>;
  Manager?: Maybe<Scalars['Map']['output']>;
};

export type AgentSelector = {
  __typename?: 'AgentSelector';
  matchLabels?: Maybe<Scalars['Map']['output']>;
};

export type AgentUpgrade = {
  __typename?: 'AgentUpgrade';
  error?: Maybe<Scalars['String']['output']>;
  status: Scalars['Int']['output'];
  version: Scalars['String']['output'];
};

export type Agents = {
  __typename?: 'Agents';
  agents: Array<Agent>;
  latestVersion: Scalars['String']['output'];
  query?: Maybe<Scalars['String']['output']>;
  suggestions?: Maybe<Array<Suggestion>>;
};

export type ClearAgentUpgradeErrorInput = {
  agentId: Scalars['String']['input'];
};

export type Configuration = {
  __typename?: 'Configuration';
  activeTypes?: Maybe<Array<Scalars['String']['output']>>;
  agentCount?: Maybe<Scalars['Int']['output']>;
  apiVersion: Scalars['String']['output'];
  graph?: Maybe<Graph>;
  kind: Scalars['String']['output'];
  metadata: Metadata;
  rendered?: Maybe<Scalars['String']['output']>;
  spec: ConfigurationSpec;
  status: ConfigurationStatus;
};

export type ConfigurationChange = {
  __typename?: 'ConfigurationChange';
  configuration: Configuration;
  eventType: EventType;
};

export type ConfigurationSpec = {
  __typename?: 'ConfigurationSpec';
  contentType?: Maybe<Scalars['String']['output']>;
  destinations?: Maybe<Array<ResourceConfiguration>>;
  raw?: Maybe<Scalars['String']['output']>;
  selector?: Maybe<AgentSelector>;
  sources?: Maybe<Array<ResourceConfiguration>>;
};

export type ConfigurationStatus = {
  __typename?: 'ConfigurationStatus';
  current: Scalars['Boolean']['output'];
  currentVersion: Scalars['Version']['output'];
  latest: Scalars['Boolean']['output'];
  pending: Scalars['Boolean']['output'];
  rollout: Rollout;
};

export type Configurations = {
  __typename?: 'Configurations';
  configurations: Array<Configuration>;
  query?: Maybe<Scalars['String']['output']>;
  suggestions?: Maybe<Array<Suggestion>>;
};

export type Destination = {
  __typename?: 'Destination';
  apiVersion: Scalars['String']['output'];
  kind: Scalars['String']['output'];
  metadata: Metadata;
  spec: ParameterizedSpec;
};

export type DestinationType = {
  __typename?: 'DestinationType';
  apiVersion: Scalars['String']['output'];
  kind: Scalars['String']['output'];
  metadata: Metadata;
  spec: ResourceTypeSpec;
};

export type DestinationWithType = {
  __typename?: 'DestinationWithType';
  destination?: Maybe<Destination>;
  destinationType?: Maybe<DestinationType>;
};

export type DocumentationLink = {
  __typename?: 'DocumentationLink';
  text: Scalars['String']['output'];
  url: Scalars['String']['output'];
};

export type Edge = {
  __typename?: 'Edge';
  id: Scalars['String']['output'];
  source: Scalars['String']['output'];
  target: Scalars['String']['output'];
};

export type EditConfigurationDescriptionInput = {
  description: Scalars['String']['input'];
  name: Scalars['String']['input'];
};

export enum EventType {
  Insert = 'INSERT',
  Remove = 'REMOVE',
  Update = 'UPDATE'
}

export type Graph = {
  __typename?: 'Graph';
  attributes: Scalars['Map']['output'];
  edges: Array<Edge>;
  intermediates: Array<Node>;
  sources: Array<Node>;
  targets: Array<Node>;
};

export type GraphMetric = {
  __typename?: 'GraphMetric';
  agentID?: Maybe<Scalars['ID']['output']>;
  name: Scalars['String']['output'];
  nodeID: Scalars['String']['output'];
  pipelineType: Scalars['String']['output'];
  unit: Scalars['String']['output'];
  value: Scalars['Float']['output'];
};

export type GraphMetrics = {
  __typename?: 'GraphMetrics';
  maxLogValue: Scalars['Float']['output'];
  maxMetricValue: Scalars['Float']['output'];
  maxTraceValue: Scalars['Float']['output'];
  metrics: Array<GraphMetric>;
};

export type Log = {
  __typename?: 'Log';
  attributes?: Maybe<Scalars['Map']['output']>;
  body?: Maybe<Scalars['Any']['output']>;
  resource?: Maybe<Scalars['Map']['output']>;
  severity?: Maybe<Scalars['String']['output']>;
  timestamp?: Maybe<Scalars['Time']['output']>;
};

export type Metadata = {
  __typename?: 'Metadata';
  dateModified?: Maybe<Scalars['Time']['output']>;
  description?: Maybe<Scalars['String']['output']>;
  displayName?: Maybe<Scalars['String']['output']>;
  icon?: Maybe<Scalars['String']['output']>;
  id: Scalars['ID']['output'];
  labels?: Maybe<Scalars['Map']['output']>;
  name: Scalars['String']['output'];
  version: Scalars['Version']['output'];
};

export type Metric = {
  __typename?: 'Metric';
  attributes?: Maybe<Scalars['Map']['output']>;
  name?: Maybe<Scalars['String']['output']>;
  resource?: Maybe<Scalars['Map']['output']>;
  timestamp?: Maybe<Scalars['Time']['output']>;
  type?: Maybe<Scalars['String']['output']>;
  unit?: Maybe<Scalars['String']['output']>;
  value?: Maybe<Scalars['Any']['output']>;
};

export type MetricCategory = {
  __typename?: 'MetricCategory';
  column: Scalars['Int']['output'];
  label: Scalars['String']['output'];
  metrics: Array<MetricOption>;
};

export type MetricOption = {
  __typename?: 'MetricOption';
  defaultDisabled?: Maybe<Scalars['Boolean']['output']>;
  description?: Maybe<Scalars['String']['output']>;
  kpi?: Maybe<Scalars['Boolean']['output']>;
  name: Scalars['String']['output'];
};

export type Mutation = {
  __typename?: 'Mutation';
  clearAgentUpgradeError?: Maybe<Scalars['Boolean']['output']>;
  editConfigurationDescription?: Maybe<Scalars['Boolean']['output']>;
  removeAgentConfiguration?: Maybe<Agent>;
  updateProcessors?: Maybe<Scalars['Boolean']['output']>;
};


export type MutationClearAgentUpgradeErrorArgs = {
  input: ClearAgentUpgradeErrorInput;
};


export type MutationEditConfigurationDescriptionArgs = {
  input: EditConfigurationDescriptionInput;
};


export type MutationRemoveAgentConfigurationArgs = {
  input?: InputMaybe<RemoveAgentConfigurationInput>;
};


export type MutationUpdateProcessorsArgs = {
  input: UpdateProcessorsInput;
};

export type Node = {
  __typename?: 'Node';
  attributes: Scalars['Map']['output'];
  id: Scalars['String']['output'];
  label: Scalars['String']['output'];
  type: Scalars['String']['output'];
};

export type OverviewPage = {
  __typename?: 'OverviewPage';
  graph: Graph;
};

export type Parameter = {
  __typename?: 'Parameter';
  name: Scalars['String']['output'];
  value: Scalars['Any']['output'];
};

export type ParameterDefinition = {
  __typename?: 'ParameterDefinition';
  advancedConfig?: Maybe<Scalars['Boolean']['output']>;
  default?: Maybe<Scalars['Any']['output']>;
  description: Scalars['String']['output'];
  documentation?: Maybe<Array<DocumentationLink>>;
  label: Scalars['String']['output'];
  name: Scalars['String']['output'];
  options: ParameterOptions;
  relevantIf?: Maybe<Array<RelevantIfCondition>>;
  required: Scalars['Boolean']['output'];
  type: ParameterType;
  validValues?: Maybe<Array<Scalars['String']['output']>>;
};

export type ParameterInput = {
  name: Scalars['String']['input'];
  value: Scalars['Any']['input'];
};

export type ParameterOptions = {
  __typename?: 'ParameterOptions';
  creatable?: Maybe<Scalars['Boolean']['output']>;
  gridColumns?: Maybe<Scalars['Int']['output']>;
  labels?: Maybe<Scalars['Map']['output']>;
  metricCategories?: Maybe<Array<MetricCategory>>;
  multiline?: Maybe<Scalars['Boolean']['output']>;
  password?: Maybe<Scalars['Boolean']['output']>;
  sectionHeader?: Maybe<Scalars['Boolean']['output']>;
  trackUnchecked?: Maybe<Scalars['Boolean']['output']>;
};

export enum ParameterType {
  AwsCloudwatchNamedField = 'awsCloudwatchNamedField',
  Bool = 'bool',
  Enum = 'enum',
  Enums = 'enums',
  Int = 'int',
  Map = 'map',
  Metrics = 'metrics',
  String = 'string',
  Strings = 'strings',
  Timezone = 'timezone',
  Yaml = 'yaml'
}

export type ParameterizedSpec = {
  __typename?: 'ParameterizedSpec';
  disabled: Scalars['Boolean']['output'];
  parameters?: Maybe<Array<Parameter>>;
  processors?: Maybe<Array<ResourceConfiguration>>;
  type: Scalars['String']['output'];
};

export type PhaseAgentCount = {
  __typename?: 'PhaseAgentCount';
  initial: Scalars['Int']['output'];
  maximum: Scalars['Int']['output'];
  multiplier: Scalars['Float']['output'];
};

export enum PipelineType {
  Logs = 'logs',
  Metrics = 'metrics',
  Traces = 'traces'
}

export type Processor = {
  __typename?: 'Processor';
  apiVersion: Scalars['String']['output'];
  kind: Scalars['String']['output'];
  metadata: Metadata;
  spec: ParameterizedSpec;
};

export type ProcessorInput = {
  disabled?: InputMaybe<Scalars['Boolean']['input']>;
  displayName?: InputMaybe<Scalars['String']['input']>;
  name?: InputMaybe<Scalars['String']['input']>;
  parameters?: InputMaybe<Array<ParameterInput>>;
  type?: InputMaybe<Scalars['String']['input']>;
};

export type ProcessorType = {
  __typename?: 'ProcessorType';
  apiVersion: Scalars['String']['output'];
  kind: Scalars['String']['output'];
  metadata: Metadata;
  spec: ResourceTypeSpec;
};

export type Query = {
  __typename?: 'Query';
  agent?: Maybe<Agent>;
  agentMetrics: GraphMetrics;
  agents: Agents;
  configuration?: Maybe<Configuration>;
  configurationHistory: Array<Configuration>;
  configurationMetrics: GraphMetrics;
  configurations: Configurations;
  destination?: Maybe<Destination>;
  destinationType?: Maybe<DestinationType>;
  destinationTypes: Array<DestinationType>;
  destinationWithType: DestinationWithType;
  destinations: Array<Destination>;
  destinationsInConfigs: Array<Destination>;
  overviewMetrics: GraphMetrics;
  overviewPage: OverviewPage;
  processor?: Maybe<Processor>;
  processorType?: Maybe<ProcessorType>;
  processorTypes: Array<ProcessorType>;
  processors: Array<Processor>;
  snapshot: Snapshot;
  source?: Maybe<Source>;
  sourceType?: Maybe<SourceType>;
  sourceTypes: Array<SourceType>;
  sources: Array<Source>;
};


export type QueryAgentArgs = {
  id: Scalars['ID']['input'];
};


export type QueryAgentMetricsArgs = {
  ids?: InputMaybe<Array<Scalars['ID']['input']>>;
  period: Scalars['String']['input'];
};


export type QueryAgentsArgs = {
  query?: InputMaybe<Scalars['String']['input']>;
  selector?: InputMaybe<Scalars['String']['input']>;
};


export type QueryConfigurationArgs = {
  name: Scalars['String']['input'];
};


export type QueryConfigurationHistoryArgs = {
  name: Scalars['String']['input'];
};


export type QueryConfigurationMetricsArgs = {
  name?: InputMaybe<Scalars['String']['input']>;
  period: Scalars['String']['input'];
};


export type QueryConfigurationsArgs = {
  onlyDeployedConfigurations?: InputMaybe<Scalars['Boolean']['input']>;
  query?: InputMaybe<Scalars['String']['input']>;
  selector?: InputMaybe<Scalars['String']['input']>;
};


export type QueryDestinationArgs = {
  name: Scalars['String']['input'];
};


export type QueryDestinationTypeArgs = {
  name: Scalars['String']['input'];
};


export type QueryDestinationWithTypeArgs = {
  name: Scalars['String']['input'];
};


export type QueryOverviewMetricsArgs = {
  configIDs?: InputMaybe<Array<Scalars['ID']['input']>>;
  destinationIDs?: InputMaybe<Array<Scalars['ID']['input']>>;
  period: Scalars['String']['input'];
};


export type QueryOverviewPageArgs = {
  configIDs?: InputMaybe<Array<Scalars['ID']['input']>>;
  destinationIDs?: InputMaybe<Array<Scalars['ID']['input']>>;
  period: Scalars['String']['input'];
  telemetryType: Scalars['String']['input'];
};


export type QueryProcessorArgs = {
  name: Scalars['String']['input'];
};


export type QueryProcessorTypeArgs = {
  name: Scalars['String']['input'];
};


export type QuerySnapshotArgs = {
  agentID: Scalars['String']['input'];
  pipelineType: PipelineType;
  position?: InputMaybe<Scalars['String']['input']>;
  resourceName?: InputMaybe<Scalars['String']['input']>;
};


export type QuerySourceArgs = {
  name: Scalars['String']['input'];
};


export type QuerySourceTypeArgs = {
  name: Scalars['String']['input'];
};

export type RelevantIfCondition = {
  __typename?: 'RelevantIfCondition';
  name: Scalars['String']['output'];
  operator: RelevantIfOperatorType;
  value: Scalars['Any']['output'];
};

export enum RelevantIfOperatorType {
  ContainsAny = 'containsAny',
  Equals = 'equals',
  NotEquals = 'notEquals'
}

export type RemoveAgentConfigurationInput = {
  agentId: Scalars['String']['input'];
};

export type ResourceConfiguration = {
  __typename?: 'ResourceConfiguration';
  disabled: Scalars['Boolean']['output'];
  displayName?: Maybe<Scalars['String']['output']>;
  name?: Maybe<Scalars['String']['output']>;
  parameters?: Maybe<Array<Parameter>>;
  processors?: Maybe<Array<ResourceConfiguration>>;
  type?: Maybe<Scalars['String']['output']>;
};

export enum ResourceTypeKind {
  Destination = 'DESTINATION',
  Source = 'SOURCE'
}

export type ResourceTypeSpec = {
  __typename?: 'ResourceTypeSpec';
  parameters: Array<ParameterDefinition>;
  supportedPlatforms: Array<Scalars['String']['output']>;
  telemetryTypes: Array<PipelineType>;
  version: Scalars['String']['output'];
};

export type Rollout = {
  __typename?: 'Rollout';
  completed: Scalars['Int']['output'];
  errors: Scalars['Int']['output'];
  options?: Maybe<RolloutOptions>;
  pending: Scalars['Int']['output'];
  phase: Scalars['Int']['output'];
  status: Scalars['RolloutStatus']['output'];
  waiting: Scalars['Int']['output'];
};

export type RolloutOptions = {
  __typename?: 'RolloutOptions';
  maxErrors: Scalars['Int']['output'];
  phaseAgentCount?: Maybe<PhaseAgentCount>;
  rollbackOnFailure: Scalars['Boolean']['output'];
  startAutomatically: Scalars['Boolean']['output'];
};

export type Snapshot = {
  __typename?: 'Snapshot';
  logs: Array<Log>;
  metrics: Array<Metric>;
  traces: Array<Trace>;
};

export type Source = {
  __typename?: 'Source';
  apiVersion: Scalars['String']['output'];
  kind: Scalars['String']['output'];
  metadata: Metadata;
  spec: ParameterizedSpec;
};

export type SourceType = {
  __typename?: 'SourceType';
  apiVersion: Scalars['String']['output'];
  kind: Scalars['String']['output'];
  metadata: Metadata;
  spec: ResourceTypeSpec;
};

export type Subscription = {
  __typename?: 'Subscription';
  agentChanges: Array<AgentChange>;
  agentMetrics: GraphMetrics;
  configurationChanges: Array<ConfigurationChange>;
  configurationMetrics: GraphMetrics;
  overviewMetrics: GraphMetrics;
};


export type SubscriptionAgentChangesArgs = {
  query?: InputMaybe<Scalars['String']['input']>;
  selector?: InputMaybe<Scalars['String']['input']>;
};


export type SubscriptionAgentMetricsArgs = {
  ids?: InputMaybe<Array<Scalars['ID']['input']>>;
  period: Scalars['String']['input'];
};


export type SubscriptionConfigurationChangesArgs = {
  query?: InputMaybe<Scalars['String']['input']>;
  selector?: InputMaybe<Scalars['String']['input']>;
};


export type SubscriptionConfigurationMetricsArgs = {
  agent?: InputMaybe<Scalars['String']['input']>;
  name?: InputMaybe<Scalars['String']['input']>;
  period: Scalars['String']['input'];
};


export type SubscriptionOverviewMetricsArgs = {
  configIDs?: InputMaybe<Array<Scalars['ID']['input']>>;
  destinationIDs?: InputMaybe<Array<Scalars['ID']['input']>>;
  period: Scalars['String']['input'];
};

export type Suggestion = {
  __typename?: 'Suggestion';
  label: Scalars['String']['output'];
  query: Scalars['String']['output'];
};

export type Trace = {
  __typename?: 'Trace';
  attributes?: Maybe<Scalars['Map']['output']>;
  end?: Maybe<Scalars['Time']['output']>;
  name?: Maybe<Scalars['String']['output']>;
  parentSpanID?: Maybe<Scalars['String']['output']>;
  resource?: Maybe<Scalars['Map']['output']>;
  spanID?: Maybe<Scalars['String']['output']>;
  start?: Maybe<Scalars['Time']['output']>;
  traceID?: Maybe<Scalars['String']['output']>;
};

export type UpdateProcessorsInput = {
  configuration: Scalars['String']['input'];
  processors: Array<ProcessorInput>;
  resourceIndex: Scalars['Int']['input'];
  resourceType: ResourceTypeKind;
};

export type GetLatestConfigVersionQueryVariables = Exact<{
  name: Scalars['String']['input'];
}>;


export type GetLatestConfigVersionQuery = { __typename?: 'Query', configuration?: { __typename?: 'Configuration', metadata: { __typename?: 'Metadata', id: string, name: string, version: number } } | null };

export type GetRenderedConfigQueryVariables = Exact<{
  name: Scalars['String']['input'];
}>;


export type GetRenderedConfigQuery = { __typename?: 'Query', configuration?: { __typename?: 'Configuration', rendered?: string | null, metadata: { __typename?: 'Metadata', name: string, id: string, version: number } } | null };

export type SourceTypeQueryVariables = Exact<{
  name: Scalars['String']['input'];
}>;


export type SourceTypeQuery = { __typename?: 'Query', sourceType?: { __typename?: 'SourceType', metadata: { __typename?: 'Metadata', id: string, name: string, version: number, displayName?: string | null, icon?: string | null, description?: string | null }, spec: { __typename?: 'ResourceTypeSpec', parameters: Array<{ __typename?: 'ParameterDefinition', label: string, name: string, description: string, required: boolean, type: ParameterType, default?: any | null, advancedConfig?: boolean | null, validValues?: Array<string> | null, documentation?: Array<{ __typename?: 'DocumentationLink', text: string, url: string }> | null, relevantIf?: Array<{ __typename?: 'RelevantIfCondition', name: string, operator: RelevantIfOperatorType, value: any }> | null, options: { __typename?: 'ParameterOptions', creatable?: boolean | null, trackUnchecked?: boolean | null, sectionHeader?: boolean | null, gridColumns?: number | null, labels?: any | null, password?: boolean | null, metricCategories?: Array<{ __typename?: 'MetricCategory', label: string, column: number, metrics: Array<{ __typename?: 'MetricOption', name: string, description?: string | null, kpi?: boolean | null }> }> | null } }> } } | null };

export type GetDestinationWithTypeQueryVariables = Exact<{
  name: Scalars['String']['input'];
}>;


export type GetDestinationWithTypeQuery = { __typename?: 'Query', destinationWithType: { __typename?: 'DestinationWithType', destination?: { __typename?: 'Destination', metadata: { __typename?: 'Metadata', name: string, version: number, id: string, labels?: any | null }, spec: { __typename?: 'ParameterizedSpec', type: string, disabled: boolean, parameters?: Array<{ __typename?: 'Parameter', name: string, value: any }> | null } } | null, destinationType?: { __typename?: 'DestinationType', metadata: { __typename?: 'Metadata', id: string, name: string, version: number, icon?: string | null, description?: string | null }, spec: { __typename?: 'ResourceTypeSpec', parameters: Array<{ __typename?: 'ParameterDefinition', label: string, name: string, description: string, required: boolean, type: ParameterType, default?: any | null, advancedConfig?: boolean | null, validValues?: Array<string> | null, relevantIf?: Array<{ __typename?: 'RelevantIfCondition', name: string, operator: RelevantIfOperatorType, value: any }> | null, documentation?: Array<{ __typename?: 'DocumentationLink', text: string, url: string }> | null, options: { __typename?: 'ParameterOptions', multiline?: boolean | null, creatable?: boolean | null, trackUnchecked?: boolean | null, sectionHeader?: boolean | null, gridColumns?: number | null, labels?: any | null, password?: boolean | null, metricCategories?: Array<{ __typename?: 'MetricCategory', label: string, column: number, metrics: Array<{ __typename?: 'MetricOption', name: string, description?: string | null, kpi?: boolean | null }> }> | null } }> } } | null } };

export type GetCurrentConfigVersionQueryVariables = Exact<{
  configurationName: Scalars['String']['input'];
}>;


export type GetCurrentConfigVersionQuery = { __typename?: 'Query', configuration?: { __typename?: 'Configuration', agentCount?: number | null, metadata: { __typename?: 'Metadata', id: string, name: string, version: number, labels?: any | null } } | null };

export type GetLatestConfigDescriptionQueryVariables = Exact<{
  configurationName: Scalars['String']['input'];
}>;


export type GetLatestConfigDescriptionQuery = { __typename?: 'Query', configuration?: { __typename?: 'Configuration', metadata: { __typename?: 'Metadata', id: string, name: string, version: number, description?: string | null } } | null };

export type EditConfigDescriptionMutationVariables = Exact<{
  input: EditConfigurationDescriptionInput;
}>;


export type EditConfigDescriptionMutation = { __typename?: 'Mutation', editConfigurationDescription?: boolean | null };

export type GetConfigurationVersionsQueryVariables = Exact<{
  name: Scalars['String']['input'];
}>;


export type GetConfigurationVersionsQuery = { __typename?: 'Query', configurationHistory: Array<{ __typename?: 'Configuration', activeTypes?: Array<string> | null, metadata: { __typename?: 'Metadata', name: string, id: string, version: number }, status: { __typename?: 'ConfigurationStatus', current: boolean, pending: boolean, latest: boolean } }> };

export type RemoveAgentConfigurationMutationVariables = Exact<{
  input: RemoveAgentConfigurationInput;
}>;


export type RemoveAgentConfigurationMutation = { __typename?: 'Mutation', removeAgentConfiguration?: { __typename?: 'Agent', id: string, configuration?: { __typename?: 'AgentConfiguration', Collector?: string | null, Logging?: string | null, Manager?: any | null } | null } | null };

export type ConfigurationMetricsSubscriptionVariables = Exact<{
  period: Scalars['String']['input'];
  name: Scalars['String']['input'];
  agent?: InputMaybe<Scalars['String']['input']>;
}>;


export type ConfigurationMetricsSubscription = { __typename?: 'Subscription', configurationMetrics: { __typename?: 'GraphMetrics', maxMetricValue: number, maxLogValue: number, maxTraceValue: number, metrics: Array<{ __typename?: 'GraphMetric', name: string, nodeID: string, pipelineType: string, value: number, unit: string }> } };

export type GetProcessorTypesQueryVariables = Exact<{ [key: string]: never; }>;


export type GetProcessorTypesQuery = { __typename?: 'Query', processorTypes: Array<{ __typename?: 'ProcessorType', metadata: { __typename?: 'Metadata', displayName?: string | null, description?: string | null, name: string, labels?: any | null, version: number, id: string }, spec: { __typename?: 'ResourceTypeSpec', telemetryTypes: Array<PipelineType>, parameters: Array<{ __typename?: 'ParameterDefinition', label: string, name: string, description: string, required: boolean, type: ParameterType, default?: any | null, advancedConfig?: boolean | null, validValues?: Array<string> | null, relevantIf?: Array<{ __typename?: 'RelevantIfCondition', name: string, operator: RelevantIfOperatorType, value: any }> | null, documentation?: Array<{ __typename?: 'DocumentationLink', text: string, url: string }> | null, options: { __typename?: 'ParameterOptions', creatable?: boolean | null, trackUnchecked?: boolean | null, gridColumns?: number | null, sectionHeader?: boolean | null, multiline?: boolean | null, labels?: any | null, password?: boolean | null, metricCategories?: Array<{ __typename?: 'MetricCategory', label: string, column: number, metrics: Array<{ __typename?: 'MetricOption', name: string, description?: string | null, kpi?: boolean | null }> }> | null } }> } }> };

export type GetProcessorTypeQueryVariables = Exact<{
  type: Scalars['String']['input'];
}>;


export type GetProcessorTypeQuery = { __typename?: 'Query', processorType?: { __typename?: 'ProcessorType', metadata: { __typename?: 'Metadata', displayName?: string | null, name: string, version: number, id: string, description?: string | null }, spec: { __typename?: 'ResourceTypeSpec', parameters: Array<{ __typename?: 'ParameterDefinition', label: string, name: string, description: string, required: boolean, type: ParameterType, default?: any | null, advancedConfig?: boolean | null, validValues?: Array<string> | null, relevantIf?: Array<{ __typename?: 'RelevantIfCondition', name: string, operator: RelevantIfOperatorType, value: any }> | null, documentation?: Array<{ __typename?: 'DocumentationLink', text: string, url: string }> | null, options: { __typename?: 'ParameterOptions', creatable?: boolean | null, trackUnchecked?: boolean | null, gridColumns?: number | null, sectionHeader?: boolean | null, multiline?: boolean | null, labels?: any | null, password?: boolean | null, metricCategories?: Array<{ __typename?: 'MetricCategory', label: string, column: number, metrics: Array<{ __typename?: 'MetricOption', name: string, description?: string | null, kpi?: boolean | null }> }> | null } }> } } | null };

export type ProcessorDialogSourceTypeQueryVariables = Exact<{
  name: Scalars['String']['input'];
}>;


export type ProcessorDialogSourceTypeQuery = { __typename?: 'Query', sourceType?: { __typename?: 'SourceType', metadata: { __typename?: 'Metadata', name: string, id: string, version: number, displayName?: string | null, description?: string | null }, spec: { __typename?: 'ResourceTypeSpec', telemetryTypes: Array<PipelineType> } } | null };

export type ProcessorDialogDestinationTypeQueryVariables = Exact<{
  name: Scalars['String']['input'];
}>;


export type ProcessorDialogDestinationTypeQuery = { __typename?: 'Query', destinationWithType: { __typename?: 'DestinationWithType', destinationType?: { __typename?: 'DestinationType', metadata: { __typename?: 'Metadata', id: string, name: string, version: number, displayName?: string | null, description?: string | null }, spec: { __typename?: 'ResourceTypeSpec', telemetryTypes: Array<PipelineType> } } | null } };

export type UpdateProcessorsMutationVariables = Exact<{
  input: UpdateProcessorsInput;
}>;


export type UpdateProcessorsMutation = { __typename?: 'Mutation', updateProcessors?: boolean | null };

export type GetRolloutHistoryQueryVariables = Exact<{
  name: Scalars['String']['input'];
}>;


export type GetRolloutHistoryQuery = { __typename?: 'Query', configurationHistory: Array<{ __typename?: 'Configuration', metadata: { __typename?: 'Metadata', name: string, id: string, version: number, dateModified?: any | null }, status: { __typename?: 'ConfigurationStatus', rollout: { __typename?: 'Rollout', status: number, errors: number } } }> };

export type GetConfigRolloutStatusQueryVariables = Exact<{
  name: Scalars['String']['input'];
}>;


export type GetConfigRolloutStatusQuery = { __typename?: 'Query', configuration?: { __typename?: 'Configuration', agentCount?: number | null, metadata: { __typename?: 'Metadata', name: string, id: string, version: number }, status: { __typename?: 'ConfigurationStatus', pending: boolean, current: boolean, latest: boolean, rollout: { __typename?: 'Rollout', status: number, phase: number, completed: number, errors: number, pending: number, waiting: number } } } | null };

export type AgentsWithConfigurationQueryVariables = Exact<{
  selector?: InputMaybe<Scalars['String']['input']>;
  query?: InputMaybe<Scalars['String']['input']>;
}>;


export type AgentsWithConfigurationQuery = { __typename?: 'Query', agents: { __typename?: 'Agents', agents: Array<{ __typename?: 'Agent', id: string, name: string }> } };

export type SnapshotQueryVariables = Exact<{
  agentID: Scalars['String']['input'];
  pipelineType: PipelineType;
  position?: InputMaybe<Scalars['String']['input']>;
  resourceName?: InputMaybe<Scalars['String']['input']>;
}>;


export type SnapshotQuery = { __typename?: 'Query', snapshot: { __typename?: 'Snapshot', metrics: Array<{ __typename?: 'Metric', name?: string | null, timestamp?: any | null, value?: any | null, unit?: string | null, type?: string | null, attributes?: any | null, resource?: any | null }>, logs: Array<{ __typename?: 'Log', timestamp?: any | null, body?: any | null, severity?: string | null, attributes?: any | null, resource?: any | null }>, traces: Array<{ __typename?: 'Trace', name?: string | null, traceID?: string | null, spanID?: string | null, parentSpanID?: string | null, start?: any | null, end?: any | null, attributes?: any | null, resource?: any | null }> } };

export type AgentsTableQueryVariables = Exact<{
  selector?: InputMaybe<Scalars['String']['input']>;
  query?: InputMaybe<Scalars['String']['input']>;
}>;


export type AgentsTableQuery = { __typename?: 'Query', agents: { __typename?: 'Agents', query?: string | null, latestVersion: string, agents: Array<{ __typename?: 'Agent', id: string, architecture?: string | null, hostName?: string | null, labels?: any | null, platform?: string | null, version?: string | null, name: string, home?: string | null, operatingSystem?: string | null, macAddress?: string | null, type?: string | null, status: number, connectedAt?: any | null, disconnectedAt?: any | null, configurationResource?: { __typename?: 'Configuration', metadata: { __typename?: 'Metadata', id: string, name: string, version: number } } | null }>, suggestions?: Array<{ __typename?: 'Suggestion', query: string, label: string }> | null } };

export type AgentsTableMetricsSubscriptionVariables = Exact<{
  period: Scalars['String']['input'];
  ids?: InputMaybe<Array<Scalars['ID']['input']> | Scalars['ID']['input']>;
}>;


export type AgentsTableMetricsSubscription = { __typename?: 'Subscription', agentMetrics: { __typename?: 'GraphMetrics', metrics: Array<{ __typename?: 'GraphMetric', name: string, nodeID: string, pipelineType: string, value: number, unit: string, agentID?: string | null }> } };

export type GetConfigurationTableQueryVariables = Exact<{
  selector?: InputMaybe<Scalars['String']['input']>;
  query?: InputMaybe<Scalars['String']['input']>;
  onlyDeployedConfigurations?: InputMaybe<Scalars['Boolean']['input']>;
}>;


export type GetConfigurationTableQuery = { __typename?: 'Query', configurations: { __typename?: 'Configurations', query?: string | null, configurations: Array<{ __typename?: 'Configuration', agentCount?: number | null, metadata: { __typename?: 'Metadata', id: string, version: number, name: string, labels?: any | null, description?: string | null } }>, suggestions?: Array<{ __typename?: 'Suggestion', query: string, label: string }> | null } };

export type ConfigurationChangesSubscriptionVariables = Exact<{
  selector?: InputMaybe<Scalars['String']['input']>;
  query?: InputMaybe<Scalars['String']['input']>;
}>;


export type ConfigurationChangesSubscription = { __typename?: 'Subscription', configurationChanges: Array<{ __typename?: 'ConfigurationChange', eventType: EventType, configuration: { __typename?: 'Configuration', agentCount?: number | null, metadata: { __typename?: 'Metadata', id: string, version: number, name: string, description?: string | null, labels?: any | null } } }> };

export type ConfigurationTableMetricsSubscriptionVariables = Exact<{
  period: Scalars['String']['input'];
}>;


export type ConfigurationTableMetricsSubscription = { __typename?: 'Subscription', overviewMetrics: { __typename?: 'GraphMetrics', metrics: Array<{ __typename?: 'GraphMetric', name: string, nodeID: string, pipelineType: string, value: number, unit: string }> } };

export type GetDestinationTypeDisplayInfoQueryVariables = Exact<{
  name: Scalars['String']['input'];
}>;


export type GetDestinationTypeDisplayInfoQuery = { __typename?: 'Query', destinationType?: { __typename?: 'DestinationType', metadata: { __typename?: 'Metadata', id: string, name: string, version: number, displayName?: string | null, icon?: string | null } } | null };

export type GetSourceTypeDisplayInfoQueryVariables = Exact<{
  name: Scalars['String']['input'];
}>;


export type GetSourceTypeDisplayInfoQuery = { __typename?: 'Query', sourceType?: { __typename?: 'SourceType', metadata: { __typename?: 'Metadata', id: string, name: string, version: number, displayName?: string | null, icon?: string | null } } | null };

export type ClearAgentUpgradeErrorMutationVariables = Exact<{
  input: ClearAgentUpgradeErrorInput;
}>;


export type ClearAgentUpgradeErrorMutation = { __typename?: 'Mutation', clearAgentUpgradeError?: boolean | null };

export type AgentChangesSubscriptionVariables = Exact<{
  selector?: InputMaybe<Scalars['String']['input']>;
  query?: InputMaybe<Scalars['String']['input']>;
}>;


export type AgentChangesSubscription = { __typename?: 'Subscription', agentChanges: Array<{ __typename?: 'AgentChange', changeType: AgentChangeType, agent: { __typename?: 'Agent', id: string, name: string, architecture?: string | null, operatingSystem?: string | null, labels?: any | null, hostName?: string | null, platform?: string | null, version?: string | null, macAddress?: string | null, home?: string | null, type?: string | null, status: number, connectedAt?: any | null, disconnectedAt?: any | null, configuration?: { __typename?: 'AgentConfiguration', Collector?: string | null } | null, configurationResource?: { __typename?: 'Configuration', metadata: { __typename?: 'Metadata', id: string, name: string, version: number } } | null } }> };

export type GetAgentAndConfigurationsQueryVariables = Exact<{
  agentId: Scalars['ID']['input'];
}>;


export type GetAgentAndConfigurationsQuery = { __typename?: 'Query', agent?: { __typename?: 'Agent', id: string, name: string, architecture?: string | null, operatingSystem?: string | null, labels?: any | null, hostName?: string | null, platform?: string | null, version?: string | null, macAddress?: string | null, remoteAddress?: string | null, home?: string | null, status: number, connectedAt?: any | null, disconnectedAt?: any | null, errorMessage?: string | null, upgradeAvailable?: string | null, features: number, configuration?: { __typename?: 'AgentConfiguration', Collector?: string | null } | null, configurationResource?: { __typename?: 'Configuration', metadata: { __typename?: 'Metadata', id: string, version: number, name: string } } | null, upgrade?: { __typename?: 'AgentUpgrade', status: number, version: string, error?: string | null } | null } | null, configurations: { __typename?: 'Configurations', configurations: Array<{ __typename?: 'Configuration', metadata: { __typename?: 'Metadata', id: string, name: string, version: number, labels?: any | null }, spec: { __typename?: 'ConfigurationSpec', raw?: string | null } }> } };

export type GetConfigurationNamesQueryVariables = Exact<{ [key: string]: never; }>;


export type GetConfigurationNamesQuery = { __typename?: 'Query', configurations: { __typename?: 'Configurations', configurations: Array<{ __typename?: 'Configuration', metadata: { __typename?: 'Metadata', id: string, name: string, version: number, labels?: any | null } }> } };

export type GetConfigRolloutAgentsQueryVariables = Exact<{
  name: Scalars['String']['input'];
}>;


export type GetConfigRolloutAgentsQuery = { __typename?: 'Query', configuration?: { __typename?: 'Configuration', agentCount?: number | null, metadata: { __typename?: 'Metadata', name: string, id: string, version: number } } | null };

export type GetRenderedConfigValueQueryVariables = Exact<{
  name: Scalars['String']['input'];
}>;


export type GetRenderedConfigValueQuery = { __typename?: 'Query', configuration?: { __typename?: 'Configuration', rendered?: string | null, metadata: { __typename?: 'Metadata', name: string, id: string, version: number } } | null };

export type GetConfigurationQueryVariables = Exact<{
  name: Scalars['String']['input'];
}>;


export type GetConfigurationQuery = { __typename?: 'Query', configuration?: { __typename?: 'Configuration', agentCount?: number | null, metadata: { __typename?: 'Metadata', id: string, name: string, description?: string | null, labels?: any | null, version: number }, spec: { __typename?: 'ConfigurationSpec', raw?: string | null, sources?: Array<{ __typename?: 'ResourceConfiguration', type?: string | null, name?: string | null, displayName?: string | null, disabled: boolean, parameters?: Array<{ __typename?: 'Parameter', name: string, value: any }> | null, processors?: Array<{ __typename?: 'ResourceConfiguration', type?: string | null, displayName?: string | null, disabled: boolean, parameters?: Array<{ __typename?: 'Parameter', name: string, value: any }> | null }> | null }> | null, destinations?: Array<{ __typename?: 'ResourceConfiguration', type?: string | null, name?: string | null, displayName?: string | null, disabled: boolean, parameters?: Array<{ __typename?: 'Parameter', name: string, value: any }> | null, processors?: Array<{ __typename?: 'ResourceConfiguration', type?: string | null, displayName?: string | null, disabled: boolean, parameters?: Array<{ __typename?: 'Parameter', name: string, value: any }> | null }> | null }> | null, selector?: { __typename?: 'AgentSelector', matchLabels?: any | null } | null }, graph?: { __typename?: 'Graph', attributes: any, sources: Array<{ __typename?: 'Node', id: string, type: string, label: string, attributes: any }>, intermediates: Array<{ __typename?: 'Node', id: string, type: string, label: string, attributes: any }>, targets: Array<{ __typename?: 'Node', id: string, type: string, label: string, attributes: any }>, edges: Array<{ __typename?: 'Edge', id: string, source: string, target: string }> } | null } | null };

export type DestinationsAndTypesQueryVariables = Exact<{ [key: string]: never; }>;


export type DestinationsAndTypesQuery = { __typename?: 'Query', destinationTypes: Array<{ __typename?: 'DestinationType', kind: string, apiVersion: string, metadata: { __typename?: 'Metadata', id: string, version: number, name: string, displayName?: string | null, description?: string | null, icon?: string | null }, spec: { __typename?: 'ResourceTypeSpec', version: string, supportedPlatforms: Array<string>, telemetryTypes: Array<PipelineType>, parameters: Array<{ __typename?: 'ParameterDefinition', label: string, type: ParameterType, name: string, description: string, default?: any | null, validValues?: Array<string> | null, advancedConfig?: boolean | null, required: boolean, relevantIf?: Array<{ __typename?: 'RelevantIfCondition', name: string, value: any, operator: RelevantIfOperatorType }> | null, documentation?: Array<{ __typename?: 'DocumentationLink', text: string, url: string }> | null, options: { __typename?: 'ParameterOptions', creatable?: boolean | null, multiline?: boolean | null, trackUnchecked?: boolean | null, sectionHeader?: boolean | null, gridColumns?: number | null, labels?: any | null, password?: boolean | null, metricCategories?: Array<{ __typename?: 'MetricCategory', label: string, column: number, metrics: Array<{ __typename?: 'MetricOption', name: string, description?: string | null, kpi?: boolean | null }> }> | null } }> } }>, destinations: Array<{ __typename?: 'Destination', metadata: { __typename?: 'Metadata', id: string, version: number, name: string }, spec: { __typename?: 'ParameterizedSpec', type: string, disabled: boolean, parameters?: Array<{ __typename?: 'Parameter', name: string, value: any }> | null } }> };

export type SourceTypesQueryVariables = Exact<{ [key: string]: never; }>;


export type SourceTypesQuery = { __typename?: 'Query', sourceTypes: Array<{ __typename?: 'SourceType', apiVersion: string, kind: string, metadata: { __typename?: 'Metadata', id: string, name: string, version: number, displayName?: string | null, description?: string | null, icon?: string | null }, spec: { __typename?: 'ResourceTypeSpec', supportedPlatforms: Array<string>, version: string, telemetryTypes: Array<PipelineType>, parameters: Array<{ __typename?: 'ParameterDefinition', name: string, label: string, description: string, advancedConfig?: boolean | null, required: boolean, type: ParameterType, validValues?: Array<string> | null, default?: any | null, relevantIf?: Array<{ __typename?: 'RelevantIfCondition', name: string, operator: RelevantIfOperatorType, value: any }> | null, documentation?: Array<{ __typename?: 'DocumentationLink', text: string, url: string }> | null, options: { __typename?: 'ParameterOptions', creatable?: boolean | null, multiline?: boolean | null, trackUnchecked?: boolean | null, sectionHeader?: boolean | null, gridColumns?: number | null, labels?: any | null, password?: boolean | null, metricCategories?: Array<{ __typename?: 'MetricCategory', label: string, column: number, metrics: Array<{ __typename?: 'MetricOption', name: string, description?: string | null, kpi?: boolean | null }> }> | null } }> } }> };

export type GetConfigNamesQueryVariables = Exact<{ [key: string]: never; }>;


export type GetConfigNamesQuery = { __typename?: 'Query', configurations: { __typename?: 'Configurations', configurations: Array<{ __typename?: 'Configuration', metadata: { __typename?: 'Metadata', id: string, name: string, version: number } }> } };

export type DestinationsQueryVariables = Exact<{ [key: string]: never; }>;


export type DestinationsQuery = { __typename?: 'Query', destinations: Array<{ __typename?: 'Destination', kind: string, metadata: { __typename?: 'Metadata', id: string, name: string, version: number }, spec: { __typename?: 'ParameterizedSpec', type: string } }> };

export type GetOverviewPageQueryVariables = Exact<{
  configIDs?: InputMaybe<Array<Scalars['ID']['input']> | Scalars['ID']['input']>;
  destinationIDs?: InputMaybe<Array<Scalars['ID']['input']> | Scalars['ID']['input']>;
  period: Scalars['String']['input'];
  telemetryType: Scalars['String']['input'];
}>;


export type GetOverviewPageQuery = { __typename?: 'Query', overviewPage: { __typename?: 'OverviewPage', graph: { __typename?: 'Graph', attributes: any, sources: Array<{ __typename?: 'Node', id: string, label: string, type: string, attributes: any }>, intermediates: Array<{ __typename?: 'Node', id: string, label: string, type: string, attributes: any }>, targets: Array<{ __typename?: 'Node', id: string, label: string, type: string, attributes: any }>, edges: Array<{ __typename?: 'Edge', id: string, source: string, target: string }> } } };

export type OverviewMetricsSubscriptionVariables = Exact<{
  period: Scalars['String']['input'];
  configIDs?: InputMaybe<Array<Scalars['ID']['input']> | Scalars['ID']['input']>;
  destinationIDs?: InputMaybe<Array<Scalars['ID']['input']> | Scalars['ID']['input']>;
}>;


export type OverviewMetricsSubscription = { __typename?: 'Subscription', overviewMetrics: { __typename?: 'GraphMetrics', maxMetricValue: number, maxLogValue: number, maxTraceValue: number, metrics: Array<{ __typename?: 'GraphMetric', name: string, nodeID: string, pipelineType: string, value: number, unit: string }> } };

export type DestinationsInConfigsQueryVariables = Exact<{ [key: string]: never; }>;


export type DestinationsInConfigsQuery = { __typename?: 'Query', destinationsInConfigs: Array<{ __typename?: 'Destination', kind: string, metadata: { __typename?: 'Metadata', id: string, version: number, name: string }, spec: { __typename?: 'ParameterizedSpec', type: string } }> };


export const GetLatestConfigVersionDocument = gql`
    query getLatestConfigVersion($name: String!) {
  configuration(name: $name) {
    metadata {
      id
      name
      version
    }
  }
}
    `;

/**
 * __useGetLatestConfigVersionQuery__
 *
 * To run a query within a React component, call `useGetLatestConfigVersionQuery` and pass it any options that fit your needs.
 * When your component renders, `useGetLatestConfigVersionQuery` returns an object from Apollo Client that contains loading, error, and data properties
 * you can use to render your UI.
 *
 * @param baseOptions options that will be passed into the query, supported options are listed on: https://www.apollographql.com/docs/react/api/react-hooks/#options;
 *
 * @example
 * const { data, loading, error } = useGetLatestConfigVersionQuery({
 *   variables: {
 *      name: // value for 'name'
 *   },
 * });
 */
export function useGetLatestConfigVersionQuery(baseOptions: Apollo.QueryHookOptions<GetLatestConfigVersionQuery, GetLatestConfigVersionQueryVariables>) {
        const options = {...defaultOptions, ...baseOptions}
        return Apollo.useQuery<GetLatestConfigVersionQuery, GetLatestConfigVersionQueryVariables>(GetLatestConfigVersionDocument, options);
      }
export function useGetLatestConfigVersionLazyQuery(baseOptions?: Apollo.LazyQueryHookOptions<GetLatestConfigVersionQuery, GetLatestConfigVersionQueryVariables>) {
          const options = {...defaultOptions, ...baseOptions}
          return Apollo.useLazyQuery<GetLatestConfigVersionQuery, GetLatestConfigVersionQueryVariables>(GetLatestConfigVersionDocument, options);
        }
export type GetLatestConfigVersionQueryHookResult = ReturnType<typeof useGetLatestConfigVersionQuery>;
export type GetLatestConfigVersionLazyQueryHookResult = ReturnType<typeof useGetLatestConfigVersionLazyQuery>;
export type GetLatestConfigVersionQueryResult = Apollo.QueryResult<GetLatestConfigVersionQuery, GetLatestConfigVersionQueryVariables>;
export const GetRenderedConfigDocument = gql`
    query getRenderedConfig($name: String!) {
  configuration(name: $name) {
    metadata {
      name
      id
      version
    }
    rendered
  }
}
    `;

/**
 * __useGetRenderedConfigQuery__
 *
 * To run a query within a React component, call `useGetRenderedConfigQuery` and pass it any options that fit your needs.
 * When your component renders, `useGetRenderedConfigQuery` returns an object from Apollo Client that contains loading, error, and data properties
 * you can use to render your UI.
 *
 * @param baseOptions options that will be passed into the query, supported options are listed on: https://www.apollographql.com/docs/react/api/react-hooks/#options;
 *
 * @example
 * const { data, loading, error } = useGetRenderedConfigQuery({
 *   variables: {
 *      name: // value for 'name'
 *   },
 * });
 */
export function useGetRenderedConfigQuery(baseOptions: Apollo.QueryHookOptions<GetRenderedConfigQuery, GetRenderedConfigQueryVariables>) {
        const options = {...defaultOptions, ...baseOptions}
        return Apollo.useQuery<GetRenderedConfigQuery, GetRenderedConfigQueryVariables>(GetRenderedConfigDocument, options);
      }
export function useGetRenderedConfigLazyQuery(baseOptions?: Apollo.LazyQueryHookOptions<GetRenderedConfigQuery, GetRenderedConfigQueryVariables>) {
          const options = {...defaultOptions, ...baseOptions}
          return Apollo.useLazyQuery<GetRenderedConfigQuery, GetRenderedConfigQueryVariables>(GetRenderedConfigDocument, options);
        }
export type GetRenderedConfigQueryHookResult = ReturnType<typeof useGetRenderedConfigQuery>;
export type GetRenderedConfigLazyQueryHookResult = ReturnType<typeof useGetRenderedConfigLazyQuery>;
export type GetRenderedConfigQueryResult = Apollo.QueryResult<GetRenderedConfigQuery, GetRenderedConfigQueryVariables>;
export const SourceTypeDocument = gql`
    query SourceType($name: String!) {
  sourceType(name: $name) {
    metadata {
      id
      name
      version
      displayName
      icon
      displayName
      description
    }
    spec {
      parameters {
        label
        name
        description
        required
        type
        default
        documentation {
          text
          url
        }
        relevantIf {
          name
          operator
          value
        }
        advancedConfig
        validValues
        options {
          creatable
          trackUnchecked
          sectionHeader
          gridColumns
          labels
          metricCategories {
            label
            column
            metrics {
              name
              description
              kpi
            }
          }
          password
        }
      }
    }
  }
}
    `;

/**
 * __useSourceTypeQuery__
 *
 * To run a query within a React component, call `useSourceTypeQuery` and pass it any options that fit your needs.
 * When your component renders, `useSourceTypeQuery` returns an object from Apollo Client that contains loading, error, and data properties
 * you can use to render your UI.
 *
 * @param baseOptions options that will be passed into the query, supported options are listed on: https://www.apollographql.com/docs/react/api/react-hooks/#options;
 *
 * @example
 * const { data, loading, error } = useSourceTypeQuery({
 *   variables: {
 *      name: // value for 'name'
 *   },
 * });
 */
export function useSourceTypeQuery(baseOptions: Apollo.QueryHookOptions<SourceTypeQuery, SourceTypeQueryVariables>) {
        const options = {...defaultOptions, ...baseOptions}
        return Apollo.useQuery<SourceTypeQuery, SourceTypeQueryVariables>(SourceTypeDocument, options);
      }
export function useSourceTypeLazyQuery(baseOptions?: Apollo.LazyQueryHookOptions<SourceTypeQuery, SourceTypeQueryVariables>) {
          const options = {...defaultOptions, ...baseOptions}
          return Apollo.useLazyQuery<SourceTypeQuery, SourceTypeQueryVariables>(SourceTypeDocument, options);
        }
export type SourceTypeQueryHookResult = ReturnType<typeof useSourceTypeQuery>;
export type SourceTypeLazyQueryHookResult = ReturnType<typeof useSourceTypeLazyQuery>;
export type SourceTypeQueryResult = Apollo.QueryResult<SourceTypeQuery, SourceTypeQueryVariables>;
export const GetDestinationWithTypeDocument = gql`
    query getDestinationWithType($name: String!) {
  destinationWithType(name: $name) {
    destination {
      metadata {
        name
        version
        id
        labels
        version
      }
      spec {
        type
        parameters {
          name
          value
        }
        disabled
      }
    }
    destinationType {
      metadata {
        id
        name
        version
        icon
        description
      }
      spec {
        parameters {
          label
          name
          description
          required
          type
          default
          relevantIf {
            name
            operator
            value
          }
          documentation {
            text
            url
          }
          advancedConfig
          validValues
          options {
            multiline
            creatable
            trackUnchecked
            sectionHeader
            gridColumns
            labels
            metricCategories {
              label
              column
              metrics {
                name
                description
                kpi
              }
            }
            password
          }
        }
      }
    }
  }
}
    `;

/**
 * __useGetDestinationWithTypeQuery__
 *
 * To run a query within a React component, call `useGetDestinationWithTypeQuery` and pass it any options that fit your needs.
 * When your component renders, `useGetDestinationWithTypeQuery` returns an object from Apollo Client that contains loading, error, and data properties
 * you can use to render your UI.
 *
 * @param baseOptions options that will be passed into the query, supported options are listed on: https://www.apollographql.com/docs/react/api/react-hooks/#options;
 *
 * @example
 * const { data, loading, error } = useGetDestinationWithTypeQuery({
 *   variables: {
 *      name: // value for 'name'
 *   },
 * });
 */
export function useGetDestinationWithTypeQuery(baseOptions: Apollo.QueryHookOptions<GetDestinationWithTypeQuery, GetDestinationWithTypeQueryVariables>) {
        const options = {...defaultOptions, ...baseOptions}
        return Apollo.useQuery<GetDestinationWithTypeQuery, GetDestinationWithTypeQueryVariables>(GetDestinationWithTypeDocument, options);
      }
export function useGetDestinationWithTypeLazyQuery(baseOptions?: Apollo.LazyQueryHookOptions<GetDestinationWithTypeQuery, GetDestinationWithTypeQueryVariables>) {
          const options = {...defaultOptions, ...baseOptions}
          return Apollo.useLazyQuery<GetDestinationWithTypeQuery, GetDestinationWithTypeQueryVariables>(GetDestinationWithTypeDocument, options);
        }
export type GetDestinationWithTypeQueryHookResult = ReturnType<typeof useGetDestinationWithTypeQuery>;
export type GetDestinationWithTypeLazyQueryHookResult = ReturnType<typeof useGetDestinationWithTypeLazyQuery>;
export type GetDestinationWithTypeQueryResult = Apollo.QueryResult<GetDestinationWithTypeQuery, GetDestinationWithTypeQueryVariables>;
export const GetCurrentConfigVersionDocument = gql`
    query getCurrentConfigVersion($configurationName: String!) {
  configuration(name: $configurationName) {
    metadata {
      id
      name
      version
      labels
    }
    agentCount
  }
}
    `;

/**
 * __useGetCurrentConfigVersionQuery__
 *
 * To run a query within a React component, call `useGetCurrentConfigVersionQuery` and pass it any options that fit your needs.
 * When your component renders, `useGetCurrentConfigVersionQuery` returns an object from Apollo Client that contains loading, error, and data properties
 * you can use to render your UI.
 *
 * @param baseOptions options that will be passed into the query, supported options are listed on: https://www.apollographql.com/docs/react/api/react-hooks/#options;
 *
 * @example
 * const { data, loading, error } = useGetCurrentConfigVersionQuery({
 *   variables: {
 *      configurationName: // value for 'configurationName'
 *   },
 * });
 */
export function useGetCurrentConfigVersionQuery(baseOptions: Apollo.QueryHookOptions<GetCurrentConfigVersionQuery, GetCurrentConfigVersionQueryVariables>) {
        const options = {...defaultOptions, ...baseOptions}
        return Apollo.useQuery<GetCurrentConfigVersionQuery, GetCurrentConfigVersionQueryVariables>(GetCurrentConfigVersionDocument, options);
      }
export function useGetCurrentConfigVersionLazyQuery(baseOptions?: Apollo.LazyQueryHookOptions<GetCurrentConfigVersionQuery, GetCurrentConfigVersionQueryVariables>) {
          const options = {...defaultOptions, ...baseOptions}
          return Apollo.useLazyQuery<GetCurrentConfigVersionQuery, GetCurrentConfigVersionQueryVariables>(GetCurrentConfigVersionDocument, options);
        }
export type GetCurrentConfigVersionQueryHookResult = ReturnType<typeof useGetCurrentConfigVersionQuery>;
export type GetCurrentConfigVersionLazyQueryHookResult = ReturnType<typeof useGetCurrentConfigVersionLazyQuery>;
export type GetCurrentConfigVersionQueryResult = Apollo.QueryResult<GetCurrentConfigVersionQuery, GetCurrentConfigVersionQueryVariables>;
export const GetLatestConfigDescriptionDocument = gql`
    query getLatestConfigDescription($configurationName: String!) {
  configuration(name: $configurationName) {
    metadata {
      id
      name
      version
      description
    }
  }
}
    `;

/**
 * __useGetLatestConfigDescriptionQuery__
 *
 * To run a query within a React component, call `useGetLatestConfigDescriptionQuery` and pass it any options that fit your needs.
 * When your component renders, `useGetLatestConfigDescriptionQuery` returns an object from Apollo Client that contains loading, error, and data properties
 * you can use to render your UI.
 *
 * @param baseOptions options that will be passed into the query, supported options are listed on: https://www.apollographql.com/docs/react/api/react-hooks/#options;
 *
 * @example
 * const { data, loading, error } = useGetLatestConfigDescriptionQuery({
 *   variables: {
 *      configurationName: // value for 'configurationName'
 *   },
 * });
 */
export function useGetLatestConfigDescriptionQuery(baseOptions: Apollo.QueryHookOptions<GetLatestConfigDescriptionQuery, GetLatestConfigDescriptionQueryVariables>) {
        const options = {...defaultOptions, ...baseOptions}
        return Apollo.useQuery<GetLatestConfigDescriptionQuery, GetLatestConfigDescriptionQueryVariables>(GetLatestConfigDescriptionDocument, options);
      }
export function useGetLatestConfigDescriptionLazyQuery(baseOptions?: Apollo.LazyQueryHookOptions<GetLatestConfigDescriptionQuery, GetLatestConfigDescriptionQueryVariables>) {
          const options = {...defaultOptions, ...baseOptions}
          return Apollo.useLazyQuery<GetLatestConfigDescriptionQuery, GetLatestConfigDescriptionQueryVariables>(GetLatestConfigDescriptionDocument, options);
        }
export type GetLatestConfigDescriptionQueryHookResult = ReturnType<typeof useGetLatestConfigDescriptionQuery>;
export type GetLatestConfigDescriptionLazyQueryHookResult = ReturnType<typeof useGetLatestConfigDescriptionLazyQuery>;
export type GetLatestConfigDescriptionQueryResult = Apollo.QueryResult<GetLatestConfigDescriptionQuery, GetLatestConfigDescriptionQueryVariables>;
export const EditConfigDescriptionDocument = gql`
    mutation editConfigDescription($input: EditConfigurationDescriptionInput!) {
  editConfigurationDescription(input: $input)
}
    `;
export type EditConfigDescriptionMutationFn = Apollo.MutationFunction<EditConfigDescriptionMutation, EditConfigDescriptionMutationVariables>;

/**
 * __useEditConfigDescriptionMutation__
 *
 * To run a mutation, you first call `useEditConfigDescriptionMutation` within a React component and pass it any options that fit your needs.
 * When your component renders, `useEditConfigDescriptionMutation` returns a tuple that includes:
 * - A mutate function that you can call at any time to execute the mutation
 * - An object with fields that represent the current status of the mutation's execution
 *
 * @param baseOptions options that will be passed into the mutation, supported options are listed on: https://www.apollographql.com/docs/react/api/react-hooks/#options-2;
 *
 * @example
 * const [editConfigDescriptionMutation, { data, loading, error }] = useEditConfigDescriptionMutation({
 *   variables: {
 *      input: // value for 'input'
 *   },
 * });
 */
export function useEditConfigDescriptionMutation(baseOptions?: Apollo.MutationHookOptions<EditConfigDescriptionMutation, EditConfigDescriptionMutationVariables>) {
        const options = {...defaultOptions, ...baseOptions}
        return Apollo.useMutation<EditConfigDescriptionMutation, EditConfigDescriptionMutationVariables>(EditConfigDescriptionDocument, options);
      }
export type EditConfigDescriptionMutationHookResult = ReturnType<typeof useEditConfigDescriptionMutation>;
export type EditConfigDescriptionMutationResult = Apollo.MutationResult<EditConfigDescriptionMutation>;
export type EditConfigDescriptionMutationOptions = Apollo.BaseMutationOptions<EditConfigDescriptionMutation, EditConfigDescriptionMutationVariables>;
export const GetConfigurationVersionsDocument = gql`
    query getConfigurationVersions($name: String!) {
  configurationHistory(name: $name) {
    metadata {
      name
      id
      version
    }
    activeTypes
    status {
      current
      pending
      latest
    }
  }
}
    `;

/**
 * __useGetConfigurationVersionsQuery__
 *
 * To run a query within a React component, call `useGetConfigurationVersionsQuery` and pass it any options that fit your needs.
 * When your component renders, `useGetConfigurationVersionsQuery` returns an object from Apollo Client that contains loading, error, and data properties
 * you can use to render your UI.
 *
 * @param baseOptions options that will be passed into the query, supported options are listed on: https://www.apollographql.com/docs/react/api/react-hooks/#options;
 *
 * @example
 * const { data, loading, error } = useGetConfigurationVersionsQuery({
 *   variables: {
 *      name: // value for 'name'
 *   },
 * });
 */
export function useGetConfigurationVersionsQuery(baseOptions: Apollo.QueryHookOptions<GetConfigurationVersionsQuery, GetConfigurationVersionsQueryVariables>) {
        const options = {...defaultOptions, ...baseOptions}
        return Apollo.useQuery<GetConfigurationVersionsQuery, GetConfigurationVersionsQueryVariables>(GetConfigurationVersionsDocument, options);
      }
export function useGetConfigurationVersionsLazyQuery(baseOptions?: Apollo.LazyQueryHookOptions<GetConfigurationVersionsQuery, GetConfigurationVersionsQueryVariables>) {
          const options = {...defaultOptions, ...baseOptions}
          return Apollo.useLazyQuery<GetConfigurationVersionsQuery, GetConfigurationVersionsQueryVariables>(GetConfigurationVersionsDocument, options);
        }
export type GetConfigurationVersionsQueryHookResult = ReturnType<typeof useGetConfigurationVersionsQuery>;
export type GetConfigurationVersionsLazyQueryHookResult = ReturnType<typeof useGetConfigurationVersionsLazyQuery>;
export type GetConfigurationVersionsQueryResult = Apollo.QueryResult<GetConfigurationVersionsQuery, GetConfigurationVersionsQueryVariables>;
export const RemoveAgentConfigurationDocument = gql`
    mutation removeAgentConfiguration($input: RemoveAgentConfigurationInput!) {
  removeAgentConfiguration(input: $input) {
    id
    configuration {
      Collector
      Logging
      Manager
    }
  }
}
    `;
export type RemoveAgentConfigurationMutationFn = Apollo.MutationFunction<RemoveAgentConfigurationMutation, RemoveAgentConfigurationMutationVariables>;

/**
 * __useRemoveAgentConfigurationMutation__
 *
 * To run a mutation, you first call `useRemoveAgentConfigurationMutation` within a React component and pass it any options that fit your needs.
 * When your component renders, `useRemoveAgentConfigurationMutation` returns a tuple that includes:
 * - A mutate function that you can call at any time to execute the mutation
 * - An object with fields that represent the current status of the mutation's execution
 *
 * @param baseOptions options that will be passed into the mutation, supported options are listed on: https://www.apollographql.com/docs/react/api/react-hooks/#options-2;
 *
 * @example
 * const [removeAgentConfigurationMutation, { data, loading, error }] = useRemoveAgentConfigurationMutation({
 *   variables: {
 *      input: // value for 'input'
 *   },
 * });
 */
export function useRemoveAgentConfigurationMutation(baseOptions?: Apollo.MutationHookOptions<RemoveAgentConfigurationMutation, RemoveAgentConfigurationMutationVariables>) {
        const options = {...defaultOptions, ...baseOptions}
        return Apollo.useMutation<RemoveAgentConfigurationMutation, RemoveAgentConfigurationMutationVariables>(RemoveAgentConfigurationDocument, options);
      }
export type RemoveAgentConfigurationMutationHookResult = ReturnType<typeof useRemoveAgentConfigurationMutation>;
export type RemoveAgentConfigurationMutationResult = Apollo.MutationResult<RemoveAgentConfigurationMutation>;
export type RemoveAgentConfigurationMutationOptions = Apollo.BaseMutationOptions<RemoveAgentConfigurationMutation, RemoveAgentConfigurationMutationVariables>;
export const ConfigurationMetricsDocument = gql`
    subscription ConfigurationMetrics($period: String!, $name: String!, $agent: String) {
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

/**
 * __useConfigurationMetricsSubscription__
 *
 * To run a query within a React component, call `useConfigurationMetricsSubscription` and pass it any options that fit your needs.
 * When your component renders, `useConfigurationMetricsSubscription` returns an object from Apollo Client that contains loading, error, and data properties
 * you can use to render your UI.
 *
 * @param baseOptions options that will be passed into the subscription, supported options are listed on: https://www.apollographql.com/docs/react/api/react-hooks/#options;
 *
 * @example
 * const { data, loading, error } = useConfigurationMetricsSubscription({
 *   variables: {
 *      period: // value for 'period'
 *      name: // value for 'name'
 *      agent: // value for 'agent'
 *   },
 * });
 */
export function useConfigurationMetricsSubscription(baseOptions: Apollo.SubscriptionHookOptions<ConfigurationMetricsSubscription, ConfigurationMetricsSubscriptionVariables>) {
        const options = {...defaultOptions, ...baseOptions}
        return Apollo.useSubscription<ConfigurationMetricsSubscription, ConfigurationMetricsSubscriptionVariables>(ConfigurationMetricsDocument, options);
      }
export type ConfigurationMetricsSubscriptionHookResult = ReturnType<typeof useConfigurationMetricsSubscription>;
export type ConfigurationMetricsSubscriptionResult = Apollo.SubscriptionResult<ConfigurationMetricsSubscription>;
export const GetProcessorTypesDocument = gql`
    query getProcessorTypes {
  processorTypes {
    metadata {
      displayName
      description
      name
      labels
      version
      id
    }
    spec {
      parameters {
        label
        name
        description
        required
        type
        default
        relevantIf {
          name
          operator
          value
        }
        documentation {
          text
          url
        }
        advancedConfig
        validValues
        options {
          creatable
          trackUnchecked
          gridColumns
          sectionHeader
          multiline
          labels
          metricCategories {
            label
            column
            metrics {
              name
              description
              kpi
            }
          }
          password
        }
        documentation {
          text
          url
        }
      }
      telemetryTypes
    }
  }
}
    `;

/**
 * __useGetProcessorTypesQuery__
 *
 * To run a query within a React component, call `useGetProcessorTypesQuery` and pass it any options that fit your needs.
 * When your component renders, `useGetProcessorTypesQuery` returns an object from Apollo Client that contains loading, error, and data properties
 * you can use to render your UI.
 *
 * @param baseOptions options that will be passed into the query, supported options are listed on: https://www.apollographql.com/docs/react/api/react-hooks/#options;
 *
 * @example
 * const { data, loading, error } = useGetProcessorTypesQuery({
 *   variables: {
 *   },
 * });
 */
export function useGetProcessorTypesQuery(baseOptions?: Apollo.QueryHookOptions<GetProcessorTypesQuery, GetProcessorTypesQueryVariables>) {
        const options = {...defaultOptions, ...baseOptions}
        return Apollo.useQuery<GetProcessorTypesQuery, GetProcessorTypesQueryVariables>(GetProcessorTypesDocument, options);
      }
export function useGetProcessorTypesLazyQuery(baseOptions?: Apollo.LazyQueryHookOptions<GetProcessorTypesQuery, GetProcessorTypesQueryVariables>) {
          const options = {...defaultOptions, ...baseOptions}
          return Apollo.useLazyQuery<GetProcessorTypesQuery, GetProcessorTypesQueryVariables>(GetProcessorTypesDocument, options);
        }
export type GetProcessorTypesQueryHookResult = ReturnType<typeof useGetProcessorTypesQuery>;
export type GetProcessorTypesLazyQueryHookResult = ReturnType<typeof useGetProcessorTypesLazyQuery>;
export type GetProcessorTypesQueryResult = Apollo.QueryResult<GetProcessorTypesQuery, GetProcessorTypesQueryVariables>;
export const GetProcessorTypeDocument = gql`
    query getProcessorType($type: String!) {
  processorType(name: $type) {
    metadata {
      displayName
      name
      version
      id
      description
    }
    spec {
      parameters {
        label
        name
        description
        required
        type
        default
        relevantIf {
          name
          operator
          value
        }
        documentation {
          text
          url
        }
        advancedConfig
        options {
          creatable
          trackUnchecked
          gridColumns
          sectionHeader
          multiline
          labels
          metricCategories {
            label
            column
            metrics {
              name
              description
              kpi
            }
          }
          password
        }
        documentation {
          text
          url
        }
        validValues
      }
    }
  }
}
    `;

/**
 * __useGetProcessorTypeQuery__
 *
 * To run a query within a React component, call `useGetProcessorTypeQuery` and pass it any options that fit your needs.
 * When your component renders, `useGetProcessorTypeQuery` returns an object from Apollo Client that contains loading, error, and data properties
 * you can use to render your UI.
 *
 * @param baseOptions options that will be passed into the query, supported options are listed on: https://www.apollographql.com/docs/react/api/react-hooks/#options;
 *
 * @example
 * const { data, loading, error } = useGetProcessorTypeQuery({
 *   variables: {
 *      type: // value for 'type'
 *   },
 * });
 */
export function useGetProcessorTypeQuery(baseOptions: Apollo.QueryHookOptions<GetProcessorTypeQuery, GetProcessorTypeQueryVariables>) {
        const options = {...defaultOptions, ...baseOptions}
        return Apollo.useQuery<GetProcessorTypeQuery, GetProcessorTypeQueryVariables>(GetProcessorTypeDocument, options);
      }
export function useGetProcessorTypeLazyQuery(baseOptions?: Apollo.LazyQueryHookOptions<GetProcessorTypeQuery, GetProcessorTypeQueryVariables>) {
          const options = {...defaultOptions, ...baseOptions}
          return Apollo.useLazyQuery<GetProcessorTypeQuery, GetProcessorTypeQueryVariables>(GetProcessorTypeDocument, options);
        }
export type GetProcessorTypeQueryHookResult = ReturnType<typeof useGetProcessorTypeQuery>;
export type GetProcessorTypeLazyQueryHookResult = ReturnType<typeof useGetProcessorTypeLazyQuery>;
export type GetProcessorTypeQueryResult = Apollo.QueryResult<GetProcessorTypeQuery, GetProcessorTypeQueryVariables>;
export const ProcessorDialogSourceTypeDocument = gql`
    query processorDialogSourceType($name: String!) {
  sourceType(name: $name) {
    metadata {
      name
      id
      version
      displayName
      description
    }
    spec {
      telemetryTypes
    }
  }
}
    `;

/**
 * __useProcessorDialogSourceTypeQuery__
 *
 * To run a query within a React component, call `useProcessorDialogSourceTypeQuery` and pass it any options that fit your needs.
 * When your component renders, `useProcessorDialogSourceTypeQuery` returns an object from Apollo Client that contains loading, error, and data properties
 * you can use to render your UI.
 *
 * @param baseOptions options that will be passed into the query, supported options are listed on: https://www.apollographql.com/docs/react/api/react-hooks/#options;
 *
 * @example
 * const { data, loading, error } = useProcessorDialogSourceTypeQuery({
 *   variables: {
 *      name: // value for 'name'
 *   },
 * });
 */
export function useProcessorDialogSourceTypeQuery(baseOptions: Apollo.QueryHookOptions<ProcessorDialogSourceTypeQuery, ProcessorDialogSourceTypeQueryVariables>) {
        const options = {...defaultOptions, ...baseOptions}
        return Apollo.useQuery<ProcessorDialogSourceTypeQuery, ProcessorDialogSourceTypeQueryVariables>(ProcessorDialogSourceTypeDocument, options);
      }
export function useProcessorDialogSourceTypeLazyQuery(baseOptions?: Apollo.LazyQueryHookOptions<ProcessorDialogSourceTypeQuery, ProcessorDialogSourceTypeQueryVariables>) {
          const options = {...defaultOptions, ...baseOptions}
          return Apollo.useLazyQuery<ProcessorDialogSourceTypeQuery, ProcessorDialogSourceTypeQueryVariables>(ProcessorDialogSourceTypeDocument, options);
        }
export type ProcessorDialogSourceTypeQueryHookResult = ReturnType<typeof useProcessorDialogSourceTypeQuery>;
export type ProcessorDialogSourceTypeLazyQueryHookResult = ReturnType<typeof useProcessorDialogSourceTypeLazyQuery>;
export type ProcessorDialogSourceTypeQueryResult = Apollo.QueryResult<ProcessorDialogSourceTypeQuery, ProcessorDialogSourceTypeQueryVariables>;
export const ProcessorDialogDestinationTypeDocument = gql`
    query processorDialogDestinationType($name: String!) {
  destinationWithType(name: $name) {
    destinationType {
      metadata {
        id
        name
        version
        displayName
        description
      }
      spec {
        telemetryTypes
      }
    }
  }
}
    `;

/**
 * __useProcessorDialogDestinationTypeQuery__
 *
 * To run a query within a React component, call `useProcessorDialogDestinationTypeQuery` and pass it any options that fit your needs.
 * When your component renders, `useProcessorDialogDestinationTypeQuery` returns an object from Apollo Client that contains loading, error, and data properties
 * you can use to render your UI.
 *
 * @param baseOptions options that will be passed into the query, supported options are listed on: https://www.apollographql.com/docs/react/api/react-hooks/#options;
 *
 * @example
 * const { data, loading, error } = useProcessorDialogDestinationTypeQuery({
 *   variables: {
 *      name: // value for 'name'
 *   },
 * });
 */
export function useProcessorDialogDestinationTypeQuery(baseOptions: Apollo.QueryHookOptions<ProcessorDialogDestinationTypeQuery, ProcessorDialogDestinationTypeQueryVariables>) {
        const options = {...defaultOptions, ...baseOptions}
        return Apollo.useQuery<ProcessorDialogDestinationTypeQuery, ProcessorDialogDestinationTypeQueryVariables>(ProcessorDialogDestinationTypeDocument, options);
      }
export function useProcessorDialogDestinationTypeLazyQuery(baseOptions?: Apollo.LazyQueryHookOptions<ProcessorDialogDestinationTypeQuery, ProcessorDialogDestinationTypeQueryVariables>) {
          const options = {...defaultOptions, ...baseOptions}
          return Apollo.useLazyQuery<ProcessorDialogDestinationTypeQuery, ProcessorDialogDestinationTypeQueryVariables>(ProcessorDialogDestinationTypeDocument, options);
        }
export type ProcessorDialogDestinationTypeQueryHookResult = ReturnType<typeof useProcessorDialogDestinationTypeQuery>;
export type ProcessorDialogDestinationTypeLazyQueryHookResult = ReturnType<typeof useProcessorDialogDestinationTypeLazyQuery>;
export type ProcessorDialogDestinationTypeQueryResult = Apollo.QueryResult<ProcessorDialogDestinationTypeQuery, ProcessorDialogDestinationTypeQueryVariables>;
export const UpdateProcessorsDocument = gql`
    mutation updateProcessors($input: UpdateProcessorsInput!) {
  updateProcessors(input: $input)
}
    `;
export type UpdateProcessorsMutationFn = Apollo.MutationFunction<UpdateProcessorsMutation, UpdateProcessorsMutationVariables>;

/**
 * __useUpdateProcessorsMutation__
 *
 * To run a mutation, you first call `useUpdateProcessorsMutation` within a React component and pass it any options that fit your needs.
 * When your component renders, `useUpdateProcessorsMutation` returns a tuple that includes:
 * - A mutate function that you can call at any time to execute the mutation
 * - An object with fields that represent the current status of the mutation's execution
 *
 * @param baseOptions options that will be passed into the mutation, supported options are listed on: https://www.apollographql.com/docs/react/api/react-hooks/#options-2;
 *
 * @example
 * const [updateProcessorsMutation, { data, loading, error }] = useUpdateProcessorsMutation({
 *   variables: {
 *      input: // value for 'input'
 *   },
 * });
 */
export function useUpdateProcessorsMutation(baseOptions?: Apollo.MutationHookOptions<UpdateProcessorsMutation, UpdateProcessorsMutationVariables>) {
        const options = {...defaultOptions, ...baseOptions}
        return Apollo.useMutation<UpdateProcessorsMutation, UpdateProcessorsMutationVariables>(UpdateProcessorsDocument, options);
      }
export type UpdateProcessorsMutationHookResult = ReturnType<typeof useUpdateProcessorsMutation>;
export type UpdateProcessorsMutationResult = Apollo.MutationResult<UpdateProcessorsMutation>;
export type UpdateProcessorsMutationOptions = Apollo.BaseMutationOptions<UpdateProcessorsMutation, UpdateProcessorsMutationVariables>;
export const GetRolloutHistoryDocument = gql`
    query getRolloutHistory($name: String!) {
  configurationHistory(name: $name) {
    metadata {
      name
      id
      version
      dateModified
    }
    status {
      rollout {
        status
        errors
      }
    }
  }
}
    `;

/**
 * __useGetRolloutHistoryQuery__
 *
 * To run a query within a React component, call `useGetRolloutHistoryQuery` and pass it any options that fit your needs.
 * When your component renders, `useGetRolloutHistoryQuery` returns an object from Apollo Client that contains loading, error, and data properties
 * you can use to render your UI.
 *
 * @param baseOptions options that will be passed into the query, supported options are listed on: https://www.apollographql.com/docs/react/api/react-hooks/#options;
 *
 * @example
 * const { data, loading, error } = useGetRolloutHistoryQuery({
 *   variables: {
 *      name: // value for 'name'
 *   },
 * });
 */
export function useGetRolloutHistoryQuery(baseOptions: Apollo.QueryHookOptions<GetRolloutHistoryQuery, GetRolloutHistoryQueryVariables>) {
        const options = {...defaultOptions, ...baseOptions}
        return Apollo.useQuery<GetRolloutHistoryQuery, GetRolloutHistoryQueryVariables>(GetRolloutHistoryDocument, options);
      }
export function useGetRolloutHistoryLazyQuery(baseOptions?: Apollo.LazyQueryHookOptions<GetRolloutHistoryQuery, GetRolloutHistoryQueryVariables>) {
          const options = {...defaultOptions, ...baseOptions}
          return Apollo.useLazyQuery<GetRolloutHistoryQuery, GetRolloutHistoryQueryVariables>(GetRolloutHistoryDocument, options);
        }
export type GetRolloutHistoryQueryHookResult = ReturnType<typeof useGetRolloutHistoryQuery>;
export type GetRolloutHistoryLazyQueryHookResult = ReturnType<typeof useGetRolloutHistoryLazyQuery>;
export type GetRolloutHistoryQueryResult = Apollo.QueryResult<GetRolloutHistoryQuery, GetRolloutHistoryQueryVariables>;
export const GetConfigRolloutStatusDocument = gql`
    query getConfigRolloutStatus($name: String!) {
  configuration(name: $name) {
    metadata {
      name
      id
      version
    }
    agentCount
    status {
      pending
      current
      latest
      rollout {
        status
        phase
        completed
        errors
        pending
        waiting
      }
    }
  }
}
    `;

/**
 * __useGetConfigRolloutStatusQuery__
 *
 * To run a query within a React component, call `useGetConfigRolloutStatusQuery` and pass it any options that fit your needs.
 * When your component renders, `useGetConfigRolloutStatusQuery` returns an object from Apollo Client that contains loading, error, and data properties
 * you can use to render your UI.
 *
 * @param baseOptions options that will be passed into the query, supported options are listed on: https://www.apollographql.com/docs/react/api/react-hooks/#options;
 *
 * @example
 * const { data, loading, error } = useGetConfigRolloutStatusQuery({
 *   variables: {
 *      name: // value for 'name'
 *   },
 * });
 */
export function useGetConfigRolloutStatusQuery(baseOptions: Apollo.QueryHookOptions<GetConfigRolloutStatusQuery, GetConfigRolloutStatusQueryVariables>) {
        const options = {...defaultOptions, ...baseOptions}
        return Apollo.useQuery<GetConfigRolloutStatusQuery, GetConfigRolloutStatusQueryVariables>(GetConfigRolloutStatusDocument, options);
      }
export function useGetConfigRolloutStatusLazyQuery(baseOptions?: Apollo.LazyQueryHookOptions<GetConfigRolloutStatusQuery, GetConfigRolloutStatusQueryVariables>) {
          const options = {...defaultOptions, ...baseOptions}
          return Apollo.useLazyQuery<GetConfigRolloutStatusQuery, GetConfigRolloutStatusQueryVariables>(GetConfigRolloutStatusDocument, options);
        }
export type GetConfigRolloutStatusQueryHookResult = ReturnType<typeof useGetConfigRolloutStatusQuery>;
export type GetConfigRolloutStatusLazyQueryHookResult = ReturnType<typeof useGetConfigRolloutStatusLazyQuery>;
export type GetConfigRolloutStatusQueryResult = Apollo.QueryResult<GetConfigRolloutStatusQuery, GetConfigRolloutStatusQueryVariables>;
export const AgentsWithConfigurationDocument = gql`
    query agentsWithConfiguration($selector: String, $query: String) {
  agents(selector: $selector, query: $query) {
    agents {
      id
      name
    }
  }
}
    `;

/**
 * __useAgentsWithConfigurationQuery__
 *
 * To run a query within a React component, call `useAgentsWithConfigurationQuery` and pass it any options that fit your needs.
 * When your component renders, `useAgentsWithConfigurationQuery` returns an object from Apollo Client that contains loading, error, and data properties
 * you can use to render your UI.
 *
 * @param baseOptions options that will be passed into the query, supported options are listed on: https://www.apollographql.com/docs/react/api/react-hooks/#options;
 *
 * @example
 * const { data, loading, error } = useAgentsWithConfigurationQuery({
 *   variables: {
 *      selector: // value for 'selector'
 *      query: // value for 'query'
 *   },
 * });
 */
export function useAgentsWithConfigurationQuery(baseOptions?: Apollo.QueryHookOptions<AgentsWithConfigurationQuery, AgentsWithConfigurationQueryVariables>) {
        const options = {...defaultOptions, ...baseOptions}
        return Apollo.useQuery<AgentsWithConfigurationQuery, AgentsWithConfigurationQueryVariables>(AgentsWithConfigurationDocument, options);
      }
export function useAgentsWithConfigurationLazyQuery(baseOptions?: Apollo.LazyQueryHookOptions<AgentsWithConfigurationQuery, AgentsWithConfigurationQueryVariables>) {
          const options = {...defaultOptions, ...baseOptions}
          return Apollo.useLazyQuery<AgentsWithConfigurationQuery, AgentsWithConfigurationQueryVariables>(AgentsWithConfigurationDocument, options);
        }
export type AgentsWithConfigurationQueryHookResult = ReturnType<typeof useAgentsWithConfigurationQuery>;
export type AgentsWithConfigurationLazyQueryHookResult = ReturnType<typeof useAgentsWithConfigurationLazyQuery>;
export type AgentsWithConfigurationQueryResult = Apollo.QueryResult<AgentsWithConfigurationQuery, AgentsWithConfigurationQueryVariables>;
export const SnapshotDocument = gql`
    query snapshot($agentID: String!, $pipelineType: PipelineType!, $position: String, $resourceName: String) {
  snapshot(
    agentID: $agentID
    pipelineType: $pipelineType
    position: $position
    resourceName: $resourceName
  ) {
    metrics {
      name
      timestamp
      value
      unit
      type
      attributes
      resource
    }
    logs {
      timestamp
      body
      severity
      attributes
      resource
    }
    traces {
      name
      traceID
      spanID
      parentSpanID
      start
      end
      attributes
      resource
    }
  }
}
    `;

/**
 * __useSnapshotQuery__
 *
 * To run a query within a React component, call `useSnapshotQuery` and pass it any options that fit your needs.
 * When your component renders, `useSnapshotQuery` returns an object from Apollo Client that contains loading, error, and data properties
 * you can use to render your UI.
 *
 * @param baseOptions options that will be passed into the query, supported options are listed on: https://www.apollographql.com/docs/react/api/react-hooks/#options;
 *
 * @example
 * const { data, loading, error } = useSnapshotQuery({
 *   variables: {
 *      agentID: // value for 'agentID'
 *      pipelineType: // value for 'pipelineType'
 *      position: // value for 'position'
 *      resourceName: // value for 'resourceName'
 *   },
 * });
 */
export function useSnapshotQuery(baseOptions: Apollo.QueryHookOptions<SnapshotQuery, SnapshotQueryVariables>) {
        const options = {...defaultOptions, ...baseOptions}
        return Apollo.useQuery<SnapshotQuery, SnapshotQueryVariables>(SnapshotDocument, options);
      }
export function useSnapshotLazyQuery(baseOptions?: Apollo.LazyQueryHookOptions<SnapshotQuery, SnapshotQueryVariables>) {
          const options = {...defaultOptions, ...baseOptions}
          return Apollo.useLazyQuery<SnapshotQuery, SnapshotQueryVariables>(SnapshotDocument, options);
        }
export type SnapshotQueryHookResult = ReturnType<typeof useSnapshotQuery>;
export type SnapshotLazyQueryHookResult = ReturnType<typeof useSnapshotLazyQuery>;
export type SnapshotQueryResult = Apollo.QueryResult<SnapshotQuery, SnapshotQueryVariables>;
export const AgentsTableDocument = gql`
    query AgentsTable($selector: String, $query: String) {
  agents(selector: $selector, query: $query) {
    agents {
      id
      architecture
      hostName
      labels
      platform
      version
      name
      home
      operatingSystem
      macAddress
      type
      status
      connectedAt
      disconnectedAt
      configurationResource {
        metadata {
          id
          name
          version
        }
      }
    }
    query
    suggestions {
      query
      label
    }
    latestVersion
  }
}
    `;

/**
 * __useAgentsTableQuery__
 *
 * To run a query within a React component, call `useAgentsTableQuery` and pass it any options that fit your needs.
 * When your component renders, `useAgentsTableQuery` returns an object from Apollo Client that contains loading, error, and data properties
 * you can use to render your UI.
 *
 * @param baseOptions options that will be passed into the query, supported options are listed on: https://www.apollographql.com/docs/react/api/react-hooks/#options;
 *
 * @example
 * const { data, loading, error } = useAgentsTableQuery({
 *   variables: {
 *      selector: // value for 'selector'
 *      query: // value for 'query'
 *   },
 * });
 */
export function useAgentsTableQuery(baseOptions?: Apollo.QueryHookOptions<AgentsTableQuery, AgentsTableQueryVariables>) {
        const options = {...defaultOptions, ...baseOptions}
        return Apollo.useQuery<AgentsTableQuery, AgentsTableQueryVariables>(AgentsTableDocument, options);
      }
export function useAgentsTableLazyQuery(baseOptions?: Apollo.LazyQueryHookOptions<AgentsTableQuery, AgentsTableQueryVariables>) {
          const options = {...defaultOptions, ...baseOptions}
          return Apollo.useLazyQuery<AgentsTableQuery, AgentsTableQueryVariables>(AgentsTableDocument, options);
        }
export type AgentsTableQueryHookResult = ReturnType<typeof useAgentsTableQuery>;
export type AgentsTableLazyQueryHookResult = ReturnType<typeof useAgentsTableLazyQuery>;
export type AgentsTableQueryResult = Apollo.QueryResult<AgentsTableQuery, AgentsTableQueryVariables>;
export const AgentsTableMetricsDocument = gql`
    subscription AgentsTableMetrics($period: String!, $ids: [ID!]) {
  agentMetrics(period: $period, ids: $ids) {
    metrics {
      name
      nodeID
      pipelineType
      value
      unit
      agentID
    }
  }
}
    `;

/**
 * __useAgentsTableMetricsSubscription__
 *
 * To run a query within a React component, call `useAgentsTableMetricsSubscription` and pass it any options that fit your needs.
 * When your component renders, `useAgentsTableMetricsSubscription` returns an object from Apollo Client that contains loading, error, and data properties
 * you can use to render your UI.
 *
 * @param baseOptions options that will be passed into the subscription, supported options are listed on: https://www.apollographql.com/docs/react/api/react-hooks/#options;
 *
 * @example
 * const { data, loading, error } = useAgentsTableMetricsSubscription({
 *   variables: {
 *      period: // value for 'period'
 *      ids: // value for 'ids'
 *   },
 * });
 */
export function useAgentsTableMetricsSubscription(baseOptions: Apollo.SubscriptionHookOptions<AgentsTableMetricsSubscription, AgentsTableMetricsSubscriptionVariables>) {
        const options = {...defaultOptions, ...baseOptions}
        return Apollo.useSubscription<AgentsTableMetricsSubscription, AgentsTableMetricsSubscriptionVariables>(AgentsTableMetricsDocument, options);
      }
export type AgentsTableMetricsSubscriptionHookResult = ReturnType<typeof useAgentsTableMetricsSubscription>;
export type AgentsTableMetricsSubscriptionResult = Apollo.SubscriptionResult<AgentsTableMetricsSubscription>;
export const GetConfigurationTableDocument = gql`
    query GetConfigurationTable($selector: String, $query: String, $onlyDeployedConfigurations: Boolean) {
  configurations(
    selector: $selector
    query: $query
    onlyDeployedConfigurations: $onlyDeployedConfigurations
  ) {
    configurations {
      metadata {
        id
        version
        name
        labels
        description
      }
      agentCount
    }
    query
    suggestions {
      query
      label
    }
  }
}
    `;

/**
 * __useGetConfigurationTableQuery__
 *
 * To run a query within a React component, call `useGetConfigurationTableQuery` and pass it any options that fit your needs.
 * When your component renders, `useGetConfigurationTableQuery` returns an object from Apollo Client that contains loading, error, and data properties
 * you can use to render your UI.
 *
 * @param baseOptions options that will be passed into the query, supported options are listed on: https://www.apollographql.com/docs/react/api/react-hooks/#options;
 *
 * @example
 * const { data, loading, error } = useGetConfigurationTableQuery({
 *   variables: {
 *      selector: // value for 'selector'
 *      query: // value for 'query'
 *      onlyDeployedConfigurations: // value for 'onlyDeployedConfigurations'
 *   },
 * });
 */
export function useGetConfigurationTableQuery(baseOptions?: Apollo.QueryHookOptions<GetConfigurationTableQuery, GetConfigurationTableQueryVariables>) {
        const options = {...defaultOptions, ...baseOptions}
        return Apollo.useQuery<GetConfigurationTableQuery, GetConfigurationTableQueryVariables>(GetConfigurationTableDocument, options);
      }
export function useGetConfigurationTableLazyQuery(baseOptions?: Apollo.LazyQueryHookOptions<GetConfigurationTableQuery, GetConfigurationTableQueryVariables>) {
          const options = {...defaultOptions, ...baseOptions}
          return Apollo.useLazyQuery<GetConfigurationTableQuery, GetConfigurationTableQueryVariables>(GetConfigurationTableDocument, options);
        }
export type GetConfigurationTableQueryHookResult = ReturnType<typeof useGetConfigurationTableQuery>;
export type GetConfigurationTableLazyQueryHookResult = ReturnType<typeof useGetConfigurationTableLazyQuery>;
export type GetConfigurationTableQueryResult = Apollo.QueryResult<GetConfigurationTableQuery, GetConfigurationTableQueryVariables>;
export const ConfigurationChangesDocument = gql`
    subscription ConfigurationChanges($selector: String, $query: String) {
  configurationChanges(selector: $selector, query: $query) {
    configuration {
      metadata {
        id
        version
        name
        description
        labels
      }
      agentCount
    }
    eventType
  }
}
    `;

/**
 * __useConfigurationChangesSubscription__
 *
 * To run a query within a React component, call `useConfigurationChangesSubscription` and pass it any options that fit your needs.
 * When your component renders, `useConfigurationChangesSubscription` returns an object from Apollo Client that contains loading, error, and data properties
 * you can use to render your UI.
 *
 * @param baseOptions options that will be passed into the subscription, supported options are listed on: https://www.apollographql.com/docs/react/api/react-hooks/#options;
 *
 * @example
 * const { data, loading, error } = useConfigurationChangesSubscription({
 *   variables: {
 *      selector: // value for 'selector'
 *      query: // value for 'query'
 *   },
 * });
 */
export function useConfigurationChangesSubscription(baseOptions?: Apollo.SubscriptionHookOptions<ConfigurationChangesSubscription, ConfigurationChangesSubscriptionVariables>) {
        const options = {...defaultOptions, ...baseOptions}
        return Apollo.useSubscription<ConfigurationChangesSubscription, ConfigurationChangesSubscriptionVariables>(ConfigurationChangesDocument, options);
      }
export type ConfigurationChangesSubscriptionHookResult = ReturnType<typeof useConfigurationChangesSubscription>;
export type ConfigurationChangesSubscriptionResult = Apollo.SubscriptionResult<ConfigurationChangesSubscription>;
export const ConfigurationTableMetricsDocument = gql`
    subscription ConfigurationTableMetrics($period: String!) {
  overviewMetrics(period: $period) {
    metrics {
      name
      nodeID
      pipelineType
      value
      unit
    }
  }
}
    `;

/**
 * __useConfigurationTableMetricsSubscription__
 *
 * To run a query within a React component, call `useConfigurationTableMetricsSubscription` and pass it any options that fit your needs.
 * When your component renders, `useConfigurationTableMetricsSubscription` returns an object from Apollo Client that contains loading, error, and data properties
 * you can use to render your UI.
 *
 * @param baseOptions options that will be passed into the subscription, supported options are listed on: https://www.apollographql.com/docs/react/api/react-hooks/#options;
 *
 * @example
 * const { data, loading, error } = useConfigurationTableMetricsSubscription({
 *   variables: {
 *      period: // value for 'period'
 *   },
 * });
 */
export function useConfigurationTableMetricsSubscription(baseOptions: Apollo.SubscriptionHookOptions<ConfigurationTableMetricsSubscription, ConfigurationTableMetricsSubscriptionVariables>) {
        const options = {...defaultOptions, ...baseOptions}
        return Apollo.useSubscription<ConfigurationTableMetricsSubscription, ConfigurationTableMetricsSubscriptionVariables>(ConfigurationTableMetricsDocument, options);
      }
export type ConfigurationTableMetricsSubscriptionHookResult = ReturnType<typeof useConfigurationTableMetricsSubscription>;
export type ConfigurationTableMetricsSubscriptionResult = Apollo.SubscriptionResult<ConfigurationTableMetricsSubscription>;
export const GetDestinationTypeDisplayInfoDocument = gql`
    query getDestinationTypeDisplayInfo($name: String!) {
  destinationType(name: $name) {
    metadata {
      id
      name
      version
      displayName
      icon
    }
  }
}
    `;

/**
 * __useGetDestinationTypeDisplayInfoQuery__
 *
 * To run a query within a React component, call `useGetDestinationTypeDisplayInfoQuery` and pass it any options that fit your needs.
 * When your component renders, `useGetDestinationTypeDisplayInfoQuery` returns an object from Apollo Client that contains loading, error, and data properties
 * you can use to render your UI.
 *
 * @param baseOptions options that will be passed into the query, supported options are listed on: https://www.apollographql.com/docs/react/api/react-hooks/#options;
 *
 * @example
 * const { data, loading, error } = useGetDestinationTypeDisplayInfoQuery({
 *   variables: {
 *      name: // value for 'name'
 *   },
 * });
 */
export function useGetDestinationTypeDisplayInfoQuery(baseOptions: Apollo.QueryHookOptions<GetDestinationTypeDisplayInfoQuery, GetDestinationTypeDisplayInfoQueryVariables>) {
        const options = {...defaultOptions, ...baseOptions}
        return Apollo.useQuery<GetDestinationTypeDisplayInfoQuery, GetDestinationTypeDisplayInfoQueryVariables>(GetDestinationTypeDisplayInfoDocument, options);
      }
export function useGetDestinationTypeDisplayInfoLazyQuery(baseOptions?: Apollo.LazyQueryHookOptions<GetDestinationTypeDisplayInfoQuery, GetDestinationTypeDisplayInfoQueryVariables>) {
          const options = {...defaultOptions, ...baseOptions}
          return Apollo.useLazyQuery<GetDestinationTypeDisplayInfoQuery, GetDestinationTypeDisplayInfoQueryVariables>(GetDestinationTypeDisplayInfoDocument, options);
        }
export type GetDestinationTypeDisplayInfoQueryHookResult = ReturnType<typeof useGetDestinationTypeDisplayInfoQuery>;
export type GetDestinationTypeDisplayInfoLazyQueryHookResult = ReturnType<typeof useGetDestinationTypeDisplayInfoLazyQuery>;
export type GetDestinationTypeDisplayInfoQueryResult = Apollo.QueryResult<GetDestinationTypeDisplayInfoQuery, GetDestinationTypeDisplayInfoQueryVariables>;
export const GetSourceTypeDisplayInfoDocument = gql`
    query getSourceTypeDisplayInfo($name: String!) {
  sourceType(name: $name) {
    metadata {
      id
      name
      version
      displayName
      icon
    }
  }
}
    `;

/**
 * __useGetSourceTypeDisplayInfoQuery__
 *
 * To run a query within a React component, call `useGetSourceTypeDisplayInfoQuery` and pass it any options that fit your needs.
 * When your component renders, `useGetSourceTypeDisplayInfoQuery` returns an object from Apollo Client that contains loading, error, and data properties
 * you can use to render your UI.
 *
 * @param baseOptions options that will be passed into the query, supported options are listed on: https://www.apollographql.com/docs/react/api/react-hooks/#options;
 *
 * @example
 * const { data, loading, error } = useGetSourceTypeDisplayInfoQuery({
 *   variables: {
 *      name: // value for 'name'
 *   },
 * });
 */
export function useGetSourceTypeDisplayInfoQuery(baseOptions: Apollo.QueryHookOptions<GetSourceTypeDisplayInfoQuery, GetSourceTypeDisplayInfoQueryVariables>) {
        const options = {...defaultOptions, ...baseOptions}
        return Apollo.useQuery<GetSourceTypeDisplayInfoQuery, GetSourceTypeDisplayInfoQueryVariables>(GetSourceTypeDisplayInfoDocument, options);
      }
export function useGetSourceTypeDisplayInfoLazyQuery(baseOptions?: Apollo.LazyQueryHookOptions<GetSourceTypeDisplayInfoQuery, GetSourceTypeDisplayInfoQueryVariables>) {
          const options = {...defaultOptions, ...baseOptions}
          return Apollo.useLazyQuery<GetSourceTypeDisplayInfoQuery, GetSourceTypeDisplayInfoQueryVariables>(GetSourceTypeDisplayInfoDocument, options);
        }
export type GetSourceTypeDisplayInfoQueryHookResult = ReturnType<typeof useGetSourceTypeDisplayInfoQuery>;
export type GetSourceTypeDisplayInfoLazyQueryHookResult = ReturnType<typeof useGetSourceTypeDisplayInfoLazyQuery>;
export type GetSourceTypeDisplayInfoQueryResult = Apollo.QueryResult<GetSourceTypeDisplayInfoQuery, GetSourceTypeDisplayInfoQueryVariables>;
export const ClearAgentUpgradeErrorDocument = gql`
    mutation ClearAgentUpgradeError($input: ClearAgentUpgradeErrorInput!) {
  clearAgentUpgradeError(input: $input)
}
    `;
export type ClearAgentUpgradeErrorMutationFn = Apollo.MutationFunction<ClearAgentUpgradeErrorMutation, ClearAgentUpgradeErrorMutationVariables>;

/**
 * __useClearAgentUpgradeErrorMutation__
 *
 * To run a mutation, you first call `useClearAgentUpgradeErrorMutation` within a React component and pass it any options that fit your needs.
 * When your component renders, `useClearAgentUpgradeErrorMutation` returns a tuple that includes:
 * - A mutate function that you can call at any time to execute the mutation
 * - An object with fields that represent the current status of the mutation's execution
 *
 * @param baseOptions options that will be passed into the mutation, supported options are listed on: https://www.apollographql.com/docs/react/api/react-hooks/#options-2;
 *
 * @example
 * const [clearAgentUpgradeErrorMutation, { data, loading, error }] = useClearAgentUpgradeErrorMutation({
 *   variables: {
 *      input: // value for 'input'
 *   },
 * });
 */
export function useClearAgentUpgradeErrorMutation(baseOptions?: Apollo.MutationHookOptions<ClearAgentUpgradeErrorMutation, ClearAgentUpgradeErrorMutationVariables>) {
        const options = {...defaultOptions, ...baseOptions}
        return Apollo.useMutation<ClearAgentUpgradeErrorMutation, ClearAgentUpgradeErrorMutationVariables>(ClearAgentUpgradeErrorDocument, options);
      }
export type ClearAgentUpgradeErrorMutationHookResult = ReturnType<typeof useClearAgentUpgradeErrorMutation>;
export type ClearAgentUpgradeErrorMutationResult = Apollo.MutationResult<ClearAgentUpgradeErrorMutation>;
export type ClearAgentUpgradeErrorMutationOptions = Apollo.BaseMutationOptions<ClearAgentUpgradeErrorMutation, ClearAgentUpgradeErrorMutationVariables>;
export const AgentChangesDocument = gql`
    subscription AgentChanges($selector: String, $query: String) {
  agentChanges(selector: $selector, query: $query) {
    agent {
      id
      name
      architecture
      operatingSystem
      labels
      hostName
      platform
      version
      macAddress
      home
      type
      status
      connectedAt
      disconnectedAt
      configuration {
        Collector
      }
      configurationResource {
        metadata {
          id
          name
          version
        }
      }
    }
    changeType
  }
}
    `;

/**
 * __useAgentChangesSubscription__
 *
 * To run a query within a React component, call `useAgentChangesSubscription` and pass it any options that fit your needs.
 * When your component renders, `useAgentChangesSubscription` returns an object from Apollo Client that contains loading, error, and data properties
 * you can use to render your UI.
 *
 * @param baseOptions options that will be passed into the subscription, supported options are listed on: https://www.apollographql.com/docs/react/api/react-hooks/#options;
 *
 * @example
 * const { data, loading, error } = useAgentChangesSubscription({
 *   variables: {
 *      selector: // value for 'selector'
 *      query: // value for 'query'
 *   },
 * });
 */
export function useAgentChangesSubscription(baseOptions?: Apollo.SubscriptionHookOptions<AgentChangesSubscription, AgentChangesSubscriptionVariables>) {
        const options = {...defaultOptions, ...baseOptions}
        return Apollo.useSubscription<AgentChangesSubscription, AgentChangesSubscriptionVariables>(AgentChangesDocument, options);
      }
export type AgentChangesSubscriptionHookResult = ReturnType<typeof useAgentChangesSubscription>;
export type AgentChangesSubscriptionResult = Apollo.SubscriptionResult<AgentChangesSubscription>;
export const GetAgentAndConfigurationsDocument = gql`
    query GetAgentAndConfigurations($agentId: ID!) {
  agent(id: $agentId) {
    id
    name
    architecture
    operatingSystem
    labels
    hostName
    platform
    version
    macAddress
    remoteAddress
    home
    status
    connectedAt
    disconnectedAt
    errorMessage
    configuration {
      Collector
    }
    configurationResource {
      metadata {
        id
        version
        name
      }
    }
    upgrade {
      status
      version
      error
    }
    upgradeAvailable
    features
  }
  configurations {
    configurations {
      metadata {
        id
        name
        version
        labels
      }
      spec {
        raw
      }
    }
  }
}
    `;

/**
 * __useGetAgentAndConfigurationsQuery__
 *
 * To run a query within a React component, call `useGetAgentAndConfigurationsQuery` and pass it any options that fit your needs.
 * When your component renders, `useGetAgentAndConfigurationsQuery` returns an object from Apollo Client that contains loading, error, and data properties
 * you can use to render your UI.
 *
 * @param baseOptions options that will be passed into the query, supported options are listed on: https://www.apollographql.com/docs/react/api/react-hooks/#options;
 *
 * @example
 * const { data, loading, error } = useGetAgentAndConfigurationsQuery({
 *   variables: {
 *      agentId: // value for 'agentId'
 *   },
 * });
 */
export function useGetAgentAndConfigurationsQuery(baseOptions: Apollo.QueryHookOptions<GetAgentAndConfigurationsQuery, GetAgentAndConfigurationsQueryVariables>) {
        const options = {...defaultOptions, ...baseOptions}
        return Apollo.useQuery<GetAgentAndConfigurationsQuery, GetAgentAndConfigurationsQueryVariables>(GetAgentAndConfigurationsDocument, options);
      }
export function useGetAgentAndConfigurationsLazyQuery(baseOptions?: Apollo.LazyQueryHookOptions<GetAgentAndConfigurationsQuery, GetAgentAndConfigurationsQueryVariables>) {
          const options = {...defaultOptions, ...baseOptions}
          return Apollo.useLazyQuery<GetAgentAndConfigurationsQuery, GetAgentAndConfigurationsQueryVariables>(GetAgentAndConfigurationsDocument, options);
        }
export type GetAgentAndConfigurationsQueryHookResult = ReturnType<typeof useGetAgentAndConfigurationsQuery>;
export type GetAgentAndConfigurationsLazyQueryHookResult = ReturnType<typeof useGetAgentAndConfigurationsLazyQuery>;
export type GetAgentAndConfigurationsQueryResult = Apollo.QueryResult<GetAgentAndConfigurationsQuery, GetAgentAndConfigurationsQueryVariables>;
export const GetConfigurationNamesDocument = gql`
    query GetConfigurationNames {
  configurations {
    configurations {
      metadata {
        id
        name
        version
        labels
      }
    }
  }
}
    `;

/**
 * __useGetConfigurationNamesQuery__
 *
 * To run a query within a React component, call `useGetConfigurationNamesQuery` and pass it any options that fit your needs.
 * When your component renders, `useGetConfigurationNamesQuery` returns an object from Apollo Client that contains loading, error, and data properties
 * you can use to render your UI.
 *
 * @param baseOptions options that will be passed into the query, supported options are listed on: https://www.apollographql.com/docs/react/api/react-hooks/#options;
 *
 * @example
 * const { data, loading, error } = useGetConfigurationNamesQuery({
 *   variables: {
 *   },
 * });
 */
export function useGetConfigurationNamesQuery(baseOptions?: Apollo.QueryHookOptions<GetConfigurationNamesQuery, GetConfigurationNamesQueryVariables>) {
        const options = {...defaultOptions, ...baseOptions}
        return Apollo.useQuery<GetConfigurationNamesQuery, GetConfigurationNamesQueryVariables>(GetConfigurationNamesDocument, options);
      }
export function useGetConfigurationNamesLazyQuery(baseOptions?: Apollo.LazyQueryHookOptions<GetConfigurationNamesQuery, GetConfigurationNamesQueryVariables>) {
          const options = {...defaultOptions, ...baseOptions}
          return Apollo.useLazyQuery<GetConfigurationNamesQuery, GetConfigurationNamesQueryVariables>(GetConfigurationNamesDocument, options);
        }
export type GetConfigurationNamesQueryHookResult = ReturnType<typeof useGetConfigurationNamesQuery>;
export type GetConfigurationNamesLazyQueryHookResult = ReturnType<typeof useGetConfigurationNamesLazyQuery>;
export type GetConfigurationNamesQueryResult = Apollo.QueryResult<GetConfigurationNamesQuery, GetConfigurationNamesQueryVariables>;
export const GetConfigRolloutAgentsDocument = gql`
    query getConfigRolloutAgents($name: String!) {
  configuration(name: $name) {
    metadata {
      name
      id
      version
    }
    agentCount
  }
}
    `;

/**
 * __useGetConfigRolloutAgentsQuery__
 *
 * To run a query within a React component, call `useGetConfigRolloutAgentsQuery` and pass it any options that fit your needs.
 * When your component renders, `useGetConfigRolloutAgentsQuery` returns an object from Apollo Client that contains loading, error, and data properties
 * you can use to render your UI.
 *
 * @param baseOptions options that will be passed into the query, supported options are listed on: https://www.apollographql.com/docs/react/api/react-hooks/#options;
 *
 * @example
 * const { data, loading, error } = useGetConfigRolloutAgentsQuery({
 *   variables: {
 *      name: // value for 'name'
 *   },
 * });
 */
export function useGetConfigRolloutAgentsQuery(baseOptions: Apollo.QueryHookOptions<GetConfigRolloutAgentsQuery, GetConfigRolloutAgentsQueryVariables>) {
        const options = {...defaultOptions, ...baseOptions}
        return Apollo.useQuery<GetConfigRolloutAgentsQuery, GetConfigRolloutAgentsQueryVariables>(GetConfigRolloutAgentsDocument, options);
      }
export function useGetConfigRolloutAgentsLazyQuery(baseOptions?: Apollo.LazyQueryHookOptions<GetConfigRolloutAgentsQuery, GetConfigRolloutAgentsQueryVariables>) {
          const options = {...defaultOptions, ...baseOptions}
          return Apollo.useLazyQuery<GetConfigRolloutAgentsQuery, GetConfigRolloutAgentsQueryVariables>(GetConfigRolloutAgentsDocument, options);
        }
export type GetConfigRolloutAgentsQueryHookResult = ReturnType<typeof useGetConfigRolloutAgentsQuery>;
export type GetConfigRolloutAgentsLazyQueryHookResult = ReturnType<typeof useGetConfigRolloutAgentsLazyQuery>;
export type GetConfigRolloutAgentsQueryResult = Apollo.QueryResult<GetConfigRolloutAgentsQuery, GetConfigRolloutAgentsQueryVariables>;
export const GetRenderedConfigValueDocument = gql`
    query getRenderedConfigValue($name: String!) {
  configuration(name: $name) {
    metadata {
      name
      id
      version
    }
    rendered
  }
}
    `;

/**
 * __useGetRenderedConfigValueQuery__
 *
 * To run a query within a React component, call `useGetRenderedConfigValueQuery` and pass it any options that fit your needs.
 * When your component renders, `useGetRenderedConfigValueQuery` returns an object from Apollo Client that contains loading, error, and data properties
 * you can use to render your UI.
 *
 * @param baseOptions options that will be passed into the query, supported options are listed on: https://www.apollographql.com/docs/react/api/react-hooks/#options;
 *
 * @example
 * const { data, loading, error } = useGetRenderedConfigValueQuery({
 *   variables: {
 *      name: // value for 'name'
 *   },
 * });
 */
export function useGetRenderedConfigValueQuery(baseOptions: Apollo.QueryHookOptions<GetRenderedConfigValueQuery, GetRenderedConfigValueQueryVariables>) {
        const options = {...defaultOptions, ...baseOptions}
        return Apollo.useQuery<GetRenderedConfigValueQuery, GetRenderedConfigValueQueryVariables>(GetRenderedConfigValueDocument, options);
      }
export function useGetRenderedConfigValueLazyQuery(baseOptions?: Apollo.LazyQueryHookOptions<GetRenderedConfigValueQuery, GetRenderedConfigValueQueryVariables>) {
          const options = {...defaultOptions, ...baseOptions}
          return Apollo.useLazyQuery<GetRenderedConfigValueQuery, GetRenderedConfigValueQueryVariables>(GetRenderedConfigValueDocument, options);
        }
export type GetRenderedConfigValueQueryHookResult = ReturnType<typeof useGetRenderedConfigValueQuery>;
export type GetRenderedConfigValueLazyQueryHookResult = ReturnType<typeof useGetRenderedConfigValueLazyQuery>;
export type GetRenderedConfigValueQueryResult = Apollo.QueryResult<GetRenderedConfigValueQuery, GetRenderedConfigValueQueryVariables>;
export const GetConfigurationDocument = gql`
    query GetConfiguration($name: String!) {
  configuration(name: $name) {
    metadata {
      id
      name
      description
      labels
      version
    }
    agentCount
    spec {
      raw
      sources {
        type
        name
        displayName
        parameters {
          name
          value
        }
        processors {
          type
          displayName
          parameters {
            name
            value
          }
          disabled
        }
        disabled
      }
      destinations {
        type
        name
        displayName
        parameters {
          name
          value
        }
        processors {
          type
          displayName
          parameters {
            name
            value
          }
          disabled
        }
        disabled
      }
      selector {
        matchLabels
      }
    }
    graph {
      attributes
      sources {
        id
        type
        label
        attributes
      }
      intermediates {
        id
        type
        label
        attributes
      }
      targets {
        id
        type
        label
        attributes
      }
      edges {
        id
        source
        target
      }
    }
  }
}
    `;

/**
 * __useGetConfigurationQuery__
 *
 * To run a query within a React component, call `useGetConfigurationQuery` and pass it any options that fit your needs.
 * When your component renders, `useGetConfigurationQuery` returns an object from Apollo Client that contains loading, error, and data properties
 * you can use to render your UI.
 *
 * @param baseOptions options that will be passed into the query, supported options are listed on: https://www.apollographql.com/docs/react/api/react-hooks/#options;
 *
 * @example
 * const { data, loading, error } = useGetConfigurationQuery({
 *   variables: {
 *      name: // value for 'name'
 *   },
 * });
 */
export function useGetConfigurationQuery(baseOptions: Apollo.QueryHookOptions<GetConfigurationQuery, GetConfigurationQueryVariables>) {
        const options = {...defaultOptions, ...baseOptions}
        return Apollo.useQuery<GetConfigurationQuery, GetConfigurationQueryVariables>(GetConfigurationDocument, options);
      }
export function useGetConfigurationLazyQuery(baseOptions?: Apollo.LazyQueryHookOptions<GetConfigurationQuery, GetConfigurationQueryVariables>) {
          const options = {...defaultOptions, ...baseOptions}
          return Apollo.useLazyQuery<GetConfigurationQuery, GetConfigurationQueryVariables>(GetConfigurationDocument, options);
        }
export type GetConfigurationQueryHookResult = ReturnType<typeof useGetConfigurationQuery>;
export type GetConfigurationLazyQueryHookResult = ReturnType<typeof useGetConfigurationLazyQuery>;
export type GetConfigurationQueryResult = Apollo.QueryResult<GetConfigurationQuery, GetConfigurationQueryVariables>;
export const DestinationsAndTypesDocument = gql`
    query DestinationsAndTypes {
  destinationTypes {
    kind
    apiVersion
    metadata {
      id
      version
      name
      displayName
      description
      icon
      version
    }
    spec {
      version
      parameters {
        label
        type
        name
        description
        default
        validValues
        relevantIf {
          name
          value
          operator
        }
        documentation {
          text
          url
        }
        advancedConfig
        required
        options {
          creatable
          multiline
          trackUnchecked
          sectionHeader
          gridColumns
          labels
          metricCategories {
            label
            column
            metrics {
              name
              description
              kpi
            }
          }
          password
        }
      }
      supportedPlatforms
      telemetryTypes
    }
  }
  destinations {
    metadata {
      id
      version
      name
    }
    spec {
      type
      parameters {
        name
        value
      }
      disabled
    }
  }
}
    `;

/**
 * __useDestinationsAndTypesQuery__
 *
 * To run a query within a React component, call `useDestinationsAndTypesQuery` and pass it any options that fit your needs.
 * When your component renders, `useDestinationsAndTypesQuery` returns an object from Apollo Client that contains loading, error, and data properties
 * you can use to render your UI.
 *
 * @param baseOptions options that will be passed into the query, supported options are listed on: https://www.apollographql.com/docs/react/api/react-hooks/#options;
 *
 * @example
 * const { data, loading, error } = useDestinationsAndTypesQuery({
 *   variables: {
 *   },
 * });
 */
export function useDestinationsAndTypesQuery(baseOptions?: Apollo.QueryHookOptions<DestinationsAndTypesQuery, DestinationsAndTypesQueryVariables>) {
        const options = {...defaultOptions, ...baseOptions}
        return Apollo.useQuery<DestinationsAndTypesQuery, DestinationsAndTypesQueryVariables>(DestinationsAndTypesDocument, options);
      }
export function useDestinationsAndTypesLazyQuery(baseOptions?: Apollo.LazyQueryHookOptions<DestinationsAndTypesQuery, DestinationsAndTypesQueryVariables>) {
          const options = {...defaultOptions, ...baseOptions}
          return Apollo.useLazyQuery<DestinationsAndTypesQuery, DestinationsAndTypesQueryVariables>(DestinationsAndTypesDocument, options);
        }
export type DestinationsAndTypesQueryHookResult = ReturnType<typeof useDestinationsAndTypesQuery>;
export type DestinationsAndTypesLazyQueryHookResult = ReturnType<typeof useDestinationsAndTypesLazyQuery>;
export type DestinationsAndTypesQueryResult = Apollo.QueryResult<DestinationsAndTypesQuery, DestinationsAndTypesQueryVariables>;
export const SourceTypesDocument = gql`
    query sourceTypes {
  sourceTypes {
    apiVersion
    kind
    metadata {
      id
      name
      version
      displayName
      description
      icon
    }
    spec {
      parameters {
        name
        label
        description
        relevantIf {
          name
          operator
          value
        }
        documentation {
          text
          url
        }
        advancedConfig
        required
        type
        validValues
        default
        options {
          creatable
          multiline
          trackUnchecked
          sectionHeader
          gridColumns
          labels
          metricCategories {
            label
            column
            metrics {
              name
              description
              kpi
            }
          }
          password
        }
      }
      supportedPlatforms
      version
      telemetryTypes
    }
  }
}
    `;

/**
 * __useSourceTypesQuery__
 *
 * To run a query within a React component, call `useSourceTypesQuery` and pass it any options that fit your needs.
 * When your component renders, `useSourceTypesQuery` returns an object from Apollo Client that contains loading, error, and data properties
 * you can use to render your UI.
 *
 * @param baseOptions options that will be passed into the query, supported options are listed on: https://www.apollographql.com/docs/react/api/react-hooks/#options;
 *
 * @example
 * const { data, loading, error } = useSourceTypesQuery({
 *   variables: {
 *   },
 * });
 */
export function useSourceTypesQuery(baseOptions?: Apollo.QueryHookOptions<SourceTypesQuery, SourceTypesQueryVariables>) {
        const options = {...defaultOptions, ...baseOptions}
        return Apollo.useQuery<SourceTypesQuery, SourceTypesQueryVariables>(SourceTypesDocument, options);
      }
export function useSourceTypesLazyQuery(baseOptions?: Apollo.LazyQueryHookOptions<SourceTypesQuery, SourceTypesQueryVariables>) {
          const options = {...defaultOptions, ...baseOptions}
          return Apollo.useLazyQuery<SourceTypesQuery, SourceTypesQueryVariables>(SourceTypesDocument, options);
        }
export type SourceTypesQueryHookResult = ReturnType<typeof useSourceTypesQuery>;
export type SourceTypesLazyQueryHookResult = ReturnType<typeof useSourceTypesLazyQuery>;
export type SourceTypesQueryResult = Apollo.QueryResult<SourceTypesQuery, SourceTypesQueryVariables>;
export const GetConfigNamesDocument = gql`
    query getConfigNames {
  configurations {
    configurations {
      metadata {
        id
        name
        version
      }
    }
  }
}
    `;

/**
 * __useGetConfigNamesQuery__
 *
 * To run a query within a React component, call `useGetConfigNamesQuery` and pass it any options that fit your needs.
 * When your component renders, `useGetConfigNamesQuery` returns an object from Apollo Client that contains loading, error, and data properties
 * you can use to render your UI.
 *
 * @param baseOptions options that will be passed into the query, supported options are listed on: https://www.apollographql.com/docs/react/api/react-hooks/#options;
 *
 * @example
 * const { data, loading, error } = useGetConfigNamesQuery({
 *   variables: {
 *   },
 * });
 */
export function useGetConfigNamesQuery(baseOptions?: Apollo.QueryHookOptions<GetConfigNamesQuery, GetConfigNamesQueryVariables>) {
        const options = {...defaultOptions, ...baseOptions}
        return Apollo.useQuery<GetConfigNamesQuery, GetConfigNamesQueryVariables>(GetConfigNamesDocument, options);
      }
export function useGetConfigNamesLazyQuery(baseOptions?: Apollo.LazyQueryHookOptions<GetConfigNamesQuery, GetConfigNamesQueryVariables>) {
          const options = {...defaultOptions, ...baseOptions}
          return Apollo.useLazyQuery<GetConfigNamesQuery, GetConfigNamesQueryVariables>(GetConfigNamesDocument, options);
        }
export type GetConfigNamesQueryHookResult = ReturnType<typeof useGetConfigNamesQuery>;
export type GetConfigNamesLazyQueryHookResult = ReturnType<typeof useGetConfigNamesLazyQuery>;
export type GetConfigNamesQueryResult = Apollo.QueryResult<GetConfigNamesQuery, GetConfigNamesQueryVariables>;
export const DestinationsDocument = gql`
    query Destinations {
  destinations {
    kind
    metadata {
      id
      name
      version
    }
    spec {
      type
    }
  }
}
    `;

/**
 * __useDestinationsQuery__
 *
 * To run a query within a React component, call `useDestinationsQuery` and pass it any options that fit your needs.
 * When your component renders, `useDestinationsQuery` returns an object from Apollo Client that contains loading, error, and data properties
 * you can use to render your UI.
 *
 * @param baseOptions options that will be passed into the query, supported options are listed on: https://www.apollographql.com/docs/react/api/react-hooks/#options;
 *
 * @example
 * const { data, loading, error } = useDestinationsQuery({
 *   variables: {
 *   },
 * });
 */
export function useDestinationsQuery(baseOptions?: Apollo.QueryHookOptions<DestinationsQuery, DestinationsQueryVariables>) {
        const options = {...defaultOptions, ...baseOptions}
        return Apollo.useQuery<DestinationsQuery, DestinationsQueryVariables>(DestinationsDocument, options);
      }
export function useDestinationsLazyQuery(baseOptions?: Apollo.LazyQueryHookOptions<DestinationsQuery, DestinationsQueryVariables>) {
          const options = {...defaultOptions, ...baseOptions}
          return Apollo.useLazyQuery<DestinationsQuery, DestinationsQueryVariables>(DestinationsDocument, options);
        }
export type DestinationsQueryHookResult = ReturnType<typeof useDestinationsQuery>;
export type DestinationsLazyQueryHookResult = ReturnType<typeof useDestinationsLazyQuery>;
export type DestinationsQueryResult = Apollo.QueryResult<DestinationsQuery, DestinationsQueryVariables>;
export const GetOverviewPageDocument = gql`
    query getOverviewPage($configIDs: [ID!], $destinationIDs: [ID!], $period: String!, $telemetryType: String!) {
  overviewPage(
    configIDs: $configIDs
    destinationIDs: $destinationIDs
    period: $period
    telemetryType: $telemetryType
  ) {
    graph {
      attributes
      sources {
        id
        label
        type
        attributes
      }
      intermediates {
        id
        label
        type
        attributes
      }
      targets {
        id
        label
        type
        attributes
      }
      edges {
        id
        source
        target
      }
    }
  }
}
    `;

/**
 * __useGetOverviewPageQuery__
 *
 * To run a query within a React component, call `useGetOverviewPageQuery` and pass it any options that fit your needs.
 * When your component renders, `useGetOverviewPageQuery` returns an object from Apollo Client that contains loading, error, and data properties
 * you can use to render your UI.
 *
 * @param baseOptions options that will be passed into the query, supported options are listed on: https://www.apollographql.com/docs/react/api/react-hooks/#options;
 *
 * @example
 * const { data, loading, error } = useGetOverviewPageQuery({
 *   variables: {
 *      configIDs: // value for 'configIDs'
 *      destinationIDs: // value for 'destinationIDs'
 *      period: // value for 'period'
 *      telemetryType: // value for 'telemetryType'
 *   },
 * });
 */
export function useGetOverviewPageQuery(baseOptions: Apollo.QueryHookOptions<GetOverviewPageQuery, GetOverviewPageQueryVariables>) {
        const options = {...defaultOptions, ...baseOptions}
        return Apollo.useQuery<GetOverviewPageQuery, GetOverviewPageQueryVariables>(GetOverviewPageDocument, options);
      }
export function useGetOverviewPageLazyQuery(baseOptions?: Apollo.LazyQueryHookOptions<GetOverviewPageQuery, GetOverviewPageQueryVariables>) {
          const options = {...defaultOptions, ...baseOptions}
          return Apollo.useLazyQuery<GetOverviewPageQuery, GetOverviewPageQueryVariables>(GetOverviewPageDocument, options);
        }
export type GetOverviewPageQueryHookResult = ReturnType<typeof useGetOverviewPageQuery>;
export type GetOverviewPageLazyQueryHookResult = ReturnType<typeof useGetOverviewPageLazyQuery>;
export type GetOverviewPageQueryResult = Apollo.QueryResult<GetOverviewPageQuery, GetOverviewPageQueryVariables>;
export const OverviewMetricsDocument = gql`
    subscription OverviewMetrics($period: String!, $configIDs: [ID!], $destinationIDs: [ID!]) {
  overviewMetrics(
    period: $period
    configIDs: $configIDs
    destinationIDs: $destinationIDs
  ) {
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

/**
 * __useOverviewMetricsSubscription__
 *
 * To run a query within a React component, call `useOverviewMetricsSubscription` and pass it any options that fit your needs.
 * When your component renders, `useOverviewMetricsSubscription` returns an object from Apollo Client that contains loading, error, and data properties
 * you can use to render your UI.
 *
 * @param baseOptions options that will be passed into the subscription, supported options are listed on: https://www.apollographql.com/docs/react/api/react-hooks/#options;
 *
 * @example
 * const { data, loading, error } = useOverviewMetricsSubscription({
 *   variables: {
 *      period: // value for 'period'
 *      configIDs: // value for 'configIDs'
 *      destinationIDs: // value for 'destinationIDs'
 *   },
 * });
 */
export function useOverviewMetricsSubscription(baseOptions: Apollo.SubscriptionHookOptions<OverviewMetricsSubscription, OverviewMetricsSubscriptionVariables>) {
        const options = {...defaultOptions, ...baseOptions}
        return Apollo.useSubscription<OverviewMetricsSubscription, OverviewMetricsSubscriptionVariables>(OverviewMetricsDocument, options);
      }
export type OverviewMetricsSubscriptionHookResult = ReturnType<typeof useOverviewMetricsSubscription>;
export type OverviewMetricsSubscriptionResult = Apollo.SubscriptionResult<OverviewMetricsSubscription>;
export const DestinationsInConfigsDocument = gql`
    query DestinationsInConfigs {
  destinationsInConfigs {
    kind
    metadata {
      id
      version
      name
    }
    spec {
      type
    }
  }
}
    `;

/**
 * __useDestinationsInConfigsQuery__
 *
 * To run a query within a React component, call `useDestinationsInConfigsQuery` and pass it any options that fit your needs.
 * When your component renders, `useDestinationsInConfigsQuery` returns an object from Apollo Client that contains loading, error, and data properties
 * you can use to render your UI.
 *
 * @param baseOptions options that will be passed into the query, supported options are listed on: https://www.apollographql.com/docs/react/api/react-hooks/#options;
 *
 * @example
 * const { data, loading, error } = useDestinationsInConfigsQuery({
 *   variables: {
 *   },
 * });
 */
export function useDestinationsInConfigsQuery(baseOptions?: Apollo.QueryHookOptions<DestinationsInConfigsQuery, DestinationsInConfigsQueryVariables>) {
        const options = {...defaultOptions, ...baseOptions}
        return Apollo.useQuery<DestinationsInConfigsQuery, DestinationsInConfigsQueryVariables>(DestinationsInConfigsDocument, options);
      }
export function useDestinationsInConfigsLazyQuery(baseOptions?: Apollo.LazyQueryHookOptions<DestinationsInConfigsQuery, DestinationsInConfigsQueryVariables>) {
          const options = {...defaultOptions, ...baseOptions}
          return Apollo.useLazyQuery<DestinationsInConfigsQuery, DestinationsInConfigsQueryVariables>(DestinationsInConfigsDocument, options);
        }
export type DestinationsInConfigsQueryHookResult = ReturnType<typeof useDestinationsInConfigsQuery>;
export type DestinationsInConfigsLazyQueryHookResult = ReturnType<typeof useDestinationsInConfigsLazyQuery>;
export type DestinationsInConfigsQueryResult = Apollo.QueryResult<DestinationsInConfigsQuery, DestinationsInConfigsQueryVariables>;