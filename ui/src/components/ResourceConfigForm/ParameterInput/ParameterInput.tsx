import { Divider, Grid, Typography } from "@mui/material";
import { useMemo } from "react";
import {
  BoolParamInput,
  EnumParamInput,
  EnumsParamInput,
  IntParamInput,
  MapParamInput,
  MetricsParamInput,
  StringParamInput,
  StringsParamInput,
  TimezoneParamInput,
  YamlParamInput,
  FileLogSortInput,
} from ".";
import { ParameterDefinition, ParameterType } from "../../../graphql/generated";
import { useResourceFormValues } from "../ResourceFormContext";
import { AWSCloudwatchInput } from "./AWSCloudwatchFieldInput";

export interface ParamInputProps<T> {
  definition: ParameterDefinition;
  value?: T;
  onValueChange?: (v: T) => void;
  readOnly?: boolean;
}

export const ParameterInput: React.FC<{
  definition: ParameterDefinition;
  readOnly?: boolean;
}> = ({ definition, readOnly }) => {
  const { formValues, setFormValues } = useResourceFormValues();
  const onValueChange = useMemo(
    () => (newValue: any) => {
      setFormValues((prev) => ({ ...prev, [definition.name]: newValue }));
    },
    [definition.name, setFormValues]
  );

  const subHeader: JSX.Element | null = useMemo(() => {
    if (definition.options.subHeader) {
      return (
        <Grid item xs={12}>
          <Typography>{definition.options.subHeader}</Typography>
        </Grid>
      );
    }

    return null;
  }, [definition.options.subHeader]);

  const divider: JSX.Element | null = useMemo(() => {
    if (definition.options.horizontalDivider) {
      return (
        <Grid item xs={12}>
          <Divider />
        </Grid>
      );
    }
    return null;
  }, [definition.options.horizontalDivider]);

  const Control: JSX.Element = useMemo(() => {
    switch (definition.type) {
      case ParameterType.String:
        return (
          <StringParamInput
            definition={definition}
            value={formValues[definition.name]}
            onValueChange={onValueChange}
            readOnly={readOnly}
          />
        );

      case ParameterType.Strings:
        return (
          <StringsParamInput
            definition={definition}
            value={formValues[definition.name]}
            onValueChange={onValueChange}
            readOnly={readOnly}
          />
        );

      case ParameterType.Enum:
        return (
          <EnumParamInput
            definition={definition}
            value={formValues[definition.name]}
            onValueChange={onValueChange}
            readOnly={readOnly}
          />
        );

      case ParameterType.Enums:
        return (
          <EnumsParamInput
            definition={definition}
            value={formValues[definition.name]}
            onValueChange={onValueChange}
            readOnly={readOnly}
          />
        );

      case ParameterType.Bool:
        return (
          <BoolParamInput
            definition={definition}
            value={formValues[definition.name]}
            onValueChange={onValueChange}
            readOnly={readOnly}
          />
        );

      case ParameterType.Int:
        return (
          <IntParamInput
            definition={definition}
            value={formValues[definition.name]}
            onValueChange={onValueChange}
            readOnly={readOnly}
          />
        );

      case ParameterType.Map:
        return (
          <MapParamInput
            definition={definition}
            value={formValues[definition.name]}
            onValueChange={onValueChange}
            readOnly={readOnly}
          />
        );

      case ParameterType.Yaml:
        return (
          <YamlParamInput
            definition={definition}
            value={formValues[definition.name]}
            onValueChange={onValueChange}
            readOnly={readOnly}
          />
        );

      case ParameterType.Timezone:
        return (
          <TimezoneParamInput
            definition={definition}
            value={formValues[definition.name]}
            onValueChange={onValueChange}
            readOnly={readOnly}
          />
        );

      case ParameterType.Metrics:
        return (
          <MetricsParamInput
            definition={definition}
            value={formValues[definition.name]}
            onValueChange={onValueChange}
            readOnly={readOnly}
          />
        );

      case ParameterType.AwsCloudwatchNamedField:
        return (
          <AWSCloudwatchInput
            definition={definition}
            value={formValues[definition.name]}
            onValueChange={onValueChange}
            readOnly={readOnly}
          />
        );
      case ParameterType.FileLogSort:
        return (
          <FileLogSortInput
            definition={definition}
            value={formValues[definition.name]}
            onValueChange={onValueChange}
            readOnly={readOnly}
          />
        );
      case ParameterType.MapToEnum:
        return (
          <MapParamInput
            definition={definition}
            value={formValues[definition.name]}
            onValueChange={onValueChange}
            readOnly={readOnly}
          />
        );
    }
  }, [definition, formValues, onValueChange, readOnly]);

  const gridColumns = useMemo(() => {
    if (isMetricsType(definition)) {
      return 12;
    }

    if (isTelemetryHeader(definition)) {
      return 12;
    }

    if (isSectionHeader(definition)) {
      return 12;
    }

    if (isAWSCloudwatch(definition)) {
      return 12;
    }

    if (isFileLogSort(definition)) {
      return 12;
    }

    return definition.options.gridColumns ?? 6;
  }, [definition]);

  return (
    <>
      {subHeader}
      <Grid item xs={gridColumns}>
        {Control}
      </Grid>
      {divider}
    </>
  );
};

function isTelemetryHeader(definition: ParameterDefinition) {
  return ["enable_metrics", "enable_logs", "enable_traces"].includes(
    definition.name
  );
}

function isSectionHeader(definition: ParameterDefinition) {
  return definition.options.sectionHeader === true;
}

function isMetricsType(definition: ParameterDefinition) {
  return definition.type === "metrics";
}

function isAWSCloudwatch(definition: ParameterDefinition) {
  return definition.type === "awsCloudwatchNamedField";
}

function isFileLogSort(definition: ParameterDefinition) {
  return definition.type === "fileLogSort";
}
