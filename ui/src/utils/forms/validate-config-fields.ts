import { isEmpty } from "lodash";
import { GetConfigNamesQuery } from "../../graphql/generated";
import { RawConfigFormErrors, RawConfigFormValues } from "../../types/forms";
import { validateNameField } from "./validate-name-field";

export function validateFields(
  formValues: RawConfigFormValues,
  configurations?: GetConfigNamesQuery["configurations"]["configurations"],
  secondaryRequired?: boolean
): RawConfigFormErrors {
  const { name, platform, secondaryPlatform } = formValues;
  const errors: RawConfigFormErrors = {
    name: null,
    platform: null,
    secondaryPlatform: null,
    description: null,
    fileName: null,
    rawConfig: null,
  };

  // Validate the Name field
  errors.name = validateNameField(
    name,
    "configuration",
    configurations?.map((c) => c.metadata.name)
  );

  // Validate the platform field
  if (isEmpty(platform)) {
    errors.platform = "Required.";
  }
  if (secondaryRequired && isEmpty(secondaryPlatform)) {
    errors.secondaryPlatform = "Required.";
  }
  return errors;
}
