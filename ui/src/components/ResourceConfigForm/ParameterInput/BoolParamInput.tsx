import {
  Checkbox,
  FormControlLabel,
  Switch,
  Typography,
  Tooltip,
  Stack,
  Badge,
} from "@mui/material";
import { isEmpty, isFunction } from "lodash";
import { ParamInputProps } from "./ParameterInput";
import { memo } from "react";
import { ParameterDefinition } from "../../../graphql/generated";
import { HelpCircleIcon } from "../../Icons";

import styles from "./parameter-input.module.scss";
import colors from "../../../styles/colors";

const BoolParamInputComponent: React.FC<ParamInputProps<boolean>> = ({
  definition,
  value,
  readOnly,
  onValueChange,
}) => {
  return isTelemetryHeader(definition) ? (
    <TelemetrySectionHeader
      readOnly={readOnly}
      definition={definition}
      value={value}
      onValueChange={onValueChange}
    />
  ) : (
    <FormControlLabel
      control={
        <Checkbox
          onChange={(e) => {
            isFunction(onValueChange) && onValueChange(e.target.checked);
          }}
          disabled={readOnly}
          name={definition.name}
          checked={value}
        />
      }
      label={
        definition.options.sectionHeader ? (
          <Stack direction={"row"} alignItems={"center"}>
            <Typography
              fontSize={18}
              fontWeight={600}
              color={readOnly ? colors.disabled : undefined}
            >
              {definition.label}
            </Typography>
            {!isEmpty(definition.description) && (
              <Tooltip title={definition.description} disableInteractive>
                <Badge sx={{ marginLeft: 2 }} color={"primary"}>
                  <HelpCircleIcon width={19} />
                </Badge>
              </Tooltip>
            )}
            {!isEmpty(definition.documentation) && (
              <>
                {definition.documentation?.map((doc) => (
                  <a
                    target="_blank"
                    rel="noreferrer"
                    href={doc.url}
                    style={{ marginLeft: 18, fontSize: 12 }}
                  >
                    {doc.text}
                  </a>
                ))}
              </>
            )}
          </Stack>
        ) : (
          definition.label
        )
      }
    />
  );
};

const TelemetrySectionHeader: React.FC<ParamInputProps<boolean>> = ({
  definition,
  value,
  readOnly,
  onValueChange,
}) => {
  return (
    <FormControlLabel
      control={
        <Switch
          onChange={(e) => {
            isFunction(onValueChange) && onValueChange(e.target.checked);
          }}
          name={definition.name}
          checked={value}
          classes={{
            root: styles["telemetry-header-switch"],
          }}
          disabled={readOnly}
        />
      }
      label={
        <Typography fontSize={18} fontWeight={600}>
          {definition.label}
        </Typography>
      }
      labelPlacement={"start"}
      classes={{
        root: styles["telemetry-header-root"],
        label: "label",
      }}
    />
  );
};

function isTelemetryHeader(definition: ParameterDefinition) {
  return ["enable_metrics", "enable_logs", "enable_traces"].includes(
    definition.name
  );
}

export const BoolParamInput = memo(BoolParamInputComponent);
