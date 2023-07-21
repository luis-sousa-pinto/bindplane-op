import { DialogActions, Grid, Button, Stack, Typography } from "@mui/material";
import { isFunction } from "lodash";
import { AdditionalInfo, ParameterDefinition } from "../../graphql/generated";
import { ResourceNameInput, useValidationContext, isValid } from ".";
import { useResourceFormValues } from "./ResourceFormContext";
import { useResourceDialog } from "../ResourceDialog/ResourceDialogContext";
import { memo, useMemo } from "react";
import {
  TitleSection,
  ContentSection,
  ActionsSection,
} from "../DialogComponents";
import { initFormErrors } from "./init-form-values";
import { ParameterSection } from "./ParameterSection";
import { PauseIcon, PlayIcon } from "../Icons";
import { ResourceDisplayNameInput } from "./ParameterInput/ResourceDisplayNameInput";

import mixins from "../../styles/mixins.module.scss";
import { ViewHeading } from "../ResourceDialog/ProcessorsDialog/ViewHeading";

export interface ParameterGroup {
  advanced: boolean;
  parameters: ParameterDefinition[];
}

export function groupParameters(
  parameters: ParameterDefinition[]
): ParameterGroup[] {
  const groups: ParameterGroup[] = [];
  let group: ParameterGroup | undefined;

  for (const p of parameters) {
    const advanced = p.advancedConfig ?? false;
    if (group == null || advanced !== group.advanced) {
      // start a new group
      group = {
        advanced,
        parameters: [],
      };
      groups.push(group);
    }
    group.parameters.push(p);
  }

  return groups;
}

interface ConfigureResourceViewProps {
  kind: "source" | "destination" | "processor";
  resourceTypeDisplayName: string;
  description: string;
  additionalInfo?: AdditionalInfo | null;
  formValues: { [key: string]: any };
  includeNameField?: boolean;
  includeDisplayNameField?: boolean;
  displayName?: string;
  existingResourceNames?: string[];
  parameterDefinitions: ParameterDefinition[];
  onBack?: () => void;
  onSave?: (formValues: { [key: string]: any }) => void;
  saveButtonLabel?: string;
  onDelete?: () => void;
  onTogglePause?: () => void;
  disableSave?: boolean;
  paused?: boolean;
  readOnly?: boolean;
  embedded?: boolean;
}

export const ConfigureResourceContent: React.FC<ConfigureResourceViewProps> = ({
  kind,
  resourceTypeDisplayName,
  description,
  additionalInfo,
  formValues,
  includeNameField,
  includeDisplayNameField,
  displayName,
  existingResourceNames,
  parameterDefinitions,
  onBack,
  onSave,
  saveButtonLabel,
  onDelete,
  disableSave,
  onTogglePause,
  paused,
  readOnly,
  embedded,
}) => {
  const { touchAll, setErrors } = useValidationContext();
  const { setFormValues } = useResourceFormValues();
  const { purpose, onClose } = useResourceDialog();

  const groups = useMemo(
    () => groupParameters(parameterDefinitions),
    [parameterDefinitions]
  );

  function handleSubmit() {
    const errors = initFormErrors(
      parameterDefinitions,
      formValues,
      kind,
      includeNameField
    );

    if (!isValid(errors)) {
      touchAll();
      setErrors(errors);
      return;
    }

    isFunction(onSave) && onSave(formValues);
  }

  const primaryButton: JSX.Element = (
    <Button
      disabled={disableSave}
      type="submit"
      variant="contained"
      data-testid="resource-form-save"
      onClick={handleSubmit}
    >
      {saveButtonLabel ?? "Save"}
    </Button>
  );

  const backButton: JSX.Element | null = isFunction(onBack) ? (
    <Button variant="outlined" color="secondary" onClick={onBack}>
      Back
    </Button>
  ) : null;

  const deleteButton: JSX.Element | null = isFunction(onDelete) ? (
    <Button variant="outlined" color="error" onClick={onDelete}>
      Delete
    </Button>
  ) : null;

  const togglePauseButton: JSX.Element | null = isFunction(onTogglePause) ? (
    <Button
      variant="outlined"
      color={paused ? "primary" : "secondary"}
      onClick={onTogglePause}
      data-testid="resource-form-toggle-pause"
    >
      {paused ? "Resume" : "Pause"}
    </Button>
  ) : null;

  const title = useMemo(() => {
    const capitalizedResource = kind[0].toUpperCase() + kind.slice(1);
    const action = purpose === "create" ? "Add" : "Edit";
    return `${action} ${capitalizedResource}: ${resourceTypeDisplayName}`;
  }, [resourceTypeDisplayName, kind, purpose]);

  const ActionsContainer = embedded ? ActionsSection : DialogActions;

  const playPauseButtons =
    !readOnly && togglePauseButton ? (
      <ActionsContainer>
        {paused != null &&
          (paused ? (
            <Button disabled={true} startIcon={<PauseIcon />}>
              Paused
            </Button>
          ) : (
            <Button disabled={true} startIcon={<PlayIcon />}>
              Running
            </Button>
          ))}
        {!readOnly && togglePauseButton}
      </ActionsContainer>
    ) : null;

  const actionButtons = (
    <ActionsContainer
      sx={{
        marginLeft: "auto",
      }}
    >
      {!readOnly && deleteButton}
      {backButton}
      {!readOnly && primaryButton}
    </ActionsContainer>
  );

  const form = (
    <form data-testid="resource-form">
      <Grid container spacing={3} className={mixins["mb-5"]}>
        {includeNameField && (
          <Grid item xs={6}>
            <ResourceNameInput
              readOnly={readOnly}
              kind={kind}
              value={formValues.name}
              onValueChange={(v: string) =>
                setFormValues((prev) => ({ ...prev, name: v }))
              }
              existingNames={existingResourceNames}
            />
          </Grid>
        )}
        <Grid item xs={12}>
          {embedded ? (
            <ViewHeading
              heading={resourceTypeDisplayName}
              subHeading={description}
            />
          ) : (
            <Typography fontWeight={600} fontSize={24}>
              Configure
            </Typography>
          )}
        </Grid>

        {includeDisplayNameField && (
          <Grid item xs={7}>
            <ResourceDisplayNameInput
              readOnly={readOnly}
              value={displayName}
              onValueChange={(v: string) =>
                setFormValues((prev) => ({ ...prev, displayName: v }))
              }
            />
          </Grid>
        )}
        {groups.length === 0 ? (
          <Grid item>
            <Typography>No additional configuration needed.</Typography>
          </Grid>
        ) : (
          <>
            {groups.map((g, ix) => (
              <ParameterSection
                key={`param-group-${ix}`}
                group={g}
                readOnly={readOnly}
              />
            ))}
          </>
        )}
      </Grid>
    </form>
  );

  return (
    <Stack className={mixins["flex-grow"]}>
      {!embedded && (
        <TitleSection
          title={title}
          description={description}
          additionalInfo={additionalInfo}
          onClose={onClose}
        />
      )}

      {embedded ? (
        <Stack className={mixins["flex-grow"]} overflow="auto">
          {form}
        </Stack>
      ) : (
        <ContentSection dividers>{form}</ContentSection>
      )}

      {playPauseButtons ? (
        <Stack direction="row">
          {playPauseButtons}
          {actionButtons}
        </Stack>
      ) : (
        actionButtons
      )}
    </Stack>
  );
};

export const ConfigureResourceView = memo(ConfigureResourceContent);
