import { createContext } from "react";
import { Role } from "../graphql/generated";

export type Roles = Role | "single-user";

interface RBACContextValue {
  role: Roles;
}

const defaultValue: RBACContextValue = {
  role: "single-user",
};

export const RBACContext = createContext(defaultValue);
