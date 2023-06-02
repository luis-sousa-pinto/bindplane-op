import {
  FormHelperText,
  IconButton,
  InputAdornment,
  Stack,
  TextField,
} from "@mui/material";
import { isEmpty, isFunction } from "lodash";
import { ChangeEvent, memo, useState } from "react";
import { EyeIcon, EyeOffIcon } from "../../Icons";
import { validateStringField } from "../validation-functions";
import { useValidationContext } from "../ValidationContext";
import { ParamInputProps } from "./ParameterInput";

const StringParamInputComponent: React.FC<ParamInputProps<string>> = ({
  definition,
  value,
  readOnly,
  onValueChange,
}) => {
  const { errors, setError, touched, touch } = useValidationContext();
  const [hideField, setHideField] = useState(
    definition.options.password === true
  );

  function handleValueChange(e: ChangeEvent<HTMLInputElement>) {
    const newValue = e.target.value;
    isFunction(onValueChange) && onValueChange(e.target.value);

    if (!touched[definition.name]) {
      touch(definition.name);
    }

    setError(
      definition.name,
      validateStringField(newValue, definition.required)
    );
  }

  return (
    <TextField
      multiline={definition.options.multiline ?? undefined}
      type={hideField ? "password" : "text"}
      InputProps={{
        endAdornment: definition.options.password && (
          <HideFieldToggle
            hideField={hideField}
            onToggle={() => {
              setHideField(!hideField);
            }}
          />
        ),
        autoComplete: definition.options.password ? "new-password" : "off",
      }}
      value={value}
      onChange={handleValueChange}
      disabled={readOnly}
      onBlur={() => touch(definition.name)}
      name={definition.name}
      fullWidth
      size="small"
      label={definition.label}
      FormHelperTextProps={{
        sx: { paddingLeft: "-2px" },
      }}
      helperText={
        <>
          {errors[definition.name] && touched[definition.name] && (
            <>
              <FormHelperText sx={{ marginLeft: 0 }} component="span" error>
                {errors[definition.name]}
              </FormHelperText>
              <br />
            </>
          )}

          {!isEmpty(definition.description) && (
            <FormHelperText sx={{ marginLeft: 0 }} component="span">
              {definition.description}
            </FormHelperText>
          )}
          {definition.documentation && (
            <Stack component={"span"}>
              {definition.documentation.map((d) => (
                <a href={d.url} rel="noreferrer" target="_blank">
                  {d.text}
                </a>
              ))}
            </Stack>
          )}
        </>
      }
      required={definition.required}
      autoComplete="off"
      autoCorrect="off"
      autoCapitalize="off"
      spellCheck="false"
    />
  );
};

interface HideFieldToggleProps {
  hideField: boolean;
  onToggle: () => void;
}
const HideFieldToggle: React.FC<HideFieldToggleProps> = ({
  hideField,
  onToggle,
}) => {
  return (
    <InputAdornment position="end">
      <IconButton onClick={onToggle}>
        {hideField ? <EyeIcon /> : <EyeOffIcon />}
      </IconButton>
    </InputAdornment>
  );
};

export const StringParamInput = memo(StringParamInputComponent);
