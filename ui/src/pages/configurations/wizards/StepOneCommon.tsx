import { Button, Stack, TextField, Typography } from "@mui/material";
import { Box } from "@mui/system";
import React, { ChangeEvent, useEffect, useState } from "react";

import { PlatformSelect } from "../../../components/PlatformSelect";
import { useWizard } from "../../../components/Wizard/WizardContext";
import { useGetConfigNamesQuery } from "../../../graphql/generated";
import { RawConfigFormValues, RawConfigFormErrors } from "../../../types/forms";
import { validateFields } from "../../../utils/forms/validate-config-fields";

import mixins from "../../../styles/mixins.module.scss";
import styles from "./AssistedConfigWizard/assisted-config-wizard.module.scss";
import { gql } from "@apollo/client";

gql`
  query getConfigNames {
    configurations {
      configurations {
        metadata {
          id
          name
          version
        }
      }
    }
  }
`;

interface StepOneProps {
  renderCopy: () => JSX.Element;
}
export const StepOneCommon: React.FC<StepOneProps> = ({ renderCopy }) => {
  const {
    formValues,
    formErrors,
    formTouched,
    setValues,
    setFormErrors,
    setTouched,
    goToStep,
  } = useWizard<RawConfigFormValues>();

  const { data } = useGetConfigNamesQuery();

  const [secondaryPlatformRequired, setSecondaryPlatformRequired] =
    useState(false);
  function handleNameChange(e: ChangeEvent<HTMLInputElement>) {
    setValues({ name: e.target.value });
  }

  function handleSelectChange(v: string) {
    setValues({ platform: v, secondaryPlatform: "" });
    setTouched({ platform: true, secondaryPlatform: false });
  }

  function handleSecondarySelectChange(v: string) {
    setValues({ secondaryPlatform: v });
    setTouched({ secondaryPlatform: true });
  }

  useEffect(() => {
    const errors = validateFields(
      formValues,
      data?.configurations.configurations,
      secondaryPlatformRequired
    );
    setFormErrors((prev) => ({ ...prev, ...errors }));
  }, [
    formValues,
    formTouched,
    secondaryPlatformRequired,
    data?.configurations.configurations,
    setFormErrors,
  ]);

  function handleNextClick() {
    const errors = validateFields(
      formValues,
      data?.configurations.configurations,
      secondaryPlatformRequired
    );
    if (formInvalid(errors)) {
      setTouched({
        name: true,
        platform: true,
        secondaryPlatform: true,
      });
      setFormErrors((prev) => ({ ...prev, ...errors }));
      return;
    }

    goToStep(1);
  }

  return (
    <>
      <div className={styles.container} data-testid="step-one">
        {renderCopy()}
        <Typography
          fontWeight={600}
          variant="subtitle1"
          classes={{ root: mixins["mb-3"] }}
        >
          Config Details
        </Typography>

        <Stack spacing={2} component="form" className={styles.form}>
          <TextField
            autoComplete="off"
            fullWidth
            size="small"
            label="Name"
            name="name"
            id="name"
            error={formErrors.name != null && formTouched.name}
            helperText={formTouched.name ? formErrors.name : null}
            onChange={handleNameChange}
            onBlur={() => setTouched({ name: true })}
            value={formValues.name}
            classes={{ root: mixins["mb-1"] }}
          />

          <PlatformSelect
            size="small"
            name="platform"
            id="platform"
            label="Platform"
            onPlatformSelected={handleSelectChange}
            onSecondaryPlatformSelected={handleSecondarySelectChange}
            platformValue={formValues.platform}
            secondaryPlatformValue={formValues.secondaryPlatform}
            setSecondaryPlatformRequired={setSecondaryPlatformRequired}
            inputProps={{
              "data-testid": "platform-select-input",
            }}
            error={formErrors.platform != null && formTouched.platform}
            secondaryError={
              formErrors.secondaryPlatform != null &&
              formTouched.secondaryPlatform
            }
            helperText={formTouched.platform ? formErrors.platform : null}
            secondaryHelperText={
              formTouched.secondaryPlatform
                ? formErrors.secondaryPlatform
                : null
            }
            onBlur={() => {
              setTouched({ platform: true });
            }}
            onSecondaryBlur={() => {
              setTouched({ secondaryPlatform: true });
            }}
          ></PlatformSelect>

          <TextField
            autoComplete="off"
            fullWidth
            size="small"
            minRows={3}
            multiline
            name="description"
            label="Description"
            value={formValues.description}
            onChange={(e: ChangeEvent<HTMLTextAreaElement>) =>
              setValues({ description: e.target.value })
            }
            onBlur={() => setTouched({ description: true })}
          />
        </Stack>
      </div>
      <Box className={styles.buttons}>
        <div />
        <Button
          variant="contained"
          disabled={
            formTouched.name && formTouched.platform && formInvalid(formErrors)
          }
          onClick={handleNextClick}
          data-testid="step-one-next"
        >
          Next
        </Button>
      </Box>
    </>
  );
  function formInvalid(errors: RawConfigFormErrors): boolean {
    for (const val of Object.values(errors)) {
      if (val != null) {
        return true;
      }
    }

    return false;
  }
};
