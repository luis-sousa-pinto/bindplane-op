import { debounce } from "lodash";
import { useConfigurationChangesSubscription } from "../graphql/generated";
import { trimVersion } from "../utils/version-helpers";

/**
 * useRefetchOnConfigurationChange will call the provided refetch when
 * changes are received via the useConfigurationChanges query.  The
 * refetch is called right away on first change and then debounced
 * with a wait of 1 second.
 *
 * @param configurationName the name of the configuration.
 * @param refetch the refetch function to call when changes are received.
 */
export function useRefetchOnConfigurationChange(
  configurationName: string,
  refetch: () => void
) {
  const debouncedRefetch = debounce(refetch, 1000, { leading: true });
  useConfigurationChangesSubscription({
    variables: {
      query: `name:${configurationName}`,
    },
    onData: ({ data }) => {
      const found = data?.data?.configurationChanges.some(
        (c) => c.configuration.metadata.name === trimVersion(configurationName)
      );
      if (found) {
        debouncedRefetch();
      }
    },
  });
}
