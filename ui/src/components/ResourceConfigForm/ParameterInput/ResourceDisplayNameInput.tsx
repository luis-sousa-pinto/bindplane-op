import { FormHelperText, TextField, useTheme } from "@mui/material";
import { isFunction } from "lodash";
import { ChangeEvent, memo, useState } from "react";
import { ParamInputProps } from "./ParameterInput";

interface ResourceDisplayNameInputProps
  extends Omit<ParamInputProps<string>, "definition"> {}

const ResourceDisplayNameInputComponent: React.FC<
  ResourceDisplayNameInputProps
> = ({ value, readOnly, onValueChange }) => {
  const [text, setText] = useState("");
  const maxLength = 40;

  function handleChange(e: ChangeEvent<HTMLInputElement>) {
    if (!isFunction(onValueChange)) {
      return;
    }
    const inputValue = e.target.value;
    if (inputValue.length <= maxLength) {
      setText(inputValue);
    }

    onValueChange(e.target.value);
  }
  const isError = text.length === maxLength;
  const theme = useTheme();
  return (
    <>
      <TextField
        defaultValue={value}
        onChange={handleChange}
        inputProps={{
          "data-testid": "display-name-field",
          maxLength: maxLength,
        }}
        disabled={readOnly}
        name={"displayName"}
        fullWidth
        size="small"
        label={"Short Description"}
        autoComplete="off"
        autoCorrect="off"
        autoCapitalize="off"
        spellCheck="false"
      />
      {isError ? (
        <FormHelperText
          style={{ color: theme.palette.error.main, paddingLeft: "16px" }}
        >
          Character limit reached
        </FormHelperText>
      ) : (
        <FormHelperText style={{ paddingLeft: "16px" }}>
          A friendly name for the resource
        </FormHelperText>
      )}
    </>
  );
};

export const ResourceDisplayNameInput = memo(ResourceDisplayNameInputComponent);
