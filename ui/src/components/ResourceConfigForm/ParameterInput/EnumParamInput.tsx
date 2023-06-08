import {
  TextField,
  createFilterOptions,
  Autocomplete,
  Stack,
  Typography,
} from "@mui/material";
import { isFunction } from "lodash";
import { ChangeEvent, memo } from "react";
import { ParamInputProps } from "./ParameterInput";

import colors from "../../../styles/colors";

const EnumParamInputComponent: React.FC<ParamInputProps<string>> = (props) => {
  return props.definition.options.creatable ? (
    <CreatableSelectInput {...props} />
  ) : (
    <SelectParamInput {...props} />
  );
};

const SelectParamInput: React.FC<ParamInputProps<string>> = ({
  definition,
  value,
  readOnly,
  onValueChange,
}) => {
  return (
    <TextField
      value={value}
      onChange={
        isFunction(onValueChange)
          ? (e: ChangeEvent<HTMLInputElement>) => onValueChange(e.target.value)
          : undefined
      }
      disabled={readOnly}
      name={definition.name}
      fullWidth
      size="small"
      label={definition.label}
      helperText={
        <Typography
          component={"span"}
          whiteSpace={"pre-wrap"}
          fontSize="0.75rem"
          color={readOnly ? colors.disabled : undefined}
        >
          {definition.description}
          {definition.documentation && (
            // Block display so that the link doesn't span the whole length of the outer div.
            <Stack component={"span"} style={{ display: "block" }}>
              {definition.documentation?.map((d) => (
                <a href={d.url} rel="noreferrer" target="_blank" key={d.url}>
                  {d.text}
                </a>
              ))}
            </Stack>
          )}
        </Typography>
      }
      required={definition.required}
      select
      SelectProps={{ native: true }}
    >
      {definition.validValues?.map((v) => (
        <option key={v} value={v}>
          {v}
        </option>
      ))}
    </TextField>
  );
};

const filter = createFilterOptions<string>();

const CreatableSelectInput: React.FC<ParamInputProps<string>> = ({
  definition,
  value,
  onValueChange,
}) => {
  return (
    <Autocomplete
      value={value}
      onChange={(event, newValue) => {
        if (newValue && isFunction(onValueChange)) {
          onValueChange(newValue);
        }
      }}
      filterOptions={(options, params) => {
        const filtered = filter(options, params);

        const { inputValue } = params;
        // Suggest the creation of a new value
        const isExisting = options.some((option) => inputValue === option);
        if (inputValue !== "" && !isExisting) {
          filtered.push(inputValue);
        }

        return filtered;
      }}
      selectOnFocus
      clearOnBlur
      handleHomeEndKeys
      options={definition.validValues ?? []}
      getOptionLabel={(option) => option}
      renderOption={(props, option) => <li {...props}>{option}</li>}
      freeSolo
      renderInput={(params) => (
        <TextField
          {...params}
          fullWidth
          label={definition.label}
          helperText={definition.description}
          required={definition.required}
          size="small"
        />
      )}
    />
  );
};

export const EnumParamInput = memo(EnumParamInputComponent);
