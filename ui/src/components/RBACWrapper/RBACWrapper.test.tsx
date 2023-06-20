import { render, screen } from "@testing-library/react";
import { RBACWrapper } from "./RBACWrapper";
import { Button } from "@mui/material";
import { Role } from "../../graphql/generated";
import { RBACContext } from "../../contexts/RBAC";

describe("RBACWrapper", () => {
  it("should render children when the requiredRole matches the current role", async () => {
    render(
      <RBACContext.Provider
        value={{
          role: Role.Admin,
        }}
      >
        <RBACWrapper requiredRole={Role.User}>
          <Button>Test button</Button>
        </RBACWrapper>
      </RBACContext.Provider>
    );

    screen.getByText("Test button");
  });

  it("should not render children when the requiredRole does not match the current role", async () => {
    render(
      <RBACContext.Provider value={{ role: Role.User }}>
        <RBACWrapper requiredRole={Role.Admin}>
          <Button>Test button</Button>
        </RBACWrapper>
      </RBACContext.Provider>
    );

    expect(screen.queryByText("Test button")).not.toBeInTheDocument();
  });
});
