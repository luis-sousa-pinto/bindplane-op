import { Role } from "../../graphql/generated";
import { useRole } from "../../hooks/useRole";
import { hasPermission } from "../../utils/has-permission";

interface RBACWrapperProps {
  requiredRole: Role;
}

/**
 * RBACWrapper is a wrapper component that will only render its children if the
 * user has the required role.
 */
export const RBACWrapper: React.FC<RBACWrapperProps> = ({
  children,
  requiredRole,
}) => {
  const role = useRole();

  if (hasPermission(requiredRole, role)) {
    return <>{children}</>;
  }

  return null;
};
