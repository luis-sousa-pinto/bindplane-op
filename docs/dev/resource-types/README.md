# Resource Types

BindPlane has resource types for building sources, processors, and destinations.

Source, processor, and destination types embed a 
[ResourceType](https://pkg.go.dev/github.com/observiq/bindplane-op@v1.5.0/model#SourceType), 
which means they each share the same fields for configuration.
- [SourceType](https://pkg.go.dev/github.com/observiq/bindplane-op@v1.5.0/model#SourceType)
- [DestinationType](https://pkg.go.dev/github.com/observiq/bindplane-op@v1.5.0/model#DestinationType)
- [ProcessorType](https://pkg.go.dev/github.com/observiq/bindplane-op@v1.5.0/model#ProcessorType)

BindPlane can generate collector configurations from resource types based on parameters set by the user. Resource types allow BindPlane to "package" usecase specific configuration and allows for default values and conditional
configuration.

## Example

Reference the following pages for resource type examples
- [Source types](./source-type.md)
- [Processor types](./processor-type.md)
- [Destination types](./destination-type.md)
- [Resource usage doc](./usage.md)
- [Built in source types](../../../resources/source-types/)
- [Built in processor types](../../../resources/processor-types/)
- [Built in destination types](../../../resources/destination-types/)

## Configuration

Resources have the following fields

### ResourceType 

| Field        | Description                    |
| ------------ | ------------------------------ |
| `apiVersion` | The resource type API version. |
| `kind`       | The underlying resource type (`SourceType`, `DestinationType`, `ProcessorType`). |
| `metadata`   | Metadata fields which describe the resource. |
| `metadata.id`          | Optional resource id. If not set, BindPlane will generate one.|
| `metadata.name`        | The name of the resource, must not contain spaces. |
| `metadata.displayName` | Friendly name, which can contain spaces. Displayed in the BindPlane web UI. |
| `metadata.description` | Description of the resource. Should indicate the usecase and the underlying collector component(s) that it implements. |
| `metadata.icon`        | Optional path to the icon. See the [example section](./README.md#example). |
| `metadata.labels`      | Optional labels. Currently, labels are ignored but can be used for specifiying friendly metadata, such as `datacenter=us-east1`. |
| `spec`                 | Spec fields which define the resource configuration options. |
| `spec.version`         | The resource type's version. Currently, resource versions are ignored. |
| `spec.supported_platforms`       | An array of supported platforms. Valid values include `linux`, `windows`, `macos` |
| `spec.parameters`                | Parameters define a resource paramter, such as "collection interval". See [parameter types](./README.md#parameter-types). |
| `spec.parameters.name`           | The name of the parameter.    |
| `spec.parameters.label`          | A friendly name for the parameter. Shown in the UI. |
| `spec.parameters.description`    | A summary of the parameter's purpose and usage. |
| `spec.parameters.required`       | A boolean value indicating whether the parameter is required or optional. |
| `spec.parameters.type`           | The parameters type. Valid options include `string`, `strings`, `int`, `bool`, or `enum`. |
| `spec.parameters.validValues`    | Required when `type` is set to `enum`. Contains a list of valid input values. |
| `spec.parameters.default`        | The default value if left unset. |
| `spec.parameters.relevantIf`     | An array of [relavant if conditions](https://pkg.go.dev/github.com/observiq/bindplane-op@v1.5.0/model#RelevantIfCondition). The parameter will be ignored if the conditions are not satisfied. |
| `spec.parameters.hidden`         | Boolean value, hides the parameter from the UI. Useful for deprecated parameters with default values. |
| `spec.parameters.advancedConfig` | Boolean value, determines if the parameter should be nested under the "advanced config" section in the UI. |
| `spec.parameters.options`        | Options for formating the parameter in the UI. See [the options type](https://pkg.go.dev/github.com/observiq/bindplane-op@v1.5.0/model#ParameterOptions). |
| `spec.parameters.documentation`  | An array of [documentation links](https://pkg.go.dev/github.com/observiq/bindplane-op@v1.5.0/model#DocumentationLink). |
| `spec.logs`                      | Defines the `logs` OpenTelemetry pipeline.    |
| `spec.metrics`                   | Defines the `metrics` OpenTelemetry pipeline. |
| `spec.traces`                    | Defines the `traces` OpenTelemetry pipeline.  |
| `spec.logs+metrics`              | Defines a special case where generic configuration should be applied to `logs` and `metrics` pipelines.   |
| `spec.logs+traces`               | Defines a special case where generic configuration should be applied to `logs` and `metrics` pipelines.   |
| `spec.metrics+traces`            | Defines a special case where generic configuration should be applied to `metrics` and `traces` pipelines. |
| `spec.logs+metrics+traces`       | Defines a special case where generic configuration should be applied to `logs`, `metrics`, and `traces` pipelines. Common with destination types which use a single exporter for all three pipelines. |


### Parameter Types

Parameters can have the following types:

| Type        | Description          |
| ----------- | -------------------- |
| `string`    | A string value.      |
| `strings`   | An array of strings. |
| `int`       | An integer value.    |
| `bool`      | A boolean value.     |
| `enum`      | A list of string values, selectable by the user. |

## References

- [Built in source type documentation](https://github.com/observIQ/bindplane-op/tree/main/docs/www/integrations/sources)
- [Built in processor type documentation](https://github.com/observIQ/bindplane-op/tree/main/docs/www/integrations/processors)
- [Built in destination type docuemtnation](https://github.com/observIQ/bindplane-op/tree/main/docs/www/integrations/destinations)
