import { MockedResponse } from "@apollo/client/testing";
import {
  GetDestinationWithTypeDocument,
  GetDestinationWithTypeQuery,
  GetSourceWithTypeDocument,
  GetSourceWithTypeQuery,
  SourceTypeDocument,
} from "../../../graphql/generated";
import { MinimumRequiredConfig } from "../../PipelineGraph/PipelineGraph";
import {
  destination0,
  destination0Type,
  destination0Name,
  destination0_PAUSED,
  source1Name,
  source1,
  source1Type,
  source1_PAUSED,
} from "./test-resources";

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
  __typename: "ParameterOptions",
};

export const fileSourceTypeQuery: MockedResponse<Record<string, any>>[] = [
  {
    request: {
      operationName: "SourceType",
      query: SourceTypeDocument,
      variables: { name: "file" },
    },
    result: () => {
      return {
        data: {
          sourceType: {
            metadata: {
              id: "8cfe19b4-35c2-4aa9-b1f5-5477a7cb7832",
              version: 1,
              name: "file",
              displayName: "File",
              icon: "/icons/sources/file.svg",
              description: "Collect logs from generic log files.",
              additionalInfo: null,
            },
            spec: {
              parameters: [
                {
                  label: "File Path(s)",
                  name: "file_path",
                  description: "File or directory paths to tail for logs.",
                  required: true,
                  type: "strings",
                  default: [],
                  documentation: null,
                  relevantIf: null,
                  advancedConfig: false,
                  validValues: null,
                  options: DEFAULT_PARAMETER_OPTIONS,
                },
                {
                  label: "Exclude File Path(s)",
                  name: "exclude_file_path",
                  description: "File or directory paths to exclude.",
                  required: false,
                  type: "strings",
                  default: [],
                  documentation: null,
                  relevantIf: null,
                  advancedConfig: false,
                  validValues: null,
                  options: DEFAULT_PARAMETER_OPTIONS,
                },
                {
                  label: "Log Type",
                  name: "log_type",
                  description:
                    "A friendly name that will be added to each log entry as an attribute.",
                  required: false,
                  type: "string",
                  default: "file",
                  documentation: null,
                  relevantIf: null,
                  advancedConfig: false,
                  validValues: null,
                  options: DEFAULT_PARAMETER_OPTIONS,
                },
                {
                  label: "Parse Format",
                  name: "parse_format",
                  description:
                    "Method to use when parsing. When regex is selected, 'Regex Pattern' must be set.",
                  required: false,
                  type: "enum",
                  default: "none",
                  documentation: null,
                  relevantIf: null,
                  advancedConfig: false,
                  validValues: ["none", "json", "regex"],
                  options: DEFAULT_PARAMETER_OPTIONS,
                },
                {
                  label: "Regex Pattern",
                  name: "regex_pattern",
                  description:
                    "The regex pattern used when parsing log entries.",
                  required: true,
                  type: "string",
                  default: "",
                  documentation: null,
                  relevantIf: [
                    {
                      name: "parse_format",
                      operator: "equals",
                      value: "regex",
                    },
                  ],
                  advancedConfig: false,
                  validValues: null,
                  options: DEFAULT_PARAMETER_OPTIONS,
                },
                {
                  label: "Multiline Parsing",
                  name: "multiline_parsing",
                  description:
                    "Enable multiline parsing options. Either specifying a regex for where a log starts or ends.",
                  required: false,
                  type: "enum",
                  default: "none",
                  documentation: null,
                  relevantIf: null,
                  advancedConfig: false,
                  validValues: [
                    "none",
                    "specify line start",
                    "specify line end",
                  ],
                  options: DEFAULT_PARAMETER_OPTIONS,
                },
                {
                  label: "Multiline Start Pattern",
                  name: "multiline_line_start_pattern",
                  description:
                    "Regex pattern that matches beginning of a log entry, for handling multiline logs.",
                  required: true,
                  type: "string",
                  default: "",
                  documentation: null,
                  relevantIf: [
                    {
                      name: "multiline_parsing",
                      operator: "equals",
                      value: "specify line start",
                    },
                  ],
                  advancedConfig: false,
                  validValues: null,
                  options: DEFAULT_PARAMETER_OPTIONS,
                },
                {
                  label: "Multiline End Pattern",
                  name: "multiline_line_end_pattern",
                  description:
                    "Regex pattern that matches end of a log entry, useful for terminating parsing of multiline logs.",
                  required: true,
                  type: "string",
                  default: "",
                  documentation: null,
                  relevantIf: [
                    {
                      name: "multiline_parsing",
                      operator: "equals",
                      value: "specify line end",
                    },
                  ],
                  advancedConfig: false,
                  validValues: null,
                  options: DEFAULT_PARAMETER_OPTIONS,
                },
                {
                  label: "Parse Timestamp",
                  name: "parse_timestamp",
                  description:
                    "Whether to parse the timestamp from the log entry.",
                  required: false,
                  type: "bool",
                  default: false,
                  documentation: null,
                  relevantIf: [
                    {
                      name: "parse_format",
                      operator: "notEquals",
                      value: "none",
                    },
                  ],
                  advancedConfig: false,
                  validValues: null,
                  options: DEFAULT_PARAMETER_OPTIONS,
                },
                {
                  label: "Timestamp Field",
                  name: "timestamp_field",
                  description:
                    "The field containing the timestamp in the log entry.",
                  required: true,
                  type: "string",
                  default: "timestamp",
                  documentation: null,
                  relevantIf: [
                    {
                      name: "parse_timestamp",
                      operator: "equals",
                      value: true,
                    },
                    {
                      name: "parse_format",
                      operator: "notEquals",
                      value: "none",
                    },
                  ],
                  advancedConfig: false,
                  validValues: null,
                  options: DEFAULT_PARAMETER_OPTIONS,
                },
                {
                  label: "Timestamp Format",
                  name: "parse_timestamp_format",
                  description:
                    "The format of the timestamp in the log entry. Choose a common format, or specify a custom format.",
                  required: false,
                  type: "enum",
                  default: "ISO8601",
                  documentation: null,
                  relevantIf: [
                    {
                      name: "parse_timestamp",
                      operator: "equals",
                      value: true,
                    },
                    {
                      name: "parse_format",
                      operator: "notEquals",
                      value: "none",
                    },
                  ],
                  advancedConfig: false,
                  validValues: ["ISO8601", "RFC3339", "Epoch", "Manual"],
                  options: DEFAULT_PARAMETER_OPTIONS,
                },
                {
                  label: "Epoch Layout",
                  name: "epoch_timestamp_format",
                  description: "The layout of the epoch-based timestamp.",
                  required: true,
                  type: "enum",
                  default: "s",
                  documentation: [
                    {
                      text: "Supported Epoch Layouts",
                      url: "https://github.com/open-telemetry/opentelemetry-collector-contrib/blob/main/pkg/stanza/docs/types/timestamp.md#parse-a-timestamp-using-an-epoch-layout",
                    },
                  ],
                  relevantIf: [
                    {
                      name: "parse_timestamp",
                      operator: "equals",
                      value: true,
                    },
                    {
                      name: "parse_timestamp_format",
                      operator: "equals",
                      value: "Epoch",
                    },
                    {
                      name: "parse_format",
                      operator: "notEquals",
                      value: "none",
                    },
                  ],
                  advancedConfig: false,
                  validValues: ["s", "ms", "us", "ns", "s.ms", "s.us", "s.ns"],
                  options: DEFAULT_PARAMETER_OPTIONS,
                },
                {
                  label: "Timestamp Layout",
                  name: "manual_timestamp_format",
                  description: "The strptime layout of the timestamp.",
                  required: true,
                  type: "string",
                  default: "%Y-%m-%dT%H:%M:%S.%f%z",
                  documentation: [
                    {
                      text: "Supported Layout Directives",
                      url: "https://github.com/observiq/ctimefmt/blob/3e07deba22cf7a753f197ef33892023052f26614/ctimefmt.go#L63",
                    },
                  ],
                  relevantIf: [
                    {
                      name: "parse_timestamp",
                      operator: "equals",
                      value: true,
                    },
                    {
                      name: "parse_timestamp_format",
                      operator: "equals",
                      value: "Manual",
                    },
                    {
                      name: "parse_format",
                      operator: "notEquals",
                      value: "none",
                    },
                  ],
                  advancedConfig: false,
                  validValues: null,
                  options: DEFAULT_PARAMETER_OPTIONS,
                },
                {
                  label: "Parse Severity",
                  name: "parse_severity",
                  description: "Whether to parse severity from the log entry.",
                  required: false,
                  type: "bool",
                  default: false,
                  documentation: null,
                  relevantIf: [
                    {
                      name: "parse_format",
                      operator: "notEquals",
                      value: "none",
                    },
                  ],
                  advancedConfig: false,
                  validValues: null,
                  options: DEFAULT_PARAMETER_OPTIONS,
                },
                {
                  label: "Severity Field",
                  name: "severity_field",
                  description:
                    "The field containing the severity in the log entry.",
                  required: true,
                  type: "string",
                  default: "severity",
                  documentation: null,
                  relevantIf: [
                    {
                      name: "parse_severity",
                      operator: "equals",
                      value: true,
                    },
                    {
                      name: "parse_format",
                      operator: "notEquals",
                      value: "none",
                    },
                  ],
                  advancedConfig: false,
                  validValues: null,
                  options: DEFAULT_PARAMETER_OPTIONS,
                },
                {
                  label: "Include File Name Attribute",
                  name: "include_file_name_attribute",
                  description:
                    'Whether to add the file name as the attribute "log.file.name".',
                  required: false,
                  type: "bool",
                  default: true,
                  documentation: null,
                  relevantIf: null,
                  advancedConfig: true,
                  validValues: null,
                  options: DEFAULT_PARAMETER_OPTIONS,
                },
                {
                  label: "Include File Path Attribute",
                  name: "include_file_path_attribute",
                  description:
                    'Whether to add the file path as the attribute "log.file.path".',
                  required: false,
                  type: "bool",
                  default: false,
                  documentation: null,
                  relevantIf: null,
                  advancedConfig: true,
                  validValues: null,
                  options: DEFAULT_PARAMETER_OPTIONS,
                },
                {
                  label: "Include File Name Resolved Attribute",
                  name: "include_file_name_resolved_attribute",
                  description:
                    'Whether to add the file name after symlinks resolution as the attribute "log.file.name_resolved".',
                  required: false,
                  type: "bool",
                  default: false,
                  documentation: null,
                  relevantIf: null,
                  advancedConfig: true,
                  validValues: null,
                  options: DEFAULT_PARAMETER_OPTIONS,
                },
                {
                  label: "Include File Path Resolved Attribute",
                  name: "include_file_path_resolved_attribute",
                  description:
                    'Whether to add the file path after symlinks resolution as the attribute "log.file.path_resolved".',
                  required: false,
                  type: "bool",
                  default: false,
                  documentation: null,
                  relevantIf: null,
                  advancedConfig: true,
                  validValues: null,
                  options: DEFAULT_PARAMETER_OPTIONS,
                },
                {
                  label: "Encoding",
                  name: "encoding",
                  description: "The encoding of the file being read.",
                  required: false,
                  type: "enum",
                  default: "utf-8",
                  documentation: null,
                  relevantIf: null,
                  advancedConfig: true,
                  validValues: [
                    "nop",
                    "utf-8",
                    "utf-16le",
                    "utf-16be",
                    "ascii",
                    "big5",
                  ],
                  options: DEFAULT_PARAMETER_OPTIONS,
                },
                {
                  label: "Offset Storage Directory",
                  name: "offset_storage_dir",
                  description:
                    "The directory that the offset storage file will be created.",
                  required: false,
                  type: "string",
                  default: "$OIQ_OTEL_COLLECTOR_HOME/storage",
                  documentation: null,
                  relevantIf: null,
                  advancedConfig: true,
                  validValues: null,
                  options: DEFAULT_PARAMETER_OPTIONS,
                },
                {
                  label: "Filesystem Poll Interval",
                  name: "poll_interval",
                  description:
                    "The duration of time in milliseconds between filesystem polls.",
                  required: false,
                  type: "int",
                  default: 200,
                  documentation: null,
                  relevantIf: null,
                  advancedConfig: true,
                  validValues: null,
                  options: DEFAULT_PARAMETER_OPTIONS,
                },
                {
                  label: "Max Concurrent Files",
                  name: "max_concurrent_files",
                  description:
                    "The maximum number of log files from which logs will be read concurrently. If the number of files matched exceeds this number, then files will be processed in batches.",
                  required: false,
                  type: "int",
                  default: 1024,
                  documentation: null,
                  relevantIf: null,
                  advancedConfig: true,
                  validValues: null,
                  options: DEFAULT_PARAMETER_OPTIONS,
                },
                {
                  label: "Parse To",
                  name: "parse_to",
                  description: "The field to which the log will be parsed.",
                  required: false,
                  type: "enum",
                  default: "body",
                  documentation: null,
                  relevantIf: null,
                  advancedConfig: true,
                  validValues: ["body", "attributes"],
                  options: DEFAULT_PARAMETER_OPTIONS,
                },
                {
                  label: "Start At",
                  name: "start_at",
                  description: "Start reading logs from 'beginning' or 'end'.",
                  required: false,
                  type: "enum",
                  default: "end",
                  documentation: null,
                  relevantIf: null,
                  advancedConfig: true,
                  validValues: ["beginning", "end"],
                  options: DEFAULT_PARAMETER_OPTIONS,
                },
              ],
            },
          },
        },
      };
    },
  },
];

export const redisSourceTypeQuery: MockedResponse<Record<string, any>>[] = [
  {
    request: {
      operationName: "SourceType",
      query: SourceTypeDocument,
      variables: { name: "redis" },
    },
    result: () => {
      return {
        data: {
          sourceType: {
            apiVersion: "bindplane.observiq.com/v1",
            kind: "SourceType",
            metadata: {
              id: "252423b1-e3de-4e35-b6b6-f1ecffa66106",
              version: 1,
              name: "redis",
              displayName: "Redis",
              description: "Collect metrics and logs from Redis.",
              icon: "/icons/sources/redis.svg",
              additionalInfo: null,
              __typename: "Metadata",
            },
            spec: {
              parameters: [
                {
                  name: "enable_metrics",
                  label: "Enable Metrics",
                  description: "Enable to collect metrics.",
                  relevantIf: null,
                  documentation: null,
                  advancedConfig: false,
                  required: false,
                  type: "bool",
                  validValues: null,
                  default: true,
                  options: DEFAULT_PARAMETER_OPTIONS,
                  __typename: "ParameterDefinition",
                },
                {
                  name: "endpoint",
                  label: "Endpoint",
                  description: "The endpoint of the Redis server.",
                  relevantIf: [
                    {
                      name: "enable_metrics",
                      operator: "equals",
                      value: true,
                      __typename: "RelevantIfCondition",
                    },
                  ],
                  documentation: null,
                  advancedConfig: false,
                  required: false,
                  type: "string",
                  validValues: null,
                  default: "localhost:6379",
                  options: DEFAULT_PARAMETER_OPTIONS,
                  __typename: "ParameterDefinition",
                },
                {
                  name: "transport",
                  label: "Transport",
                  description:
                    "The transport protocol being used to connect to Redis.",
                  relevantIf: [
                    {
                      name: "enable_metrics",
                      operator: "equals",
                      value: true,
                      __typename: "RelevantIfCondition",
                    },
                  ],
                  documentation: null,
                  advancedConfig: false,
                  required: false,
                  type: "enum",
                  validValues: ["tcp", "unix"],
                  default: "tcp",
                  options: DEFAULT_PARAMETER_OPTIONS,
                  __typename: "ParameterDefinition",
                },
                {
                  name: "password",
                  label: "Password",
                  description:
                    "The password used to access the Redis instance; must match the password specified in the requirepass server configuration option.",
                  relevantIf: [
                    {
                      name: "enable_metrics",
                      operator: "equals",
                      value: true,
                      __typename: "RelevantIfCondition",
                    },
                  ],
                  documentation: null,
                  advancedConfig: true,
                  required: false,
                  type: "string",
                  validValues: null,
                  default: "",
                  options: DEFAULT_PARAMETER_OPTIONS,
                  __typename: "ParameterDefinition",
                },
                {
                  name: "collection_interval",
                  label: "Collection Interval",
                  description: "How often (seconds) to scrape for metrics.",
                  relevantIf: [
                    {
                      name: "enable_metrics",
                      operator: "equals",
                      value: true,
                      __typename: "RelevantIfCondition",
                    },
                  ],
                  documentation: null,
                  advancedConfig: true,
                  required: false,
                  type: "int",
                  validValues: null,
                  default: 60,
                  options: DEFAULT_PARAMETER_OPTIONS,
                  __typename: "ParameterDefinition",
                },
                {
                  name: "enable_tls",
                  label: "Enable TLS",
                  description: "Whether or not to use TLS.",
                  relevantIf: [
                    {
                      name: "enable_metrics",
                      operator: "equals",
                      value: true,
                      __typename: "RelevantIfCondition",
                    },
                  ],
                  documentation: null,
                  advancedConfig: true,
                  required: false,
                  type: "bool",
                  validValues: null,
                  default: false,
                  options: DEFAULT_PARAMETER_OPTIONS,
                  __typename: "ParameterDefinition",
                },
                {
                  name: "insecure_skip_verify",
                  label: "Skip TLS Certificate Verification",
                  description: "Enable to skip TLS certificate verification.",
                  relevantIf: [
                    {
                      name: "enable_tls",
                      operator: "equals",
                      value: true,
                      __typename: "RelevantIfCondition",
                    },
                  ],
                  documentation: null,
                  advancedConfig: true,
                  required: false,
                  type: "bool",
                  validValues: null,
                  default: false,
                  options: DEFAULT_PARAMETER_OPTIONS,
                  __typename: "ParameterDefinition",
                },
                {
                  name: "ca_file",
                  label: "TLS Certificate Authority File",
                  description:
                    "Certificate authority used to validate TLS certificates.",
                  relevantIf: [
                    {
                      name: "enable_tls",
                      operator: "equals",
                      value: true,
                      __typename: "RelevantIfCondition",
                    },
                  ],
                  documentation: null,
                  advancedConfig: true,
                  required: false,
                  type: "string",
                  validValues: null,
                  default: "",
                  options: DEFAULT_PARAMETER_OPTIONS,
                  __typename: "ParameterDefinition",
                },
                {
                  name: "cert_file",
                  label: "Mutual TLS Client Certificate File",
                  description:
                    "A TLS certificate used for client authentication, if mutual TLS is enabled.",
                  relevantIf: [
                    {
                      name: "enable_tls",
                      operator: "equals",
                      value: true,
                      __typename: "RelevantIfCondition",
                    },
                  ],
                  documentation: null,
                  advancedConfig: true,
                  required: false,
                  type: "string",
                  validValues: null,
                  default: "",
                  options: DEFAULT_PARAMETER_OPTIONS,
                  __typename: "ParameterDefinition",
                },
                {
                  name: "key_file",
                  label: "Mutual TLS Client Private Key File",
                  description:
                    "A TLS private key used for client authentication, if mutual TLS is enabled.",
                  relevantIf: [
                    {
                      name: "enable_tls",
                      operator: "equals",
                      value: true,
                      __typename: "RelevantIfCondition",
                    },
                  ],
                  documentation: null,
                  advancedConfig: true,
                  required: false,
                  type: "string",
                  validValues: null,
                  default: "",
                  options: DEFAULT_PARAMETER_OPTIONS,
                  __typename: "ParameterDefinition",
                },
                {
                  name: "enable_logs",
                  label: "Enable Logs",
                  description: "Enable to collect logs.",
                  relevantIf: null,
                  documentation: null,
                  advancedConfig: false,
                  required: false,
                  type: "bool",
                  validValues: null,
                  default: true,
                  options: DEFAULT_PARAMETER_OPTIONS,
                  __typename: "ParameterDefinition",
                },
                {
                  name: "file_path",
                  label: "Log Paths",
                  description: "Path to Redis log file(s).",
                  relevantIf: [
                    {
                      name: "enable_logs",
                      operator: "equals",
                      value: true,
                      __typename: "RelevantIfCondition",
                    },
                  ],
                  documentation: null,
                  advancedConfig: true,
                  required: false,
                  type: "strings",
                  validValues: null,
                  default: [
                    "/var/log/redis/redis-server.log",
                    "/var/log/redis_6379.log",
                    "/var/log/redis/redis.log",
                    "/var/log/redis/default.log",
                    "/var/log/redis/redis_6379.log",
                  ],
                  options: DEFAULT_PARAMETER_OPTIONS,
                  __typename: "ParameterDefinition",
                },
                {
                  name: "start_at",
                  label: "Start At",
                  description: "Start reading logs from 'beginning' or 'end'.",
                  relevantIf: [
                    {
                      name: "enable_logs",
                      operator: "equals",
                      value: true,
                      __typename: "RelevantIfCondition",
                    },
                  ],
                  documentation: null,
                  advancedConfig: true,
                  required: false,
                  type: "enum",
                  validValues: ["beginning", "end"],
                  default: "end",
                  options: DEFAULT_PARAMETER_OPTIONS,
                  __typename: "ParameterDefinition",
                },
              ],
              supportedPlatforms: [],
              version: "0.0.1",
              telemetryTypes: ["logs", "metrics"],
              __typename: "ResourceTypeSpec",
            },
            __typename: "SourceType",
          },
        },
      };
    },
  },
];

export const testConfig: MinimumRequiredConfig = {
  metadata: {
    id: "test",
    name: "test",
    description: "",
    version: 1,
    labels: {
      platform: "macos",
    },
    __typename: "Metadata",
  },
  spec: {
    raw: "",
    sources: [
      {
        type: "file",
        name: "",
        displayName: "file display name",
        parameters: [
          {
            name: "file_path",
            value: ["/tmp/test.log"],
            __typename: "Parameter",
          },
          {
            name: "exclude_file_path",
            value: [],
            __typename: "Parameter",
          },
          {
            name: "log_type",
            value: "file",
            __typename: "Parameter",
          },
          {
            name: "parse_format",
            value: "none",
            __typename: "Parameter",
          },
          {
            name: "regex_pattern",
            value: "",
            __typename: "Parameter",
          },
          {
            name: "multiline_line_start_pattern",
            value: "",
            __typename: "Parameter",
          },
          {
            name: "encoding",
            value: "utf-8",
            __typename: "Parameter",
          },
          {
            name: "start_at",
            value: "end",
            __typename: "Parameter",
          },
        ],
        processors: null,
        disabled: false,
        __typename: "ResourceConfiguration",
      },
      {
        type: "redis",
        name: "",
        displayName: "redis display name",
        parameters: [
          {
            name: "enable_metrics",
            value: true,
            __typename: "Parameter",
          },
          {
            name: "endpoint",
            value: "localhost:6379",
            __typename: "Parameter",
          },
          {
            name: "transport",
            value: "tcp",
            __typename: "Parameter",
          },
          {
            name: "disable_metrics",
            value: [],
            __typename: "Parameter",
          },
          {
            name: "password",
            value: "",
            __typename: "Parameter",
          },
          {
            name: "collection_interval",
            value: 10,
            __typename: "Parameter",
          },
          {
            name: "enable_tls",
            value: false,
            __typename: "Parameter",
          },
          {
            name: "insecure_skip_verify",
            value: false,
            __typename: "Parameter",
          },
          {
            name: "ca_file",
            value: "",
            __typename: "Parameter",
          },
          {
            name: "cert_file",
            value: "",
            __typename: "Parameter",
          },
          {
            name: "key_file",
            value: "",
            __typename: "Parameter",
          },
          {
            name: "enable_logs",
            value: true,
            __typename: "Parameter",
          },
          {
            name: "file_path",
            value: [
              "/var/log/redis/redis-server.log",
              "/var/log/redis_6379.log",
              "/var/log/redis/redis.log",
              "/var/log/redis/default.log",
              "/var/log/redis/redis_6379.log",
            ],
            __typename: "Parameter",
          },
          {
            name: "start_at",
            value: "end",
            __typename: "Parameter",
          },
        ],
        processors: null,
        disabled: true,
        __typename: "ResourceConfiguration",
      },
    ],
    destinations: [],
    selector: {
      matchLabels: {
        configuration: "test",
      },
      __typename: "AgentSelector",
    },
    __typename: "ConfigurationSpec",
  },
  graph: {
    attributes: {
      activeTypeFlags: 1,
    },
    sources: [
      {
        id: "source/source0",
        type: "sourceNode",
        label: "file",
        attributes: {
          activeTypeFlags: 1,
          kind: "Source",
          resourceId: "source0",
          supportedTypeFlags: 1,
        },
        __typename: "Node",
      },
      {
        id: "source/source1",
        type: "sourceNode",
        label: "redis",
        attributes: {
          activeTypeFlags: 0,
          kind: "Source",
          resourceId: "source1",
          supportedTypeFlags: 0,
        },
        __typename: "Node",
      },
    ],
    intermediates: [
      {
        id: "source/source0/processors",
        type: "processorNode",
        label: "Processors",
        attributes: {
          activeTypeFlags: 1,
          supportedTypeFlags: 1,
        },
        __typename: "Node",
      },
      {
        id: "source/source1/processors",
        type: "processorNode",
        label: "Processors",
        attributes: {
          activeTypeFlags: 0,
          supportedTypeFlags: 0,
        },
        __typename: "Node",
      },
      {
        id: "destination/otlphttp/processors",
        type: "processorNode",
        label: "Processors",
        attributes: {
          activeTypeFlags: 0,
          supportedTypeFlags: 0,
        },
        __typename: "Node",
      },
      {
        id: "destination/otlphttp-2/processors",
        type: "processorNode",
        label: "Processors",
        attributes: {
          activeTypeFlags: 1,
          supportedTypeFlags: 0,
        },
        __typename: "Node",
      },
    ],
    targets: [
      {
        id: "destination/otlphttp",
        type: "destinationNode",
        label: "otlphttp",
        attributes: {
          activeTypeFlags: 0,
          isInline: false,
          kind: "Destination",
          resourceId: "otlphttp",
          supportedTypeFlags: 0,
        },
        __typename: "Node",
      },
      {
        id: "destination/otlphttp-2",
        type: "destinationNode",
        label: "otlphttp-2",
        attributes: {
          activeTypeFlags: 1,
          isInline: false,
          kind: "Destination",
          resourceId: "otlphttp-2",
          supportedTypeFlags: 0,
        },
        __typename: "Node",
      },
    ],
    edges: [
      {
        id: "source/source0|source/source0/processors",
        source: "source/source0",
        target: "source/source0/processors",
        __typename: "Edge",
      },
      {
        id: "source/source1|source/source1/processors",
        source: "source/source1",
        target: "source/source1/processors",
        __typename: "Edge",
      },
      {
        id: "source/source0/processors|destination/otlphttp/processors",
        source: "source/source0/processors",
        target: "destination/otlphttp/processors",
        __typename: "Edge",
      },
      {
        id: "source/source1/processors|destination/otlphttp/processors",
        source: "source/source1/processors",
        target: "destination/otlphttp/processors",
        __typename: "Edge",
      },
      {
        id: "destination/otlphttp/processors|destination/otlphttp",
        source: "destination/otlphttp/processors",
        target: "destination/otlphttp",
        __typename: "Edge",
      },
      {
        id: "source/source0/processors|destination/otlphttp-2/processors",
        source: "source/source0/processors",
        target: "destination/otlphttp-2/processors",
        __typename: "Edge",
      },
      {
        id: "source/source1/processors|destination/otlphttp-2/processors",
        source: "source/source1/processors",
        target: "destination/otlphttp-2/processors",
        __typename: "Edge",
      },
      {
        id: "destination/otlphttp-2/processors|destination/otlphttp-2",
        source: "destination/otlphttp-2/processors",
        target: "destination/otlphttp-2",
        __typename: "Edge",
      },
    ],
    __typename: "Graph",
  },
  __typename: "Configuration",
};

export const mockedDestinationAndTypeResponse: MockedResponse<GetDestinationWithTypeQuery> =
  {
    request: {
      query: GetDestinationWithTypeDocument,
      variables: {
        name: destination0Name,
      },
    },
    result: {
      data: {
        destinationWithType: {
          destination: destination0,
          destinationType: destination0Type,
        },
      },
    },
  };

export const mockedDestationAndTypeResponse_PAUSED: MockedResponse<GetDestinationWithTypeQuery> =
  {
    request: {
      query: GetDestinationWithTypeDocument,
      variables: {
        name: destination0Name,
      },
    },
    result: {
      data: {
        destinationWithType: {
          destination: destination0_PAUSED,
          destinationType: destination0Type,
        },
      },
    },
  };

export const mockedSourceAndTypeResponse: MockedResponse<GetSourceWithTypeQuery> =
  {
    request: {
      query: GetSourceWithTypeDocument,
      variables: {
        name: source1Name,
      },
    },
    result: {
      data: {
        sourceWithType: {
          source: source1,
          sourceType: source1Type,
        },
      },
    },
  };

export const mockedSourceAndTypeResponse_PAUSED: MockedResponse<GetSourceWithTypeQuery> =
  {
    request: {
      query: GetSourceWithTypeDocument,
      variables: {
        name: source1Name,
      },
    },
    result: {
      data: {
        sourceWithType: {
          source: source1_PAUSED,
          sourceType: source1Type,
        },
      },
    },
  };
