import { intersection, isArray, isEqual, get } from "lodash";
import {
  ParameterDefinition,
  RelevantIfOperatorType,
} from "../../graphql/generated";

// Helper functions to perform checks
function isEqualToValue(formValue: any, conditionValue: any): boolean {
  return isEqual(formValue, conditionValue);
}

/**
 * Check if form values satisfy the relevantIf conditions of a ParameterDefinition,
 * if any.
 */
export function satisfiesRelevantIf(
  formValues: { [name: string]: any },
  definition: ParameterDefinition
): boolean {
  const relevantIf = definition.relevantIf;

  if (relevantIf == null) {
    return true;
  }

  for (const condition of relevantIf) {
    const formValue = get(formValues, condition.name);
    switch (condition.operator) {
      case RelevantIfOperatorType.Equals:
        if (!isEqualToValue(formValue, condition.value)) {
          return false;
        }
        break;

      case RelevantIfOperatorType.NotEquals:
        if (isEqualToValue(formValue, condition.value)) {
          return false;
        }
        break;

      case RelevantIfOperatorType.ContainsAny:
        if (isArray(formValue)) {
          if (intersection(formValue, condition.value).length === 0) {
            return false;
          }
        }
        break;
    }
  }

  return true;
}
