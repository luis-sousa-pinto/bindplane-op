import { GetConfigRolloutStatusQuery } from "../../graphql/generated";

type NonNullConfig = NonNullable<GetConfigRolloutStatusQuery["configuration"]>;

export class RolloutProgressData implements GetConfigRolloutStatusQuery {
  configuration: NonNullConfig;
  constructor(configuration: NonNullConfig) {
    this.configuration = configuration;
  }

  /**
   *  total returns the sum of completed, pending, and waiting
   */
  total() {
    const { completed, pending, waiting } = this.configuration.status.rollout;
    return completed + pending + waiting;
  }

  /**
   * completed returns the completed field from the configuration status
   */
  completed() {
    return this.configuration.status.rollout.completed;
  }

  /**
   * errored returns the errors field from the configuration status
   */
  errored() {
    return this.configuration.status.rollout.errors;
  }

  /**
   * agentCount returns the agentCount field
   */
  agentCount() {
    return this.configuration.agentCount ?? 0;
  }

  /**
   * rolloutIsPaused returns true if the rollout status is paused
   */
  rolloutIsPaused() {
    return this.configuration.status.rollout.status === 2;
  }

  /**
   * rolloutIsStarted returns true if the rollout status is started
   */
  rolloutIsStarted() {
    return this.configuration.status.rollout.status === 1;
  }

  /**
   * rolloutIsPending returns true if the rollout status is pending
   */
  rolloutIsPending() {
    return this.configuration.status.rollout.status === 0;
  }

  /**
   * rolloutStatus returns the status field from the rollout status
   */
  rolloutStatus() {
    return this.configuration.status.rollout.status;
  }
}
