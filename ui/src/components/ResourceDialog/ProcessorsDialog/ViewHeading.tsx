import { Stack, Typography } from "@mui/material";

interface ViewHeadingProps {
  heading?: string;
  subHeading?: string;
}

export const ViewHeading: React.FC<ViewHeadingProps> = ({
  heading,
  subHeading,
}) => {
  return (
    <Stack paddingBottom={1}>
      {heading && (
        <Typography fontWeight={600} fontSize={24}>
          {heading}
        </Typography>
      )}
      {subHeading && <Typography fontSize={16}>{subHeading}</Typography>}
    </Stack>
  );
};
