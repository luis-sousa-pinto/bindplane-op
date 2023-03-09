# Destination Type

For real world examples, check out BindPlane's included [destination types](../../../resources/destination-types/)

## Create Basic Destination Type

The following destination type defines a Google destination without configuration.
The destination type can be implemented into log, metric, and trace pipelines.

```yaml
# mydestinationtype.yaml
apiVersion: bindplane.observiq.com/v1
kind: DestinationType
metadata:
  name: custom_googlecloud
  displayName: Custom Google Cloud
  icon: /icons/destinations/google-cloud-logging.svg
  description: Send metrics, traces, and logs to Google Cloud.
spec:
  parameters:
  logs+metrics+traces:
    exporters: |
      - googlecloud:
```

You can deploy the destination type with:

```bash
bindplanectl apply -f ./mydestinationtype.yaml
```

> **_NOTE:_**  Refresh the browser in order to see the new destination type.

Lets say you want to expose the `project` paramter for the Google exporter. You can add a
`project` paramter to the destination type and make it `required`.

```yaml
# mydestinationtype.yaml
apiVersion: bindplane.observiq.com/v1
kind: DestinationType
metadata:
  name: custom_googlecloud
  displayName: Custom Google Cloud
  icon: /icons/destinations/google-cloud-logging.svg
  description: Send metrics, traces, and logs to Google Cloud.
spec:
  parameters:
    - name: project
      label: Project ID
      description: The Google Cloud Project ID to send logs, metrics, and traces to.
      type: string
      default: ""
      required: true
  logs+metrics+traces:
    exporters: |
      - googlecloud:
          project: "{{ .project }}"
```

You can update the destination type with:

```bash
bindplanectl apply -f ./mydestinationtype.yaml
```

> **_NOTE:_**  Refresh the browser in order to see changes.

## Advanced Destination Type Example

Destination types are flexible. They can support multiple pipeline types, and multiple exporters. See
the [logzio](https://github.com/observIQ/bindplane-op/blob/main/resources/destination-types/logzio.yaml) destination type for an advanced example.

The `logz.io` destination type has the following
- Three different exporters, one for each pipeline (prometheus remote write, logzio/tracing, logzio/logs)
- Batch processor
- Conditions for enabling or disabling metrics, logs, traces
