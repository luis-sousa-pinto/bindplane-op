import { FormValues, satisfiesRelevantIf } from ".";
import {
  ParameterDefinition,
  Parameter,
  ParameterType,
} from "../../graphql/generated";
import { validateNameField } from "../../utils/forms/validate-name-field";
import {
  validateAWSNamedField,
  validateFileLogSortField,
  validateIntField,
  validateMapField,
  validateStringField,
  validateStringsField,
  validateYamlField,
} from "./validation-functions";

export function initFormValues(
  definitions: ParameterDefinition[],
  parameters?: Parameter[] | null,
  includeNameField?: boolean,
  displayName?: string
): FormValues {
  // Assign defaults
  let defaults: FormValues = {};
  if (includeNameField) {
    defaults.name = "";
  }
  if (displayName) {
    defaults.displayName = displayName;
  }
  for (const definition of definitions) {
    defaults[definition.name] = definition.default;
  }

  // Override with existing values if present
  if (parameters != null) {
    for (const parameter of parameters) {
      defaults[parameter.name] = parameter.value;
    }
  }

  return defaults;
}

/**
 * Check for errors in the form based on the current values and definitions
 *
 * @param definitions       ParameterDefinitions for the form
 * @param initValues        Form values
 * @param kind              Resource kind
 * @param includeNameField  Whether to include the name field
 * @param existingNames     Existing resource names to validate against
 * @returns Errors that should be displayed for each field
 */
export function initFormErrors(
  definitions: ParameterDefinition[],
  initValues: Record<string, any>,
  kind: "processor" | "destination" | "source",
  includeNameField?: boolean,
  existingNames?: string[]
): Record<string, null | string> {
  const initErrors: Record<string, null | string> = {};

  if (includeNameField) {
    initErrors.name = validateNameField(initValues.name, kind, existingNames);
  }

  for (const definition of definitions) {
    if (!satisfiesRelevantIf(initValues, definition)) {
      initErrors[definition.name] = null;
      continue;
    }

    switch (definition.type) {
      case ParameterType.MapToEnum:
      case ParameterType.Map:
        initErrors[definition.name] = validateMapField(
          initValues[definition.name],
          definition.required
        );
        break;

      case ParameterType.String:
        initErrors[definition.name] = validateStringField(
          initValues[definition.name],
          definition.required
        );
        break;
      case ParameterType.Strings:
        initErrors[definition.name] = validateStringsField(
          initValues[definition.name],
          definition.required
        );
        break;

      case ParameterType.Yaml:
        initErrors[definition.name] = validateYamlField(
          initValues[definition.name],
          definition.required
        );
        break;

      case ParameterType.Int:
        initErrors[definition.name] = validateIntField(
          definition,
          initValues[definition.name]
        );
        break;

      case ParameterType.AwsCloudwatchNamedField:
        initErrors[definition.name] = validateAWSNamedField(
          initValues[definition.name]
        );
        break;

      case ParameterType.FileLogSort:
        initErrors[definition.name] = validateFileLogSortField(
          initValues[definition.name]
        );
        break;

      default:
        initErrors[definition.name] = null;
    }
  }
  return initErrors;
}
