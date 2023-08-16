import {
  ParameterDefinition,
  ParameterType,
  RelevantIfOperatorType,
  SourceType,
  Destination,
  PipelineType,
  ProcessorType,
  Configuration,
  AdditionalInfo,
} from "../../../graphql/generated";
import { APIVersion } from "../../../types/resources";

const DEFAULT_PARAMETER_OPTIONS = {
  creatable: false,
  multiline: false,
  trackUnchecked: false,
  subHeader: null,
  horizontalDivider: false,
  sectionHeader: null,
  gridColumns: null,
  metricCategories: null,
  labels: null,
  password: null,
  sensitive: false,
};

// This file contains dummy resources used for testing the ResourceConfigForm
// component.  These are also used in resource form stories.

/* -------------------------- ParameterDefinitions -------------------------- */

export const stringDef: ParameterDefinition = {
  name: "string_name",
  label: "String Input",
  description: "Here is the description.",
  required: false,
  options: DEFAULT_PARAMETER_OPTIONS,
  advancedConfig: false,

  type: ParameterType.String,

  default: "default-value",
};

export const stringDefRequired: ParameterDefinition = {
  name: "string_required_name",
  label: "String Input",
  description: "Here is the description.",
  required: true,
  options: DEFAULT_PARAMETER_OPTIONS,
  advancedConfig: false,

  type: ParameterType.String,

  default: "default-required-value",
};

export const stringPasswordDef: ParameterDefinition = {
  name: "string_password_name",
  label: "String Password Input",
  description: "Here is the description.",
  required: false,
  options: {
    ...DEFAULT_PARAMETER_OPTIONS,
    sensitive: true,
  },

  type: ParameterType.String,
};

export const enumDef: ParameterDefinition = {
  name: "enum_name",
  label: "Enum Input",
  description: "Here is the description.",
  required: false,
  advancedConfig: false,
  options: DEFAULT_PARAMETER_OPTIONS,

  type: ParameterType.Enum,

  default: "option1",
  validValues: ["option1", "option2", "option3"],
};

export const enumsDef: ParameterDefinition = {
  name: "enums_name",
  label: "Enums Input",
  description: "Here is the description.",
  required: false,
  advancedConfig: false,
  options: DEFAULT_PARAMETER_OPTIONS,

  type: ParameterType.Enums,

  default: ["logs", "metrics", "traces"],
  validValues: ["logs", "metrics", "traces"],
};

export const stringsDef: ParameterDefinition = {
  name: "strings_name",
  label: "Multi String Input",
  description: "Here is the description.",
  required: false,
  advancedConfig: false,
  options: DEFAULT_PARAMETER_OPTIONS,

  type: ParameterType.Strings,

  default: ["option1", "option2"],
};

export const boolDef: ParameterDefinition = {
  name: "bool_name",
  label: "Bool Input",
  description: "Here is the description.",
  required: false,
  advancedConfig: false,
  options: DEFAULT_PARAMETER_OPTIONS,

  type: ParameterType.Bool,

  default: true,
};

export const telemetrySectionBoolDef: ParameterDefinition = {
  name: "enable_metrics",
  label: "Enable Metrics",
  description: "Here is the description.",
  required: false,
  advancedConfig: false,
  options: DEFAULT_PARAMETER_OPTIONS,

  type: ParameterType.Bool,
  default: false,
};

export const boolDefaultFalseDef: ParameterDefinition = {
  name: "bool_default_false_name",
  label: "Bool Default False Input",
  description: "Here is the description.",
  advancedConfig: false,
  required: false,
  options: DEFAULT_PARAMETER_OPTIONS,

  type: ParameterType.Bool,

  default: false,
  documentation: null,
  validValues: null,
  relevantIf: null,
};

export const intDef: ParameterDefinition = {
  name: "int_name",
  label: "Int Input",
  description: "Here is the description.",
  advancedConfig: false,
  required: false,
  options: DEFAULT_PARAMETER_OPTIONS,

  type: ParameterType.Int,

  default: 25,
};

export const mapDef: ParameterDefinition = {
  name: "map_name",
  label: "Map Input",
  description: "Here is the description.",
  advancedConfig: false,
  required: false,
  options: DEFAULT_PARAMETER_OPTIONS,

  type: ParameterType.Map,

  default: {
    one: "1",
    two: "2",
  },
};

export const timezoneDef: ParameterDefinition = {
  name: "timezone_name",
  label: "Timezone Input",
  description: "Here is the description.",
  advancedConfig: false,
  required: false,
  options: DEFAULT_PARAMETER_OPTIONS,
  type: ParameterType.Timezone,
  default: "America/New_York",
};

export const yamlDef: ParameterDefinition = {
  name: "yaml_name",
  label: "Yaml Input",
  description: "Here is the description.",
  advancedConfig: false,
  required: false,
  options: DEFAULT_PARAMETER_OPTIONS,
  type: ParameterType.Yaml,
  default: "default: value",
};

export const relevantIfDef: ParameterDefinition = {
  name: "string_name",
  label: "String Input",
  description: "Here is the description.",
  advancedConfig: false,
  required: false,
  options: DEFAULT_PARAMETER_OPTIONS,

  type: ParameterType.String,

  relevantIf: [
    {
      name: "bool_default_false_name",
      operator: RelevantIfOperatorType.Equals,
      value: true,
    },
  ],

  default: "default-value",
  documentation: null,
  validValues: null,
};

export const awsCloudWatchFieldInputDef = {
  name: "named_groups",
  label: "Groups",
  description:
    "Configuration for Log Groups, by default all Log Groups and Log Streams will be collected.",
  required: false,
  type: ParameterType.AwsCloudwatchNamedField,
  default: [],
  relevantIf: [
    {
      name: "discovery_type",
      operator: RelevantIfOperatorType.Equals,
      value: "Named",
    },
  ],
  hidden: false,
  advancedConfig: false,
  options: {
    creatable: false,
  },
};

export const FileLogSort = {
  name: "sort_rules",
  label: "Sort Rules",
  description:
    "Configuration for Log Groups, by default all Log Groups and Log Streams will be collected.",
  required: false,
  type: ParameterType.FileLogSort,
  default: [],
  hidden: false,
  advancedConfig: false,
  options: {
    creatable: false,
  },
};

export const metricParamInput = {
  name: "process_metrics_filtering",
  label: "",
  description: "",
  required: false,
  type: ParameterType.Metrics,
  default: [
    "process.context_switches",
    "process.cpu.utilization",
    "process.disk.operations",
    "process.memory.utilization",
    "process.open_file_descriptors",
    "process.paging.faults",
    "process.signals_pending",
    "process.threads",
  ],
  hidden: false,
  advancedConfig: false,
  options: {
    gridColumns: 12,
    metricCategories: [
      {
        label: "Process",
        column: 0,
        metrics: [
          {
            name: "process.threads",
            description: "",
            kpi: false,
          },
        ],
      },
    ],
  },
};

export const relevantIfNotEqualDef: ParameterDefinition = {
  name: "int_name",
  label: "Number Input",
  description: "Here is the description.",
  advancedConfig: false,
  required: false,
  options: DEFAULT_PARAMETER_OPTIONS,

  type: ParameterType.Int,

  relevantIf: [
    {
      name: "string_name",
      operator: RelevantIfOperatorType.NotEquals,
      value: "default-value",
    },
  ],

  documentation: null,
  validValues: null,
};

export const additionalInfo: AdditionalInfo = {
  message: "test message",
  documentation: [
    {
      text: "test text",
      url: "test url",
    },
  ],
};

/* ----------------------------- Resource Types ----------------------------- */

export const ResourceType1: SourceType = {
  apiVersion: APIVersion.V1,
  kind: "ResourceType",
  metadata: {
    id: "resource-type-1",
    name: "resource-type-1",
    displayName: "ResourceType One",
    description: "A description for resource one.",
    icon: "/icons/destinations/otlp.svg",
    version: 0,
  },
  spec: {
    version: "0.0.0",
    parameters: [
      stringDef,
      stringDefRequired,
      enumDef,
      stringsDef,
      boolDef,
      intDef,
    ],
    telemetryTypes: [],

    supportedPlatforms: ["linux", "macos", "windows"],
  },
};

export const ResourceType2: SourceType = {
  apiVersion: APIVersion.V1,
  kind: "ResourceType",
  metadata: {
    id: "resource-type-2",
    name: "resource-type-2",
    displayName: "ResourceType Two",
    description: "A description for resource one.",
    icon: "/icons/destinations/otlp.svg",
    version: 0,
  },
  spec: {
    version: "0.0.0",
    parameters: [boolDefaultFalseDef, relevantIfDef],

    supportedPlatforms: ["linux", "macos", "windows"],
    telemetryTypes: [],
  },
};

export const WindowsOnlyResourceType: SourceType = {
  apiVersion: APIVersion.V1,
  kind: "ResourceType",
  metadata: {
    id: "windows-only-resource-type",
    name: "windows-only-resource-type",
    displayName: "Windows Only",
    description: "A description for resource one.",
    icon: "/icons/destinations/otlp.svg",
    version: 0,
  },
  spec: {
    version: "0.0.0",
    parameters: [boolDefaultFalseDef, relevantIfDef],

    supportedPlatforms: ["windows"],
    telemetryTypes: [],
  },
};

export const ResourceType3: SourceType = {
  apiVersion: APIVersion.V1,
  kind: "ResourceType",
  metadata: {
    id: "resource-type-3",
    name: "resource-type-3",
    displayName: "ResourceType Three",
    description: "A description for resource type three.",
    icon: "/icons/destinations/otlp.svg",
    version: 0,
  },
  spec: {
    version: "0.0.0",
    parameters: [stringDef, relevantIfNotEqualDef],

    supportedPlatforms: ["linux", "macos", "windows"],
    telemetryTypes: [],
  },
};

export const SupportsLogs: SourceType = {
  apiVersion: APIVersion.V1,
  kind: "ResourceType",
  metadata: {
    id: "supports-logs",
    name: "supports-logs",
    displayName: "Supports Logs",
    description: "A resource that supports logs.",
    icon: "/icons/destinations/otlp.svg",
    version: 0,
  },
  spec: {
    version: "0.0.0",
    parameters: [boolDefaultFalseDef, relevantIfDef],

    supportedPlatforms: ["linux", "macos", "windows"],
    telemetryTypes: [PipelineType.Logs],
  },
};

export const SupportsMetrics: SourceType = {
  apiVersion: APIVersion.V1,
  kind: "ResourceType",
  metadata: {
    id: "supports-metrics",
    name: "supports-metrics",
    displayName: "Supports Metrics",
    description: "A resource that supports metrics.",
    icon: "/icons/destinations/otlp.svg",
    version: 0,
  },
  spec: {
    version: "0.0.0",
    parameters: [boolDefaultFalseDef, relevantIfDef],

    supportedPlatforms: ["linux", "macos", "windows"],
    telemetryTypes: [PipelineType.Metrics],
  },
};

export const SupportsBoth: SourceType = {
  apiVersion: APIVersion.V1,
  kind: "ResourceType",
  metadata: {
    id: "supports-logs-and-metrics",
    name: "supports-logs-and-metrics",
    displayName: "Supports Logs and Metrics",
    description: "A resource that supports logs and metrics.",
    icon: "/icons/destinations/otlp.svg",
    version: 0,
  },
  spec: {
    version: "0.0.0",
    parameters: [boolDefaultFalseDef, relevantIfDef],

    supportedPlatforms: ["linux", "macos", "windows"],
    telemetryTypes: [PipelineType.Logs, PipelineType.Metrics],
  },
};

export const ProcessorTypeSeverity: ProcessorType = {
  apiVersion: APIVersion.V1,
  kind: "ProcessorType",
  metadata: {
    id: "",
    name: "severity_processor",
    displayName: "Severity Filter",
    description: "This filters logs by severity",
    icon: "/icons/destinations/otlp.svg",
    version: 0,
  },
  spec: {
    version: "0.0.0",
    parameters: [boolDefaultFalseDef, relevantIfDef],
    supportedPlatforms: ["linux", "macos", "windows"],
    telemetryTypes: [PipelineType.Logs],
  },
};

export const ProcessorTypeMetric: ProcessorType = {
  apiVersion: APIVersion.V1,
  kind: "ProcessorType",
  metadata: {
    id: "",
    name: "metric_processor",
    displayName: "Metric Filter",
    description: "This processes metrics.",
    icon: "/icons/destinations/otlp.svg",
    version: 0,
  },
  spec: {
    version: "0.0.0",
    parameters: [boolDefaultFalseDef, relevantIfDef],
    supportedPlatforms: ["linux", "macos", "windows"],
    telemetryTypes: [PipelineType.Metrics],
  },
};

/* -------------------------------- Resources ------------------------------- */

export const Config1: Configuration = {
  apiVersion: APIVersion.V1,
  kind: "Configuration",
  metadata: {
    name: "config-1-name",
    id: "config-1-name",
    version: 1,
  },
  spec: {},
  status: {
    currentVersion: 1,
    rollout: {
      completed: 0,
      errors: 0,
      pending: 0,
      phase: 0,
      status: 0,
      waiting: 0,
    },
    current: true,
    pending: false,
    latest: true,
  },
};

export const Config2: Configuration = {
  apiVersion: APIVersion.V1,
  kind: "Configuration",
  metadata: {
    name: "config-2-name",
    id: "config-2-name",
    version: 1,
  },
  spec: {},
  status: {
    currentVersion: 1,
    rollout: {
      completed: 0,
      errors: 0,
      pending: 0,
      phase: 0,
      status: 0,
      waiting: 0,
    },
    current: true,
    pending: false,
    latest: true,
  },
};

// This destination is type resource-type-1
export const Destination1: Destination = {
  apiVersion: APIVersion.V1,
  kind: "Destination",
  metadata: {
    name: "destination-1-name",
    id: "destination-1-name",
    version: 0,
  },
  spec: {
    parameters: [],
    type: "resource-type-1:1",
    disabled: false,
  },
};

// This destination is type resource-type-1
export const Destination2: Destination = {
  apiVersion: APIVersion.V1,
  kind: "Destination",
  metadata: {
    name: "destination-2-name",
    id: "destination-2-name",
    version: 0,
  },
  spec: {
    parameters: [],
    type: "resource-type-1:1",
    disabled: false,
  },
};
