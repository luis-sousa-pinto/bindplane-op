import { memo } from "react";
import { TitleSection } from "./TitleSection";
import { DialogContent } from "@mui/material";
import { ActionsSection } from './ActionsSection';

interface Props {
  title?: string;
  description?: string;
  onClose: () => void;
  buttons?: React.ReactNode;
  children: JSX.Element;
}

const DialogContainerComponent: React.FC<Props> = ({
  title,
  description,
  onClose,
  buttons,
  children,
}) => {
  return (
    <>
      <TitleSection title={title} description={description} onClose={onClose} />
      <DialogContent dividers={!!buttons}>{children}</DialogContent>
      {buttons && <ActionsSection>{buttons}</ActionsSection>}
    </>
  );
};

export const DialogContainer = memo(DialogContainerComponent);
