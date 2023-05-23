import { GetConfigurationVersionsQuery } from "../../graphql/generated";

interface VersionMap {
  currentVersion?: number;
  pendingVersion?: number;
  newVersion?: number;
}

export class VersionsData
  implements Omit<GetConfigurationVersionsQuery, "__typename">
{
  configurationHistory: GetConfigurationVersionsQuery["configurationHistory"];
  constructor(data: GetConfigurationVersionsQuery) {
    this.configurationHistory = data.configurationHistory;
  }

  /**
   * versionHistory returns all of the versions that are not latest or current.
   */
  versionHistory() {
    return this.configurationHistory.filter(
      (version) =>
        !version.status.latest &&
        !version.status.current &&
        !version.status.pending
    );
  }

  /**
   * findCurrentVersion returns the current version or undefined if there is no current version
   */
  findCurrent() {
    return this.configurationHistory.find((version) => version.status.current);
  }

  /**
   * findPendingVersion returns the pending version or undefined if there is no pending version
   */
  findPending() {
    return this.configurationHistory.find(
      (version) => version.status.pending && !version.status.current
    );
  }

  /**
   * findNewVersion returns the latest version if it is not pending or stable
   */
  findNew() {
    return this.configurationHistory.find(
      (version) =>
        version.status.latest &&
        !version.status.pending &&
        !version.status.current
    );
  }

  /**
   * latestHistory returns the highest version number in the history
   * where status.latest and status.current are both false
   */
  latestHistoryVersion(): number {
    const history = this.versionHistory();
    if (history.length === 0) {
      return 1;
    }
    // find the highest version number in history
    return history.reduce((prev, current) =>
      prev.metadata.version > current.metadata.version ? prev : current
    ).metadata.version;
  }

  /**
   * versionMap returns a map of the latest, pending, and new versions
   */
  versionMap(): VersionMap {
    return {
      currentVersion: this.findCurrent()?.metadata.version,
      pendingVersion: this.findPending()?.metadata.version,
      newVersion: this.findNew()?.metadata.version,
    };
  }
}
