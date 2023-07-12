import { Grid, Typography } from "@mui/material";
import { Parameter, ParameterDefinition } from "../../../graphql/generated";
import { useResourceFormValues } from "../../ResourceConfigForm/ResourceFormContext";
import { ResourceDisplayNameInput } from "../../ResourceConfigForm/ParameterInput/ResourceDisplayNameInput";
import { ViewHeading } from "./ViewHeading";
import { ParameterSection } from "../../ResourceConfigForm/ParameterSection";
import { groupParameters } from "../../ResourceConfigForm/ConfigureResourceView";
import { useMemo } from "react";

import mixins from "../../../styles/mixins.module.scss";

interface Props {
  title?: string;
  description?: string;
  parameterDefinitions: ParameterDefinition[];
  parameters?: Parameter[];
}

export const ProcessorForm: React.FC<Props> = ({
  title,
  description,
  parameterDefinitions,
}) => {
  const { formValues, setFormValues } = useResourceFormValues();
  const groups = useMemo(
    () => groupParameters(parameterDefinitions),
    [parameterDefinitions]
  );

  return (
    <>
      <Grid container spacing={3} className={mixins["mb-3"]}>
        <Grid item xs={12}>
          <ViewHeading heading={title} subHeading={description} />
        </Grid>
      </Grid>
      <form style={{ flexGrow: 1, overflow: "auto" }}>
        <Grid container spacing={3} className={mixins["mb-5"]}>
          <Grid item xs={7}>
            <ResourceDisplayNameInput
              value={formValues.displayName}
              onValueChange={(v: string) =>
                setFormValues((prev) => ({ ...prev, displayName: v }))
              }
            />
          </Grid>
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
                  readOnly={false}
                />
              ))}
            </>
          )}
        </Grid>
      </form>
    </>
  );
};
