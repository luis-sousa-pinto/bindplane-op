import { useContext } from "react";
import { RBACContext, Roles } from "../contexts/RBAC";

/**
 * useRole returns the role value of the RBACContext.
 */
export function useRole(): Roles {
  const value = useContext(RBACContext);
  return value.role;
}
