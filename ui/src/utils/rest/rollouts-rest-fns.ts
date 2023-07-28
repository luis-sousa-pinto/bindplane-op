import { RolloutOptions } from "../../graphql/generated";

export async function startRollout(name: string, options?: RolloutOptions) {
  const resp = await fetch(`/v1/rollouts/${name}/start`, {
    method: "POST",
    body: JSON.stringify({ options: options }),
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
