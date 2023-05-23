import {
  InputLabel,
  FormHelperText,
  Stack,
  FormControlLabel,
  Switch,
  Typography,
} from "@mui/material";
import { ParamInputProps } from "./ParameterInput";
import { classes } from "../../../utils/styles";
import { isEmpty } from "lodash";
import { memo } from "react";
import colors from "../../../styles/colors";

import mixins from "../../../styles/mixins.module.scss";

const EnumsParamInputComponent: React.FC<ParamInputProps<string[]>> = ({
  definition,
  value,
  readOnly,
  onValueChange,
}) => {
  function handleToggleValue(toggleValue: string) {
    const newValue = [...(value ?? [])];

    if (!newValue.includes(toggleValue)) {
      // Make sure that toggleValue is in new value array
      newValue.push(toggleValue);
    } else {
      // Remove the toggle value from the array
      const atIndex = newValue.findIndex((v) => v === toggleValue);
      if (atIndex > -1) {
        newValue.splice(atIndex, 1);
      }
    }

    onValueChange && onValueChange(newValue);
  }

  return (
    <>
      <InputLabel>{definition.label}</InputLabel>

      <FormHelperText
        className={classes([
          isEmpty(definition.description) ? undefined : mixins["mb-1"],
        ])}
      >
        <Typography
          color={readOnly ? colors.disabled : undefined}
          component={"span"}
          whiteSpace={"pre-wrap"}
          fontSize="0.75rem"
        >
          {definition.description}
        </Typography>
      </FormHelperText>
      <Stack marginLeft={2}>
        {definition.validValues!.map((vv) => (
          <FormControlLabel
            key={`${definition.name}-label-${vv}`}
            disabled={readOnly}
            control={
              <Switch
                key={`${definition.name}-switch-${vv}`}
                size="small"
                onChange={() => handleToggleValue(vv)}
                checked={
                  definition.options.trackUnchecked
                    ? !value?.includes(vv)
                    : value?.includes(vv)
                }
              />
            }
            label={vv}
          />
        ))}
      </Stack>
    </>
  );
};

export const EnumsParamInput = memo(EnumsParamInputComponent);
