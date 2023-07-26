import {
  GetDestinationWithTypeQuery,
  GetSourceWithTypeQuery,
} from "../../../graphql/generated";
import { MinimumRequiredConfig } from "../../PipelineGraph/PipelineGraph";

export const destination0Name = "destination-0";

export const destination0TypeName = "destination-type-name";

export const source1Name = "source-1";

export const source1TypeName = "source-type-name";

// testConfig contains an inline source, a resource source, and a resource destination
export const testConfig: MinimumRequiredConfig = {
  metadata: {
    name: "test-config",
    version: 1,
    id: "01H5MYJEPR0F382HX6XCYZKFRE",
  },
  spec: {
    sources: [
      // Source 0 is an inline source
      {
        type: "resource-type-1",
        parameters: [],
        processors: [],
        id: "01H5MYJMKEH9P3VGQTP18FQ2KQ",
        disabled: false,
      },
      // Source 1 is a resource source
      {
        name: source1Name,
        parameters: [],
        processors: [],
        id: "01H5MYKZ88PVH9WX093M9RHHJ3",
        disabled: false,
      },
    ],
    destinations: [
      {
        name: destination0Name,
        parameters: [],
        processors: [],
        id: "01H5MYKZ88PVH9WX093M9RHHJ3",
        disabled: false,
      },
    ],
  },
};

export const destination0: GetDestinationWithTypeQuery["destinationWithType"]["destination"] =
  {
    metadata: {
      name: destination0Name,
      version: 1,
      id: "01H5MZJ89M3Z3JBEK6CVA5H1KC",
      labels: {},
    },
    spec: {
      parameters: [],
      disabled: false,
      type: destination0TypeName,
    },
  };

export const destination0_PAUSED: GetDestinationWithTypeQuery["destinationWithType"]["destination"] =
  {
    metadata: {
      name: destination0Name,
      version: 1,
      id: "01H5MZJ89M3Z3JBEK6CVA5H1KC",
      labels: {},
    },
    spec: {
      parameters: [],
      disabled: true,
      type: destination0TypeName,
    },
  };

export const destination0Type: GetDestinationWithTypeQuery["destinationWithType"]["destinationType"] =
  {
    metadata: {
      name: destination0TypeName,
      version: 1,
      id: "01H5MZNCM92CYAQYJV703K7W67",
      icon: "",
      description: "",
    },
    spec: {
      parameters: [],
    },
  };

export const source1: GetSourceWithTypeQuery["sourceWithType"]["source"] = {
  metadata: {
    name: "source-0",
    version: 1,
    id: "01H5MYJEPR0F382HX6XCYZKFRE",
    labels: {},
  },
  spec: {
    parameters: [],
    disabled: false,
    type: source1TypeName,
  },
};

export const source1_PAUSED: GetSourceWithTypeQuery["sourceWithType"]["source"] =
  {
    metadata: {
      name: "source-0",
      version: 1,
      id: "01H5MYJEPR0F382HX6XCYZKFRE",
      labels: {},
    },
    spec: {
      parameters: [],
      disabled: true,
      type: source1TypeName,
    },
  };

export const source1Type: GetSourceWithTypeQuery["sourceWithType"]["sourceType"] =
  {
    metadata: {
      name: source1TypeName,
      version: 1,
      id: "01H5MZNCM92CYAQYJV703K7W67",
      icon: "",
      description: "",
    },
    spec: {
      parameters: [],
    },
  };
