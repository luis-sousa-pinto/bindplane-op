import { RolloutOptions } from "../../graphql/generated";

// NOTE: These options are the same as the defaults in the backend in configuration.go
const DEFAULT_OPTIONS: RolloutOptions = {
  maxErrors: 0,
  phaseAgentCount: {
    initial: 3,
    multiplier: 5,
    maximum: 100,
  },
  rollbackOnFailure: false,
  startAutomatically: false,
};

export async function startRollout(name: string, options?: RolloutOptions) {
  const resp = await fetch(`/v1/rollouts/${name}/start`, {
    method: "POST",
    body: JSON.stringify({ options: options ?? DEFAULT_OPTIONS }),
  });

  if (!resp.ok) {
    throw new Error(`Got unexpected status: ${resp.status}`);
  }
}

export async function pauseRollout(name: string) {
  const resp = await fetch(`/v1/rollouts/${name}/pause`, {
    method: "PUT",
  });

  if (!resp.ok) {
    throw new Error(`Got unexpected status: ${resp.status}`);
  }
}

export async function resumeRollout(name: string) {
  const resp = await fetch(`/v1/rollouts/${name}/resume`, {
    method: "PUT",
  });

  if (!resp.ok) {
    throw new Error(`Got unexpected status: ${resp.status}`);
  }
}
