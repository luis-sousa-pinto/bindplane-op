import { Grid, Typography } from "@mui/material";
import { Parameter, ParameterDefinition } from "../../../graphql/generated";
import { ParameterInput } from "../../ResourceConfigForm/ParameterInput";
import { useResourceFormValues } from "../../ResourceConfigForm/ResourceFormContext";
import { satisfiesRelevantIf } from "../../ResourceConfigForm/satisfiesRelevantIf";
import { ResourceDisplayNameInput } from "../../ResourceConfigForm/ParameterInput/ResourceDisplayNameInput";

import mixins from "../../../styles/mixins.module.scss";
import { ViewHeading } from './ViewHeading';

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
    <form>
      <Grid container spacing={3} className={mixins["mb-5"]}>
        <Grid item xs={12}>
          <ViewHeading heading={title} subHeading={description} />
        </Grid>
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
  );
};
