import {
  DialogTitle,
  IconButton,
  Stack,
  Typography,
  Alert,
} from "@mui/material";
import { memo } from "react";
import { XIcon } from "../Icons";
import { AdditionalInfo } from "../../graphql/generated";

import styles from "./dialog-components.module.scss";

interface Props {
  description?: string;
  title?: string;
  additionalInfo?: AdditionalInfo | null;
  onClose: () => void;
}

const TitleSectionComponent: React.FC<Props> = ({
  description,
  title,
  additionalInfo,
  onClose,
}) => {
  return (
    <DialogTitle
      classes={{
        root: styles.title,
      }}
    >
      <Stack direction="row" justifyContent="space-between" alignItems="center">
        <Stack>
          <Typography fontSize={28} fontWeight={600}>
            {title}
          </Typography>

          <Typography fontSize={18}>{description}</Typography>
        </Stack>

        <IconButton className={styles.close} onClick={onClose}>
          <XIcon strokeWidth={"3"} width={"28"} />
        </IconButton>
      </Stack>
      {additionalInfo && (
        <Alert
          severity="info"
          className={styles["info"]}
          data-testid="info-alert"
        >
          <Typography>
            {additionalInfo.message}
            {additionalInfo.documentation?.map((d) => (
              <a href={d.url} rel="noreferrer" target="_blank" key={d.url}>
                {d.text}
              </a>
            ))}
          </Typography>
        </Alert>
      )}
    </DialogTitle>
  );
};

export const TitleSection = memo(TitleSectionComponent);
