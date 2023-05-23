import { VersionsData } from "./util";
import {
  HISTORY_LATEST_IS_CURRENT,
  HISTORY_LATEST_IS_NEW,
  HISTORY_LATEST_IS_NEW_WITH_PENDING,
  HISTORY_LATEST_IS_PENDING,
  NO_VERSION_HISTORY,
  NO_VERSION_HISTORY_CURRENT_IS_STABLE,
  NO_VERSION_HISTORY_WITH_PENDING,
} from "./__test__";

describe("VersionsData", () => {
  it("versionHistory", () => {
    var versionData = new VersionsData(NO_VERSION_HISTORY);
    expect(versionData.versionHistory()).toEqual([]);

    versionData = new VersionsData(HISTORY_LATEST_IS_CURRENT);

    expect(versionData.versionHistory()).toEqual(
      HISTORY_LATEST_IS_CURRENT.configurationHistory.filter(
        (h) => h.metadata.version !== 3
      )
    );

    versionData = new VersionsData(HISTORY_LATEST_IS_NEW);
    expect(versionData.versionHistory()).toEqual(
      HISTORY_LATEST_IS_NEW.configurationHistory.filter(
        (h) => h.metadata.version !== 3 && h.metadata.version !== 2
      )
    );

    versionData = new VersionsData(HISTORY_LATEST_IS_PENDING);
    expect(versionData.versionHistory()).toEqual(
      HISTORY_LATEST_IS_PENDING.configurationHistory.filter(
        (h) => h.metadata.version !== 3 && h.metadata.version !== 2
      )
    );
  });

  it("findNewVersion", () => {
    var versionData = new VersionsData(NO_VERSION_HISTORY);
    expect(versionData.findNew()?.metadata.version).toEqual(1);

    versionData = new VersionsData(NO_VERSION_HISTORY_WITH_PENDING);
    expect(versionData.findNew()?.metadata.version).toEqual(undefined);

    versionData = new VersionsData(NO_VERSION_HISTORY_CURRENT_IS_STABLE);
    expect(versionData.findNew()?.metadata.version).toEqual(undefined);

    versionData = new VersionsData(HISTORY_LATEST_IS_PENDING);
    expect(versionData.findNew()?.metadata.version).toEqual(undefined);

    versionData = new VersionsData(HISTORY_LATEST_IS_NEW_WITH_PENDING);
    expect(versionData.findNew()?.metadata.version).toEqual(3);
  });

  it("findCurrentVersion", () => {
    var versionData = new VersionsData(NO_VERSION_HISTORY);
    expect(versionData.findCurrent()).toEqual(undefined);

    versionData = new VersionsData(NO_VERSION_HISTORY_WITH_PENDING);
    expect(versionData.findCurrent()).toEqual(undefined);

    versionData = new VersionsData(HISTORY_LATEST_IS_PENDING);
    expect(versionData.findCurrent()?.metadata.version).toEqual(2);

    versionData = new VersionsData(HISTORY_LATEST_IS_CURRENT);
    expect(versionData.findCurrent()?.metadata.version).toEqual(3);
  });

  it("findPendingVersion", () => {
    var versionData = new VersionsData(NO_VERSION_HISTORY);
    expect(versionData.findPending()).toEqual(undefined);

    versionData = new VersionsData(NO_VERSION_HISTORY_WITH_PENDING);
    expect(versionData.findPending()?.metadata.version).toEqual(1);

    versionData = new VersionsData(HISTORY_LATEST_IS_PENDING);
    expect(versionData.findPending()?.metadata.version).toEqual(3);

    versionData = new VersionsData(HISTORY_LATEST_IS_CURRENT);
    expect(versionData.findPending()).toEqual(undefined);
  });

  it("latestHistoryVersion", () => {
    var versionData = new VersionsData(NO_VERSION_HISTORY);
    expect(versionData.latestHistoryVersion()).toEqual(1);

    versionData = new VersionsData(HISTORY_LATEST_IS_CURRENT);
    expect(versionData.latestHistoryVersion()).toEqual(2);

    versionData = new VersionsData(HISTORY_LATEST_IS_NEW);
    expect(versionData.latestHistoryVersion()).toEqual(1);

    versionData = new VersionsData(HISTORY_LATEST_IS_PENDING);
    expect(versionData.latestHistoryVersion()).toEqual(1);
  });
});
