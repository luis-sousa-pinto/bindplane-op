import { DialogResource, ResourceType } from ".";

export function someResourceOfType(
  resources: DialogResource[],
  type: ResourceType
): boolean {
  let exists = false;

  for (const resource of resources) {
    const [noVersionType] = resource.spec.type.split(":");
    if (noVersionType === type.metadata.name) {
      exists = true;
      break;
    }
  }

  return exists;
}

export function filterResourcesByType(
  resources: DialogResource[],
  types: ResourceType
): DialogResource[] {
  return resources.filter((resource) => {
    const [noVersionType] = resource.spec.type.split(":");
    return noVersionType === types.metadata.name;
  });
}
