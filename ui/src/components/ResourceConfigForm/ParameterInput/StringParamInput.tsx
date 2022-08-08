import { TextField } from "@mui/material";
import { isFunction } from "lodash";
import { ChangeEvent } from "react";
import { ParamInputProps } from "./ParameterInput";

export const StringParamInput: React.FC<ParamInputProps<string>> = ({
  classes,
  definition,
  value,
  onValueChange,
}) => {
  return (
    <TextField
      classes={classes}
      value={value}
      onChange={(e: ChangeEvent<HTMLInputElement>) =>
        isFunction(onValueChange) && onValueChange(e.target.value)
      }
      name={definition.name}
      fullWidth
      size="small"
      label={definition.label}
      helperText={definition.description}
      required={definition.required}
      autoComplete="off"
      autoCorrect="off"
      autoCapitalize="off"
      spellCheck="false"
    />
  );
};