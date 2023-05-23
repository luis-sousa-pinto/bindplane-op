import { gql } from "@apollo/client";
import { Box, Button, CircularProgress, Typography } from "@mui/material";
import { useSnackbar } from "notistack";
import { useEffect, useMemo, useState } from "react";
import { ProcessorType } from "../../ResourceConfigForm";
import { useGetProcessorTypesQuery } from "../../../graphql/generated";
import { metadataSatisfiesSubstring } from "../../../utils/metadata-satisfies-substring";
import { ActionsSection } from "../ActionSection";
import { ContentSection } from "../ContentSection";
import { TitleSection } from "../TitleSection";
import {
  ResourceTypeButton,
  ResourceTypeButtonContainer,
} from "../../ResourceTypeButton";

import styles from "./create-processor-select-view.module.scss"

gql`
  query getProcessorTypes {
    processorTypes {
      metadata {
        id
        displayName
        description
        name
        labels
      }
      spec {
        parameters {
          label
          name
          description
          required
          type
          default
          relevantIf {
            name
            operator
            value
          }
          documentation {
            text
            url
          }
          advancedConfig
          validValues
          options {
            creatable
            trackUnchecked
            gridColumns
            sectionHeader
            multiline
            labels
            metricCategories {
              label
              column
              metrics {
                name
                description
                kpi
              }
            }
            password
          }
          documentation {
            text
            url
          }
        }
        telemetryTypes
      }
    }
  }
`;

interface CreateProcessorSelectViewProps {
  // The display name of the Source or Destination that the processor is being added to
  displayName: string;
  // The supported telemetry types of the source that the processor will be added to
  telemetryTypes?: string[];

  onBack?: () => void;
  onSelect: (pt: ProcessorType) => void;
  onClose: () => void;
}

export const CreateProcessorSelectView: React.FC<CreateProcessorSelectViewProps> =
  ({ displayName, onBack, onSelect, telemetryTypes, onClose }) => {
    const { data, loading, error } = useGetProcessorTypesQuery();
    const [search, setSearch] = useState("");
    const { enqueueSnackbar } = useSnackbar();

    useEffect(() => {
      if (error != null) {
        enqueueSnackbar("Error retrieving data for Processor Type.", {
          variant: "error",
          key: "Error retrieving data for Processor Type.",
        });
      }
    }, [enqueueSnackbar, error]);

    const backButton: JSX.Element = (
      <Button variant="contained" color="secondary" onClick={onBack}>
        Back
      </Button>
    );

    const title = `${displayName}: Add a processor`;
    const description = `Choose a processor you'd like to configure for this source.`;

    // Filter the list of supported processor types down
    // to those whose telemetry matches the telemetry of the
    // source. i.e. don't show a log processor for a metric source
    const supportedProcessorTypes: ProcessorType[] = useMemo(
      () =>
        telemetryTypes
          ? data?.processorTypes.filter((pt) =>
            pt.spec.telemetryTypes.some((t) => telemetryTypes.includes(t))
          ) ?? []
          : data?.processorTypes ?? [],
      [data?.processorTypes, telemetryTypes]
    );

    // Filter the list of supported processor types down to those matching the search,
    // and sort them in alphabetical order by display name
    const matchingProcessorTypes: ProcessorType[] = useMemo(
      () =>
        supportedProcessorTypes
          .filter((pt) => metadataSatisfiesSubstring(pt, search))
          .sort((a, b) =>
            (a.metadata.displayName?.toLowerCase() ?? "").localeCompare(
              b.metadata.displayName?.toLowerCase() ?? ""
            )
          ),
      [supportedProcessorTypes, search]
    );
    const categorizedProcessorTypes = processorTypesByCategory(matchingProcessorTypes);

    return (
      <>
        <TitleSection
          title={title}
          description={description}
          onClose={onClose}
        />

        <ContentSection>
          <ResourceTypeButtonContainer
            onSearchChange={(v: string) => setSearch(v)}
            placeholder={"Search for a processor..."}
          >
            {loading && (
              <Box display="flex" justifyContent={"center"} marginTop={2}>
                <CircularProgress />
              </Box>
            )}
            {Object.keys(categorizedProcessorTypes)
              .sort((a, b) => a.localeCompare(b))
              .filter((k) => k !== "Advanced")
              .map((k) => (
                <ProcessorCategory
                  key={k}
                  title={k}
                  processors={categorizedProcessorTypes[k]}
                  onSelect={onSelect}
                />
              ))}
            {categorizedProcessorTypes["Advanced"] && (
              <ProcessorCategory
                key="Advanced"
                title="Advanced"
                processors={categorizedProcessorTypes["Advanced"]}
                onSelect={onSelect}
              />
            )}
          </ResourceTypeButtonContainer>
        </ContentSection>
        {onBack && <ActionsSection>{backButton} </ActionsSection>}
      </>
    );
  };

function processorTypesByCategory(processorTypes: ProcessorType[]): { [category: string]: ProcessorType[] } {
  return processorTypes.reduce((acc: { [key: string]: ProcessorType[] }, p: ProcessorType) => {
    const category: string = p.metadata.labels?.category?.replaceAll("-", " ") ?? "Other";
    if (!acc[category]) {
      acc[category] = [p];
    } else {
      acc[category] = [...acc[category]!, p];
    }

    return acc;
  }, {});
}

interface ProcessorCategoryProps {
  title: string;
  processors: ProcessorType[];
  onSelect: (pt: ProcessorType) => void;
}

function ProcessorCategory({
  onSelect,
  processors,
  title,
}: ProcessorCategoryProps) {
  return (
    <>
      <Box className={styles.category}>
        <Typography fontSize={18} fontWeight={600}>
          {title}
        </Typography>
      </Box>      {processors.map((p) => (
        <ResourceTypeButton
          hideIcon
          key={p.metadata.name}
          displayName={p.metadata.displayName!}
          onSelect={() => onSelect(p)}
          telemetryTypes={p.spec.telemetryTypes}
        />
      ))}
    </>
  );
}
