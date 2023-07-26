import { MockedProvider } from "@apollo/client/testing";
import { render, screen } from "@testing-library/react";
import { ConfigureResourceView } from "./ConfigureResourceView";
import { additionalInfo } from "./__test__/dummyResources";

describe("ConfigureResourceView", () => {
  it("supports pausing destinations", () => {
    const onTogglePause = jest.fn();
    render(
      <MockedProvider>
        <ConfigureResourceView
          kind={"destination"}
          resourceTypeDisplayName={"Friendly Name"}
          description={"description"}
          paused={false}
          formValues={{}}
          parameterDefinitions={[]}
          onTogglePause={onTogglePause}
        />
      </MockedProvider>
    );
    const togglePause = screen.getByTestId("resource-form-toggle-pause");
    expect(togglePause.textContent).toBe("Pause");
    togglePause.click();
    expect(onTogglePause).toHaveBeenCalledTimes(1);
  });
  it("supports resuming destinations", () => {
    const onTogglePause = jest.fn();
    render(
      <MockedProvider>
        <ConfigureResourceView
          kind={"destination"}
          resourceTypeDisplayName={"Friendly Name"}
          description={"description"}
          paused={true}
          formValues={{}}
          parameterDefinitions={[]}
          onTogglePause={onTogglePause}
        />
      </MockedProvider>
    );
    const togglePause = screen.getByTestId("resource-form-toggle-pause");
    expect(togglePause.textContent).toBe("Resume");
    togglePause.click();
    expect(onTogglePause).toHaveBeenCalledTimes(1);
  });

  it("renders additionalInfo on source", () => {
    render(
      <MockedProvider>
        <ConfigureResourceView
          kind={"source"}
          resourceTypeDisplayName={"Friendly Name"}
          description={"description"}
          additionalInfo={additionalInfo}
          formValues={{}}
          parameterDefinitions={[]}
        />
      </MockedProvider>
    );
    const infoAlert = screen.getByTestId("info-alert");
    expect(infoAlert.textContent).toBe("test messagetest text");
  });
});
