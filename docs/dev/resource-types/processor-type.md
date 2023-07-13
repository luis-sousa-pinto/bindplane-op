# Processor Type

For real world examples, check out BindPlane's included [processor types](../../../resources/processor-types/)


## Create Basic Processor Type

The following processor type adds metric resources to all metrics.

```yaml
# myprocessortype.yaml
apiVersion: bindplane.observiq.com/v1
kind: ProcessorType
metadata:
  name: custom_add_resource
  displayName: Custom Add Resource Attribute
  description: Upsert metric resources.
spec:
  version: 0.0.1
  parameters:
    - name: resources
      label: Resources
      type: map
      required: true

  metrics:
    processors: |
      - resource:
          attributes:
            {{ range $k, $v := .resources }}
            - key: '{{ $k }}'
              value: {{ $v }}
              action: upsert
            {{ end }}
```

You can deploy the source type with:

```bash
bindplane apply -f ./myprocessortype.yaml
```

> **_NOTE:_**  Refresh the browser in order to see the new source type.

## Advanced Source Type Example

See the [add](https://github.com/observIQ/bindplane-op/blob/main/resources/processor-types/add_resource.yaml) processor type for an advanced example.
