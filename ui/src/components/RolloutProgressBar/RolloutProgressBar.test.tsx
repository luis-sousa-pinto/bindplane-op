import { render } from "@testing-library/react";
import { RolloutProgressBar } from "./RolloutProgressBar";

describe("RolloutProgress", () => {
  it("renders", () => {
    render(
      <RolloutProgressBar
        totalCount={100}
        errors={0}
        completedCount={15}
        rolloutStatus={0}
        hideActions={false}
        paused={false}
        onPause={() => {}}
        onStart={() => {}}
        onResume={() => {}}
      />
    );
  });
});
