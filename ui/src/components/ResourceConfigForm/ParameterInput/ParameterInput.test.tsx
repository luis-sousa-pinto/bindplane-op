import { fireEvent, render, screen } from "@testing-library/react";
import {
  awsCloudWatchFieldInputDef,
  enumDef,
  enumsDef,
  intDef,
  metricParamInput,
  stringDef,
  stringsDef,
  telemetrySectionBoolDef,
  timezoneDef,
  yamlDef,
} from "../__test__/dummyResources";
import { AWSCloudwatchInput } from "./AWSCloudwatchFieldInput";
import { BoolParamInput } from "./BoolParamInput";
import { EnumParamInput } from "./EnumParamInput";
import { EnumsParamInput } from "./EnumsParamInput";
import { IntParamInput } from "./IntParamInput";
import { MapParamInput } from "./MapParamInput";
import { MetricsParamInput } from "./MetricsParamInput";
import { StringParamInput } from "./StringParamInput";
import { StringsParamInput } from "./StringsParamInput";
import { TimezoneParamInput } from "./TimezoneParamInput";
import { YamlParamInput } from "./YamlParamInput";

describe("ParameterInput supports readOnly", () => {
  it("StringParamInput editable", () => {
    render(<StringParamInput definition={stringDef} readOnly={false} />);
    const input = screen.getByRole("textbox");
    expect(input).not.toBeDisabled();
  });

  it("StringParamInput readonly", () => {
    render(<StringParamInput definition={stringDef} readOnly={true} />);
    const input = screen.getByRole("textbox");
    expect(input).toBeDisabled();
  });

  it("StringsParamInput editable", () => {
    render(<StringsParamInput definition={stringsDef} readOnly={false} />);
    const input = screen.getByRole("combobox");
    expect(input).not.toBeDisabled();
  });

  it("StringsParamInput readonly", () => {
    render(<StringsParamInput definition={stringsDef} readOnly={true} />);
    const input = screen.getByRole("combobox");
    expect(input).toBeDisabled();
  });

  it("EnumParamInput editable", () => {
    render(<EnumParamInput definition={enumDef} readOnly={false} />);
    const input = screen.getByRole("combobox");
    expect(input).not.toBeDisabled();
  });

  it("EnumParamInput readonly", () => {
    render(<EnumParamInput definition={enumDef} readOnly={true} />);
    const input = screen.getByRole("combobox");
    expect(input).toBeDisabled();
  });

  it("EnumsParamInput editable", () => {
    render(<EnumsParamInput definition={enumsDef} readOnly={false} />);
    const input = screen.getAllByRole("checkbox");
    for (const checkbox of input) {
      expect(checkbox).not.toBeDisabled();
    }
  });

  it("EnumsParamInput readonly", () => {
    render(<EnumsParamInput definition={enumsDef} readOnly={true} />);
    const input = screen.getAllByRole("checkbox");
    for (const checkbox of input) {
      expect(checkbox).toBeDisabled();
    }
  });

  it("BoolParamInput editable", () => {
    render(<BoolParamInput definition={enumDef} readOnly={false} />);
    const input = screen.getByRole("checkbox");
    expect(input).not.toBeDisabled();
  });

  it("BoolParamInput readonly", () => {
    render(<BoolParamInput definition={enumDef} readOnly={true} />);
    const input = screen.getByRole("checkbox");
    expect(input).toBeDisabled();
  });

  it("BoolParamInput Telemetry header editable", () => {
    render(
      <BoolParamInput definition={telemetrySectionBoolDef} readOnly={false} />
    );
    const input = screen.getByRole("checkbox");
    expect(input).not.toBeDisabled();
  });

  it("BoolParamInput Telemetry header readonly", () => {
    render(
      <BoolParamInput definition={telemetrySectionBoolDef} readOnly={true} />
    );
    const input = screen.getByRole("checkbox");
    expect(input).toBeDisabled();
  });

  it("IntParamInput editable", () => {
    render(<IntParamInput definition={intDef} readOnly={false} />);
    const input = screen.getByRole("textbox");
    expect(input).not.toBeDisabled();
  });

  it("IntParamInput readonly", () => {
    render(<IntParamInput definition={intDef} readOnly={true} />);
    const input = screen.getByRole("textbox");
    expect(input).toBeDisabled();
  });

  it("MapParamInput editable", () => {
    render(<MapParamInput definition={intDef} readOnly={false} />);
    const inputs = screen.getAllByRole("textbox");
    for (const textbox of inputs) {
      expect(textbox).not.toBeDisabled();
    }

    const addButton = screen.getByRole("button", { name: "New Row" });
    expect(addButton).not.toBeDisabled();
  });

  it("MapParamInput readonly", () => {
    render(<MapParamInput definition={intDef} readOnly={true} />);
    const inputs = screen.getAllByRole("textbox");
    for (const textbox of inputs) {
      expect(textbox).toBeDisabled();
    }

    const addButton = screen.getByRole("button", { name: "New Row" });
    expect(addButton).toBeDisabled();
  });

  it("TimezoneParamInput editable", () => {
    render(<TimezoneParamInput definition={timezoneDef} readOnly={false} />);
    const input = screen.getByRole("combobox");
    expect(input).not.toBeDisabled();
  });

  it("TimezoneParamInput readonly", () => {
    render(<TimezoneParamInput definition={timezoneDef} readOnly={true} />);
    const input = screen.getByRole("combobox");
    expect(input).toBeDisabled();
  });

  it("YamlParamInput editable", () => {
    render(<YamlParamInput definition={yamlDef} readOnly={false} />);
    const input = screen.getByRole("textbox");
    expect(input).not.toBeDisabled();
  });

  it("YamlParamInput readonly", () => {
    render(<YamlParamInput definition={yamlDef} readOnly={true} />);
    const input = screen.getByRole("textbox");
    expect(input).toBeDisabled();
  });

  it("AWSCloudwatchInput editable", () => {
    render(
      <AWSCloudwatchInput
        definition={awsCloudWatchFieldInputDef}
        readOnly={false}
      />
    );
    const inputs = screen.getAllByRole("textbox");
    for (const textbox of inputs) {
      expect(textbox).not.toBeDisabled();
    }
    const button = screen.getByRole("button", { name: "New field" });
    expect(button).not.toBeDisabled();
  });

  it("AWSCloudwatchInput readonly", () => {
    render(
      <AWSCloudwatchInput
        definition={awsCloudWatchFieldInputDef}
        readOnly={true}
      />
    );
    const inputs = screen.getAllByRole("textbox");
    for (const textbox of inputs) {
      expect(textbox).toBeDisabled();
    }
    const button = screen.getByRole("button", { name: "New field" });
    expect(button).toBeDisabled();
  });

  it("MetricsParamInput editable", () => {
    render(<MetricsParamInput definition={metricParamInput} />);

    expect(screen.getByText("Enable All")).not.toBeDisabled();
    expect(screen.getByText("Disable All")).not.toBeDisabled();
    expect(screen.getByRole("checkbox")).not.toBeDisabled();
  });

  it("MetricsParamInput readOnly", () => {
    render(<MetricsParamInput definition={metricParamInput} readOnly />);

    expect(screen.queryByText("Enable All")).not.toBeInTheDocument();
    expect(screen.queryByText("Disable All")).not.toBeInTheDocument();
    expect(screen.getByRole("checkbox")).toBeDisabled();
  });
});

describe("StringsParamInput trims whitespace", () => {
  var gotValue: string[] = [];

  const onValueChange = (value: string[]) => {
    gotValue = value;
  };

  render(
    <StringsParamInput definition={stringsDef} onValueChange={onValueChange} />
  );

  const autocomplete = screen.getByRole("combobox");

  fireEvent.change(autocomplete, { target: { value: "  test  " } });
  fireEvent.blur(autocomplete);

  expect(gotValue).toEqual(["test"]);

  fireEvent.change(autocomplete, { target: { value: "internal space" } });
  fireEvent.blur(autocomplete);

  expect(gotValue).toEqual(["internal space"]);
});
