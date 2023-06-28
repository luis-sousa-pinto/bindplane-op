import { Button, Stack } from "@mui/material";
import { ActionsSection } from "../../DialogComponents";
import {
  FormValues,
  initFormValues,
  isValid,
  ProcessorType,
  useValidationContext,
  ValidationContextProvider,
} from "../../ResourceConfigForm";
import { initFormErrors } from "../../ResourceConfigForm/init-form-values";
import {
  FormValueContextProvider,
  useResourceFormValues,
} from "../../ResourceConfigForm/ResourceFormContext";
import { ProcessorForm } from "./ProcessorForm";

import mixins from "../../../styles/mixins.module.scss";

interface CreateProcessorConfigureViewProps {
  processorType: ProcessorType;
  onBack: () => void;
  onSave: (formValues: FormValues) => void;
  onClose: () => void;
}

const CreateProcessorConfigureViewComponent: React.FC<
  CreateProcessorConfigureViewProps
> = ({ processorType, onSave, onBack, onClose }) => {
  const { formValues } = useResourceFormValues();
  const { touchAll, setErrors } = useValidationContext();

  function handleSave() {
    const errors = initFormErrors(
      processorType.spec.parameters,
      formValues,
      "processor"
    );

    if (!isValid(errors)) {
      setErrors(errors);
      touchAll();
      return;
    }

    onSave(formValues);
  }

  return (
    <Stack className={mixins["flex-grow"]}>
      <ProcessorForm
        title={processorType.metadata.displayName ?? ""}
        description={processorType.metadata.description ?? ""}
        parameterDefinitions={processorType.spec.parameters}
      />

      <ActionsSection>
        <Button variant="outlined" color="secondary" onClick={onBack}>
          Cancel
        </Button>

        <Button variant="contained" color="primary" onClick={handleSave}>
          Done
        </Button>
      </ActionsSection>
    </Stack>
  );
};

export const CreateProcessorConfigureView: React.FC<
  CreateProcessorConfigureViewProps
> = (props) => {
  const initValues = initFormValues(
    props.processorType.spec.parameters,
    null,
    false
  );
  const initErrors = initFormErrors(
    props.processorType.spec.parameters,
    initValues,
    "processor"
  );
  return (
    <FormValueContextProvider initValues={initValues}>
      <ValidationContextProvider
        initErrors={initErrors}
        definitions={props.processorType.spec.parameters}
      >
        <CreateProcessorConfigureViewComponent {...props} />
      </ValidationContextProvider>
    </FormValueContextProvider>
  );
};
