import {
  asCurrentVersion,
  asLatestVersion,
  nameAndVersion,
  trimVersion,
} from "./version-helpers";

describe("version helpers", () => {
  it("trimVersion", () => {
    expect(trimVersion("config-name:1")).toEqual("config-name");
    expect(trimVersion("config-name:1:3")).toEqual("config-name");
    expect(trimVersion("config-name")).toEqual("config-name");
  });

  it("asCurrentVersion", () => {
    expect(asCurrentVersion("config-name")).toEqual("config-name:current");
    expect(asCurrentVersion("config-name:1")).toEqual("config-name:current");
  });

  it("asLatestVersion", () => {
    expect(asLatestVersion("config-name")).toEqual("config-name:latest");
    expect(asLatestVersion("config-name:1")).toEqual("config-name:latest");
  });

  it("nameAndVersion", () => {
    expect(nameAndVersion("config-name")).toEqual("config-name");
    expect(nameAndVersion("config-name", 5)).toEqual("config-name:5");
    expect(nameAndVersion("config-name:1", 5)).toEqual("config-name:5");
  });
});
