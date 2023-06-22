import { useContext } from "react";
import { componentContext } from "../contexts/ComponentContext";

export function useComponents() {
  return useContext(componentContext);
}
