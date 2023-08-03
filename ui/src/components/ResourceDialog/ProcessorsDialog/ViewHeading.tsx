import { Chip, Stack, Typography } from "@mui/material";
import { AdditionalInfo } from "../../../graphql/generated";
import { Alert } from "@mui/lab";

import styles from "./create-processor-select-view.module.scss";

interface ViewHeadingProps {
  heading?: string;
  subHeading?: string;
  additionalInfo?: AdditionalInfo | null;
  deprecated?: boolean;
}

export const ViewHeading: React.FC<ViewHeadingProps> = ({
  heading,
  subHeading,
  additionalInfo,
  deprecated,
}) => {
  return (
    <Stack paddingBottom={1}>
      {heading && (
        <Typography fontWeight={600} fontSize={24}>
          {heading}
          {deprecated && (
            <Chip
              color="warning"
              label="Deprecated"
              size="small"
              variant="filled"
              style={{ float: "right" }}
            />
          )}
        </Typography>
      )}
      {subHeading && <Typography fontSize={16}>{subHeading}</Typography>}
      {additionalInfo && (
        <Alert
          severity="info"
          className={styles["info"]}
          data-testid="info-alert"
        >
          <Typography>
            {additionalInfo.message}
            {additionalInfo.documentation?.map((d) => (
              <div key={d.url}>
                <a href={d.url} rel="noreferrer" target="_blank">
                  {d.text}
                </a>
              </div>
            ))}
          </Typography>
        </Alert>
      )}
    </Stack>
  );
};
