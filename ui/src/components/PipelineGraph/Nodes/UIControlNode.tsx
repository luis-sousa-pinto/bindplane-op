import { memo } from "react";
import { AddResourceCard } from "../../Cards/AddResourceCard";
import { Handle, Position } from "reactflow";
import { Button } from "@mui/material";
import { PlusCircleIcon } from "../../Icons";

import mixins from "../../../styles/mixins.module.scss";

function UIControlNode({
  data,
}: {
  data: {
    onClick: (value: boolean) => void;
    buttonText: string;
    handlePosition: Position;
    handleType: "source" | "target";
    isButton: boolean;
  };
}) {
  return (
    <>
      {data.isButton ? (
        <Button
          onClick={() => data.onClick(true)}
          variant="contained"
          startIcon={<PlusCircleIcon />}
          classes={{ root: mixins["float-right"] }}
        >
          {data.buttonText}
        </Button>
      ) : (
        <>
          <AddResourceCard
            onClick={() => data.onClick(true)}
            buttonText={data.buttonText}
          />
          <Handle type={data.handleType} position={data.handlePosition} />
        </>
      )}
    </>
  );
}

export default memo(UIControlNode);
