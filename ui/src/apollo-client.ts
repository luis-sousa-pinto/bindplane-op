import {
  ApolloClient,
  HttpLink,
  InMemoryCache,
  split,
  from,
  FieldMergeFunction,
  InMemoryCacheConfig,
} from "@apollo/client";
import { GraphQLWsLink } from "@apollo/client/link/subscriptions";
import { getMainDefinition } from "@apollo/client/utilities";
import { onError } from "@apollo/client/link/error";
import { isFunction } from "lodash";
import { createClient } from "graphql-ws";

const httpLink = new HttpLink({
  uri: "/v1/graphql",
  credentials: "same-origin",
});

const ws = window.location.protocol === "https:" ? "wss:" : "ws:";
const url = new URL(`${ws}//${window.location.host}/v1/graphql`);

const wsLink = new GraphQLWsLink(
  createClient({
    url: url.href,
  })
);

// Use the httpLink for queries and wsLink for subscriptions
export const requestLink = split(
  ({ query }) => {
    const definition = getMainDefinition(query);
    return (
      definition.kind === "OperationDefinition" &&
      definition.operation === "subscription"
    );
  },
  wsLink,
  httpLink
);

// authErrorLink will log a user out if a graphql query or
// subscription returns with a 401 unauthorized.
export const authErrorLink = onError(({ operation }) => {
  const context = operation.getContext();

  if (context.response.status === 401) {
    // Unset the user in local storage and navigate to login on 401s
    localStorage.removeItem("user");
    if (isFunction(window.navigate)) {
      window.navigate("/login");
    }
  }
});

// Chain the auth link and request link together
const link = from([authErrorLink, requestLink]);

/**
 * merge is used to merge fields that do not have a unique identifier.
 * This is discussed in apollo documentation here:
 * https://www.apollographql.com/docs/react/caching/cache-field-behavior
 */
const merge: FieldMergeFunction = (existing, incoming, { mergeObjects }) => {
  return mergeObjects(existing, incoming);
};

export const typePolicies: InMemoryCacheConfig["typePolicies"] = {
  Agent: {
    keyFields: ["id"],
  },
  Configuration: {
    keyFields: ["metadata"],
  },
  SourceType: {
    keyFields: ["metadata"],
    fields: {
      spec: {
        merge,
      },
    },
  },
  DestinationType: {
    keyFields: ["metadata"],
    fields: {
      spec: {
        merge,
      },
    },
  },
  Destination: {
    keyFields: ["metadata"],
    fields: {
      spec: {
        merge,
      },
    },
  },
  Metadata: {
    keyFields: ["id", "name", "version"],
  },
  ResourceConfiguration: {
    keyFields: ["id", "name", "type"],
  },
};

const APOLLO_CLIENT = new ApolloClient({
  link: link,
  cache: new InMemoryCache({
    typePolicies,
  }),
});

export default APOLLO_CLIENT;
