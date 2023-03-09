# Source Type

For real world examples, check out BindPlane's included [source types](../../../resources/source-types/)

## Create Basic Source Type

The following source type has minimal configuration, and is designed to run on Linux.

```yaml
# mysourcetype.yaml
apiVersion: bindplane.observiq.com/v1
kind: SourceType
metadata:
  name: otlp_grpc_metrics
  displayName: OpenTelemetry GRPC Metrics
  icon: /icons/destinations/otlp.svg
  description: Receive metrics from OTLP exporters via GRPC.
spec:
  version: 0.0.1
  supported_platforms:
    - linux
  parameters:
  metrics:
    receivers: |
      - otlp:
          protocols:
            grpc:
              endpoint: 0.0.0.0:4317
```

You can deploy the source type with:

```bash
bindplanectl apply -f ./mysourcetype.yaml
```

> **_NOTE:_**  Refresh the browser in order to see the new source type.

In it's current form, it is not all that useful as the user cannot configure
the `otlp` receiver. Perhaps you wish to allow the "listen address" and "port" to be configurable.

In this example, the source type contains `listen_address` and `port` parameters. These allow the
user to configure an address and port if the default values are not suitable. Paramter values are
injected into the metrics pipeline using the `{{ }}` templating syntax.

```yaml
# mysourcetype.yaml
apiVersion: bindplane.observiq.com/v1
kind: SourceType
metadata:
  name: otlp_grpc_metrics
  displayName: OpenTelemetry GRPC Metrics
  icon: /icons/destinations/otlp.svg
  description: Receive metrics from OTLP exporters via GRPC.
spec:
  version: 0.0.1
  supported_platforms:
    - linux
  parameters:
    - name: listen_address
      label: Listen Address
      description: The IP address to listen on.
      type: string
      default: "0.0.0.0"

    - name: port
      label: GRPC Port
      description: TCP port to receive OTLP telemetry using the gRPC protocol.
      type: int
      default: 4317

  metrics:
    receivers: |
      - otlp:
          protocols:
            grpc:
              endpoint: {{ .listen_address }}:{{ .port }}
```

You can update the source type with:

```bash
bindplanectl apply -f ./mysourcetype.yaml
```

> **_NOTE:_**  Refresh the browser in order to see changes.

## Advanced Source Type Example

See the [nginx](https://github.com/observIQ/bindplane-op/blob/main/resources/source-types/nginx.yaml) source
type for an advanced example. This source type utilizes several advanced options.

- Metrics and Logs pipelines are enabled or disabled using parameters
- Log path array is dynamically set by iterating over a `strings` parameter value
- `relevantIf` used to enable or disable TLS parameters
