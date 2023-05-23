import { ComponentMeta, ComponentStory } from "@storybook/react";
import { useState } from "react";
import { MeasurementControlBar } from "./MeasurementControlBar";

export default {
  title: "MeasurementControlBar",
  component: MeasurementControlBar,
} as ComponentMeta<typeof MeasurementControlBar>;

const Template: ComponentStory<typeof MeasurementControlBar> = (args) => {
  const [telemetry, setTelemetry] = useState("logs");
  const [period, setPeriod] = useState("10s");

  function handleTelemetryChange(t: string) {
    setTelemetry(t);
  }

  function handlePeriodChange(p: string) {
    setPeriod(p);
  }

  return (
    <div style={{ width: 1000 }}>
      <MeasurementControlBar
        {...args}
        onTelemetryTypeChange={handleTelemetryChange}
        onPeriodChange={handlePeriodChange}
        telemetry={telemetry}
        period={period}
      />
    </div>
  );
};

export const Default = Template.bind({});

Default.args = {};
