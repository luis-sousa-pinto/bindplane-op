import { ComponentStory } from "@storybook/react";
import { SnackbarProvider } from "notistack";
import { ConfigurationEditor } from "./ConfigurationEditor";
import {
  ALL_CONFIG_MOCKS,
  CUSTOM_DESTINATION_TYPE_MOCK,
  SOURCE_TYPE_HOST_MOCK,
  VERSION_MOCK_NO_HISTORY,
  VERSION_MOCK_NO_HISTORY_WITH_PENDING,
  VERSION_MOCK_WITH_HISTORY,
  VERSION_MOCK_WITH_HISTORY_AND_NEW,
  VERSION_MOCK_WITH_NEW_AND_PENDING,
} from "./__test__/mocks";

export default {
  title: "Configuration Editor",
  component: ConfigurationEditor,
};

const Template: ComponentStory<typeof ConfigurationEditor> = (args) => {
  return (
    <SnackbarProvider>
      <div style={{ width: 1200 }}>
        <ConfigurationEditor configurationName="linux-metrics" isOtel={false} />
      </div>
    </SnackbarProvider>
  );
};

export const Loading = Template.bind({});
Loading.parameters = {
  apolloClient: {
    mocks: [VERSION_MOCK_NO_HISTORY],
  },
};

export const NoHistory = Template.bind({});
NoHistory.parameters = {
  apolloClient: {
    mocks: [VERSION_MOCK_NO_HISTORY, ...ALL_CONFIG_MOCKS],
  },
};

export const NoHistoryWithPending = Template.bind({});
NoHistoryWithPending.parameters = {
  apolloClient: {
    mocks: [VERSION_MOCK_NO_HISTORY_WITH_PENDING, ...ALL_CONFIG_MOCKS],
  },
};

/**
 * HistoryAvailable shows a configuration with 3 versions,
 * where the current version is the latest (Version 3).
 */
export const HistoryAvailable = Template.bind({});
HistoryAvailable.parameters = {
  apolloClient: {
    mocks: [
      VERSION_MOCK_WITH_HISTORY,
      SOURCE_TYPE_HOST_MOCK,
      CUSTOM_DESTINATION_TYPE_MOCK,
      ...ALL_CONFIG_MOCKS,
    ],
  },
};

/**
 * HistoryAndNewAvailable shows a configuration with 3 versions,
 * where the current version is not the latest (Version 2).
 */
export const HistoryAndNewAvailable = Template.bind({});
HistoryAndNewAvailable.parameters = {
  apolloClient: {
    mocks: [
      VERSION_MOCK_WITH_HISTORY_AND_NEW,
      SOURCE_TYPE_HOST_MOCK,
      CUSTOM_DESTINATION_TYPE_MOCK,
      ...ALL_CONFIG_MOCKS,
    ],
  },
};

export const HistoryPendingAndNewAvailable = Template.bind({});
HistoryPendingAndNewAvailable.parameters = {
  apolloClient: {
    mocks: [
      VERSION_MOCK_WITH_NEW_AND_PENDING,
      SOURCE_TYPE_HOST_MOCK,
      CUSTOM_DESTINATION_TYPE_MOCK,
      ...ALL_CONFIG_MOCKS,
    ],
  },
};
