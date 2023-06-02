/**
 * trimVersion returns the resource name without the version
 * @param name a resource name
 * @returns
 */
export function trimVersion(name: string) {
  return name.split(":")[0];
}

/**
 * asCurrentVersion returns the resource name specified at the current version
 * @param name the resource name
 */
export function asCurrentVersion(name: string) {
  return `${trimVersion(name)}:current`;
}

/**
 * latestVersion returns the resource name specified at the latest version
 * @param name the resource name
 */
export function asLatestVersion(name: string) {
  return `${trimVersion(name)}:latest`;
}

/**
 * nameAndVersion returns the name and version joined with ":".
 * If the version is not specified, the name is returned.
 * @param name the resource name
 * @param version the version number
 */
export function nameAndVersion(name: string, version?: number | string) {
  return version ? `${trimVersion(name)}:${version}` : name;
}
