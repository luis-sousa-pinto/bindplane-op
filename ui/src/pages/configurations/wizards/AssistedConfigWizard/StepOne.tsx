import { Typography } from "@mui/material";
import React from "react";
import { Link } from "react-router-dom";

import { StepOneCommon } from "../StepOneCommon";

import mixins from "../../../../styles/mixins.module.scss";

export const StepOne: React.FC = () => {
  function renderCopy() {
    return (
      <>
        <Typography variant="h6" classes={{ root: mixins["mb-5"] }}>
          Let's get started building your configuration
        </Typography>

        <Typography variant="body2" classes={{ root: mixins["mb-3"] }}>
          The BindPlane configuration builder makes it easy to assemble a valid
          OpenTelemetry config.
        </Typography>

        <Typography variant="body2" classes={{ root: mixins["mb-3"] }}>
          Already have a configuration? Use our{" "}
          <Link to="/configurations/new-raw">raw configuration wizard</Link>.
        </Typography>

        <Typography variant="body2" classes={{ root: mixins["mb-3"] }}>
          We&apos;ll walk you through configuring the data providers you want to
          ingest logs / metrics from and the destination you want to send the
          data to.
        </Typography>

        <Typography variant="body2" classes={{ root: mixins["mb-3"] }}>
          At the end, you’ll have a valid YAML file you can download directly or
          you can use BindPlane to quickly apply the config to one or more of
          your agents.
        </Typography>

        <Typography variant="body2" classes={{ root: mixins["mb-3"] }}>
          {" "}
          Let&apos;s get started importing your config.
        </Typography>
      </>
    );
  }
  return (
    <>
      <StepOneCommon renderCopy={renderCopy} />
    </>
  );
};
