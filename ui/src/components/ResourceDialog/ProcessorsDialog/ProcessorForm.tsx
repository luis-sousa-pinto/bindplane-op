import { Grid, Typography } from "@mui/material";
import { Parameter, ParameterDefinition } from "../../../graphql/generated";
import { ParameterInput } from "../../ResourceConfigForm/ParameterInput";
import { useResourceFormValues } from "../../ResourceConfigForm/ResourceFormContext";
import { satisfiesRelevantIf } from "../../ResourceConfigForm/satisfiesRelevantIf";
import { ResourceDisplayNameInput } from "../../ResourceConfigForm/ParameterInput/ResourceDisplayNameInput";
import { ViewHeading } from "./ViewHeading";

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
          {parameterDefinitions.length === 0 ? (
            <Grid item>
              <Typography>No additional configuration needed.</Typography>
            </Grid>
          ) : (
            parameterDefinitions.map((p) => {
              if (satisfiesRelevantIf(formValues, p)) {
                return <ParameterInput key={p.name} definition={p} />;
              }

              return null;
            })
          )}
        </Grid>
      </form>
    </>
  );
};
