import { initQuery } from "./utils";

describe("initQuery", () => {
  it("configuration=blah, platform=macos", () => {
    const expectQuery = "-configuration:blah platform:darwin";
    const gotQuery = initQuery({ configuration: "blah" }, "macos");

    expect(gotQuery).toEqual(expectQuery);
  });

  it("configuration=blah, platform=linux", () => {
    const expectQuery = "-configuration:blah platform:linux";
    const gotQuery = initQuery({ configuration: "blah" }, "linux");

    expect(gotQuery).toEqual(expectQuery);
  });

  it("configuration=blah, platform=windows", () => {
    const expectQuery = "-configuration:blah platform:windows";
    const gotQuery = initQuery({ configuration: "blah" }, "windows");

    expect(gotQuery).toEqual(expectQuery);
  });

  it("configuration=blah, platform=undefined", () => {
    const expectQuery = "-configuration:blah";
    const gotQuery = initQuery({ configuration: "blah" }, undefined);

    expect(gotQuery).toEqual(expectQuery);
  });
});
