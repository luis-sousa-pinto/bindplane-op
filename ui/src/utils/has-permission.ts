import { Roles } from "../contexts/RBAC";
import { Role } from "../graphql/generated";

/**
 * hasPermission returns true if the user has the required role or greater.
 * @param requiredRole
 * @param testRole
 * @returns boolean
 */
export function hasPermission(requiredRole: Role, testRole: Roles): boolean {
  if (testRole === "single-user") {
    return true;
  }

  if (testRole === Role.Admin) {
    return true;
  }

  if (testRole === Role.User) {
    return requiredRole === Role.User || requiredRole === Role.Viewer;
  }

  return false;
}
