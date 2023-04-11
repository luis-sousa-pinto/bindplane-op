import { render } from "@testing-library/react";
import { Metric } from "../../../graphql/generated";
import { MetricsRecordRow } from "./MetricsRecordRow";

describe("MetricsRecordRow", () => {
  it("renders a Summary type with Object value", () => {
    const message: Metric = {
      name: "go_gc_duration_seconds",
      timestamp: "2023-04-10T18:35:21.486Z",
      value: {
        "0": 0.000037883,
        "1": 0.003927915,
        "0.25": 0.000043135,
        "0.5": 0.000050907,
        "0.75": 0.000065501,
      },
      unit: "",
      type: "Summary",
      attributes: {},
      resource: {
        cluster: "cluster-name",
        "http.scheme": "http",
        location: "us-east1",
        namespace: "namespace-name",
        "net.host.port": "9100",
        "service.instance.id": "service-instance-id",
        "service.name": "service-name",
        "service.namespace": "service-namespace",
      },
      __typename: "Metric",
    };

    render(<MetricsRecordRow message={message} />);
  });
});
