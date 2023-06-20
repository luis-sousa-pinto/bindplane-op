import {
  FormControl,
  MenuItem,
  Select,
  SelectChangeEvent,
  SelectProps,
  InputLabel,
  FormHelperText,
} from "@mui/material";
import React, { useState } from "react";

import styles from "./platform-select.module.scss";
import { isEmpty } from "lodash";

interface Platform {
  label: string;
  value: string;
  backgroundImage: string;
  secondarySelections?: Platform[];
}

const PLATFORMS: Platform[] = [
  {
    label: "Linux",
    value: "linux",
    backgroundImage: "url('/icons/linux-platform-icon.svg",
  },
  {
    label: "Kubernetes",
    value: "kubernetes",
    backgroundImage: "url('/icons/kubernetes-platform-icon.svg",
    secondarySelections: [
      {
        label: "Kubernetes (DaemonSet)",
        value: "kubernetes-daemonset",
        backgroundImage: "url('/icons/kubernetes-platform-icon.svg",
      },
      {
        label: "Kubernetes (Deployment)",
        value: "kubernetes-deployment",
        backgroundImage: "url('/icons/kubernetes-platform-icon.svg",
      },
    ],
  },
  {
    label: "macOS",
    value: "macos",
    backgroundImage: "url('/icons/macos-platform-icon.svg",
  },
  {
    label: "OpenShift",
    value: "openshift",
    backgroundImage: "url('/icons/openshift-platform-icon.svg",
    secondarySelections: [
      {
        label: "OpenShift (DaemonSet)",
        value: "openshift-daemonset",
        backgroundImage: "url('/icons/openshift-platform-icon.svg",
      },
      {
        label: "OpenShift (Deployment)",
        value: "openshift-deployment",
        backgroundImage: "url('/icons/openshift-platform-icon.svg",
      },
    ],
  },
  {
    label: "Windows",
    value: "windows",
    backgroundImage: "url('/icons/windows-platform-icon.svg",
  },
];

interface PlatformSelectProps extends SelectProps {
  onPlatformSelected: (value: string) => void;
  onSecondaryPlatformSelected: (value: string) => void;
  helperText?: string | null;
  secondaryHelperText?: string | null;
  platformValue?: string | null;
  secondaryPlatformValue?: string | null;
  setSecondaryPlatformRequired: (value: boolean) => void;
  onSecondaryBlur?: () => void;
  secondaryError?: boolean;
}

export const PlatformSelect: React.FC<PlatformSelectProps> = ({
  onPlatformSelected,
  onSecondaryPlatformSelected,
  size,
  error,
  secondaryError,
  helperText,
  secondaryHelperText,
  platformValue,
  secondaryPlatformValue,
  setSecondaryPlatformRequired,
  onSecondaryBlur,
  ...rest
}) => {
  const [platform, setPlatform] = useState<Platform | null>(
    PLATFORMS.find((p) => p.value === platformValue) ?? null
  );
  const [secondaryPlatform, setSecondaryPlatform] = useState<Platform | null>(
    PLATFORMS.find((p) => p.value === platformValue)?.secondarySelections?.find(
      (s) => s.value === secondaryPlatformValue
    ) ?? null
  );

  function handleSelect(e: SelectChangeEvent<unknown>) {
    const value = e.target.value as string;
    const newPlatform = PLATFORMS.find((p) => p.value === value)!;
    setPlatform(newPlatform);
    setSecondaryPlatformRequired(!isEmpty(newPlatform.secondarySelections));
    onPlatformSelected(value);
    setSecondaryPlatform(null);
  }
  function handleSecondarySelect(e: SelectChangeEvent<unknown>) {
    const value = e.target.value as string;
    setSecondaryPlatform(
      platform?.secondarySelections?.find((p) => p.value === value)!
    );
    onSecondaryPlatformSelected(value);
  }

  return (
    <>
      <FormControl
        fullWidth
        margin="normal"
        variant="outlined"
        error={error}
        classes={{ root: styles.root }}
        size={size}
      >
        <InputLabel id="platform-label">Platform</InputLabel>
        <Select
          labelId="platform-label"
          id="platform"
          label="Platform"
          onChange={handleSelect}
          value={platformValue}
          startAdornment={
            platform ? (
              <span
                style={{
                  backgroundImage: platform.backgroundImage,
                }}
                className={styles["value-icon"]}
              />
            ) : undefined
          }
          inputProps={{
            "data-testid": "platform-select-input",
          }}
          size={size}
          {...rest}
        >
          {PLATFORMS.map((p) => (
            <MenuItem
              key={p.value}
              value={p.value}
              classes={{ root: styles.item }}
            >
              <span
                style={{ backgroundImage: p.backgroundImage }}
                className={styles.icon}
              ></span>
              {p.label}
            </MenuItem>
          ))}
        </Select>
        <FormHelperText>{helperText}</FormHelperText>
      </FormControl>

      {/* Create another drop down for a secondary selection */}
      {platform?.secondarySelections && (
        <>
          <FormControl
            fullWidth
            margin="normal"
            variant="outlined"
            error={secondaryError}
            classes={{ root: styles.root }}
            size={size}
          >
            <InputLabel id="platform-secondary-label">
              Platform Specifics
            </InputLabel>
            <Select
              labelId="platform-secondary-label"
              id="platform-secondary"
              label="Platform Specifics"
              onChange={handleSecondarySelect}
              value={secondaryPlatformValue}
              startAdornment={
                secondaryPlatform ? (
                  <span
                    style={{
                      backgroundImage: secondaryPlatform.backgroundImage,
                    }}
                    className={styles["value-icon"]}
                  />
                ) : undefined
              }
              inputProps={{
                "data-testid": "platform-secondary-select-input",
              }}
              size={size}
              onBlur={onSecondaryBlur}
            >
              {platform?.secondarySelections.map((p) => (
                <MenuItem
                  key={p.value}
                  value={p.value}
                  classes={{ root: styles.item }}
                >
                  <span
                    style={{ backgroundImage: p.backgroundImage }}
                    className={styles.icon}
                  ></span>
                  {p.label}
                </MenuItem>
              ))}
            </Select>
            <FormHelperText>{secondaryHelperText}</FormHelperText>
          </FormControl>
        </>
      )}
    </>
  );
};
