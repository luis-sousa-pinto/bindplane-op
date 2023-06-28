import { Typography } from "@mui/material";
import React from "react";

import { StepOneCommon } from "../StepOneCommon";

import mixins from "../../../../styles/mixins.module.scss";

interface StepOneProps {
  fromImport: boolean;
}

export const StepOne: React.FC<StepOneProps> = ({ fromImport }) => {
  function renderImportCopy() {
    return (
      <>
        <Typography variant="h6" classes={{ root: mixins["mb-5"] }}>
          Let's get started importing your configuration to Bindplane
        </Typography>

        <Typography variant="body2" classes={{ root: mixins["mb-3"] }}>
          We&apos;ve provided some basic details for this configuration, just
          verify everything looks correct.
        </Typography>

        <Typography variant="body2" classes={{ root: mixins["mb-3"] }}>
          When you&apos;re ready click Next to double check the configuration
          Yaml.
        </Typography>
      </>
    );
  }

  function renderStandardCopy() {
    return (
      <>
        <Typography variant="h6" classes={{ root: mixins["mb-5"] }}>
          Let's get started adding your configuration to Bindplane
        </Typography>

        <Typography variant="body2" classes={{ root: mixins["mb-3"] }}>
          An OpenTelemetry configuration is a YAML file that&apos;s used to
          configure your OpenTelemetry collectors. It&apos;s made up of
          receivers, processors, and exporters that are organized into one or
          more data pipelines.
        </Typography>

        <Typography variant="body2" classes={{ root: mixins["mb-3"] }}>
          If you&apos;re not familiar with the structure of these
          configurations, please take a look at our{" "}
          <a
            target="_blank"
            rel="noreferrer"
            href="https://github.com/observIQ/observiq-otel-collector/tree/main/config/google_cloud_exporter"
          >
            sample files
          </a>{" "}
          and the{" "}
          <a
            target="_blank"
            rel="noreferrer"
            href="https://opentelemetry.io/docs/collector/configuration/"
          >
            OpenTelemetry documentation
          </a>
          .
        </Typography>

        <Typography variant="body2" classes={{ root: mixins["mb-3"] }}>
          {" "}
          Let&apos;s get started importing your OpenTelemetry configuration.
        </Typography>
      </>
    );
  }

  return (
    <>
      <StepOneCommon
        renderCopy={fromImport ? renderImportCopy : renderStandardCopy}
      />
    </>
  );
};
