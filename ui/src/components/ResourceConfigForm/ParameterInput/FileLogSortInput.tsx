import { ParamInputProps } from "./ParameterInput";
import { useValidationContext } from "../ValidationContext";
import {
  Button,
  Stack,
  IconButton,
  FormLabel,
  FormHelperText,
  Grid,
} from "@mui/material";
import {
  ParameterDefinition,
  ParameterType,
  RelevantIfOperatorType,
} from "../../../graphql/generated";
import { PlusCircleIcon, TrashIcon } from "../../Icons";
import { cloneDeep } from "lodash";
import { useMemo, useState } from "react";
import { satisfiesRelevantIf } from "../satisfiesRelevantIf";
import { useResourceFormValues } from "../ResourceFormContext";
import { validateFileLogSortField } from "../validation-functions";
import { StringParamInput } from "./StringParamInput";
import { EnumParamInput } from "./EnumParamInput";
import { TimezoneParamInput } from "./TimezoneParamInput";

export type InputValue = ItemValue[];

interface ItemValue {
  regexKey: string;
  sortType: string;
  sortDirection: string;
  layout: string;
  location: string;
}

export const FileLogSortInput: React.FC<ParamInputProps<InputValue>> = ({
  definition,
  value: paramValue,
  readOnly,
  onValueChange,
}) => {
  const initValue =
    paramValue != null && paramValue.length > 0
      ? paramValue
      : [
          {
            regexKey: "",
            sortType: "numeric",
            sortDirection: "descending",
            layout: "",
            location: "",
          },
        ];
  const [controlValue, setControlValue] = useState<InputValue>(initValue);
  const { errors, touch, touched, setError } = useValidationContext();
  const { formValues } = useResourceFormValues();
  const onFieldValueChange = useMemo(() => {
    return function (v: any, index: number, field: string) {
      const newValue = cloneDeep(controlValue);
      if (!touched[definition.name]) {
        touch(definition.name);
      }

      field = field.replace(index.toString(), "");
      const fieldKey = field as keyof ItemValue;

      newValue[index][fieldKey] = v as string;

      if (fieldKey === "sortType" && v !== "timestamp") {
        newValue[index].layout = "";
        newValue[index].location = "";
      }

      setControlValue(newValue);
      onValueChange && onValueChange(newValue);
      setError(definition.name, validateFileLogSortField(newValue));
    };
  }, [controlValue, onValueChange, setError, definition.name, touch, touched]);

  function addNewField() {
    const defaultItem: ItemValue = {
      regexKey: "",
      layout: "",
      sortType: "numeric",
      location: "",
      sortDirection: "descending",
    };

    const curValue = cloneDeep(controlValue) ?? [];
    curValue.push(defaultItem);
    setControlValue(curValue);
  }

  function removeField(index: number) {
    if (controlValue.length === 1) {
      return;
    }

    const curValue = cloneDeep(controlValue) ?? [];
    curValue.splice(index, 1);
    setControlValue(curValue);
    onValueChange && onValueChange(curValue);
  }

  return (
    <>
      <FormLabel filled={true}>{definition.label}</FormLabel>
      <FormHelperText filled={true}>{definition.description}</FormHelperText>
      {errors[definition.name] && touched[definition.name] && (
        <Stack>
          <FormHelperText sx={{ marginLeft: 0 }} component="span" error>
            {errors[definition.name]}
          </FormHelperText>
        </Stack>
      )}
      {controlValue.map((itemValue, index) => {
        const definitionsArray: ParameterDefinition[] = [
          {
            name: "regexKey" + index,
            label: "Regex Key",
            type: ParameterType.String,
            required: true,
            description: "The name of the regex key to use.",
            default: itemValue.regexKey,
            options: {
              gridColumns: 12,
            },
          },
          {
            name: "sortType" + index,
            label: "Sort Type",
            type: ParameterType.Enum,
            required: true,
            default: itemValue.sortType,
            description: "The type of sort to perform.",
            options: {
              gridColumns: 6,
            },
            validValues: ["numeric", "alphabetical", "timestamp"],
          },
          {
            name: "sortDirection" + index,
            label: "Sort Direction",
            type: ParameterType.Enum,
            required: false,
            description: "",
            default: itemValue.sortDirection,
            options: {
              gridColumns: 6,
            },
            validValues: ["descending", "ascending"],
          },
          {
            name: "layout" + index,
            label: "Layout",
            type: ParameterType.String,
            required: true,
            description:
              "Defines the strptime layout of the timestamp being sorted.",
            options: {
              gridColumns: 6,
            },
            default: itemValue.layout,
            documentation: [
              {
                text: "Supported Layout Directives",
                url: "https://github.com/observiq/ctimefmt/blob/3e07deba22cf7a753f197ef33892023052f26614/ctimefmt.go#L63",
              },
            ],
            relevantIf: [
              {
                operator: RelevantIfOperatorType.Equals,
                name: "sort_rules[" + index + "].sortType",
                // sort_rules[0].sortType
                value: "timestamp",
              },
            ],
          },
          {
            name: "location" + index,
            label: "Timezone",
            type: ParameterType.Timezone,
            required: false,
            description: "The sort timezone location.",
            options: {
              gridColumns: 6,
            },
            default: itemValue.location,
            relevantIf: [
              {
                operator: RelevantIfOperatorType.Equals,
                name: "sort_rules[" + index + "].sortType",
                value: "timestamp",
              },
            ],
          },
        ];

        return (
          <Stack
            key={`filelog-sort-parameter-item-${index}`}
            direction={"row"}
            spacing={2}
            padding={2}
            alignItems={"center"}
            justifyContent="left"
          >
            <Stack
              border="2px solid rgba(0, 0, 0, 0.2)"
              borderRadius={5}
              width={"100%"}
              spacing={2}
              padding={2}
            >
              <Grid container spacing={3}>
                {definitionsArray.map((p) => {
                  const itemValueKey = p.name.replace(index.toString(), "");
                  if (satisfiesRelevantIf(formValues, p)) {
                    const renderSwitch = (itemValueKey: string) => {
                      switch (itemValueKey) {
                        case "regexKey":
                          return (
                            <StringParamInput
                              key={p.name}
                              readOnly={readOnly}
                              definition={p}
                              value={itemValue.regexKey || ""}
                              onValueChange={(v) =>
                                onFieldValueChange(v, index, "regexKey")
                              }
                            />
                          );
                        case "sortType":
                          return (
                            <EnumParamInput
                              key={p.name}
                              readOnly={readOnly}
                              definition={p}
                              value={itemValue.sortType || ""}
                              onValueChange={(v) =>
                                onFieldValueChange(v, index, "sortType")
                              }
                            />
                          );
                        case "sortDirection":
                          return (
                            <EnumParamInput
                              key={p.name}
                              readOnly={readOnly}
                              definition={p}
                              value={itemValue.sortDirection ?? ""}
                              onValueChange={(v) =>
                                onFieldValueChange(v, index, "sortDirection")
                              }
                            />
                          );
                        case "layout":
                          return (
                            <StringParamInput
                              key={p.name}
                              readOnly={readOnly}
                              definition={p}
                              value={itemValue.layout || ""}
                              onValueChange={(v) =>
                                onFieldValueChange(v, index, "layout")
                              }
                            />
                          );
                        case "location":
                          return (
                            <TimezoneParamInput
                              key={p.name}
                              readOnly={readOnly}
                              definition={p}
                              value={itemValue.location || "UTC"}
                              onValueChange={(v) =>
                                onFieldValueChange(v, index, "location")
                              }
                            />
                          );
                        default:
                          return null;
                      }
                    };

                    return (
                      <Grid
                        item
                        xs={p.options.gridColumns || 12}
                        key={"file-sort-grid" + p.name}
                      >
                        {renderSwitch(itemValueKey)}
                      </Grid>
                    );
                  }
                  return null;
                })}
              </Grid>
            </Stack>
            {!(index === 0 && controlValue.length === 1) && (
              <IconButton size={"small"} onClick={() => removeField(index)}>
                <TrashIcon width={18} />
              </IconButton>
            )}
          </Stack>
        );
      })}
      <Stack width={"100%"} alignItems="center">
        <Button
          startIcon={<PlusCircleIcon />}
          onClick={addNewField}
          disabled={readOnly}
        >
          New field
        </Button>
      </Stack>
    </>
  );
};
