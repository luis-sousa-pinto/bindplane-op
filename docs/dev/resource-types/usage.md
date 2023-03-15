# Resource Type Usage

The `bindplanectl` command line interface can be used to interact with resource types.

Resource types are used to generate agent configurations. Changes to resource types
will immediately result in updates to related configurations.

## List Resource Types

You can list source, processor, and destination types with the `get` command.

```bash
bindplanectl get source-type
bindplanectl get processor-types
bindplanectl get destination-types
```

## Delete Resource Type

Resource types can be deleted with the `delete` command.

```
bindplanectl delete source-type <source type name>
```

> **_NOTE:_**  Build in resource types can be deleted, however, they will be recreated when BindPlane is restarted.

## Describe Resources

Individual resource types can be described by passing their name as
an argument along with the `yaml` output option. 

For example, you can describe the `host` source type like this:

```bash
bindplanectl get source-type host -o yaml
```

## Create New Resource Type From Built In Resource Type

If you wish to make a copy of the source type, output it to a file:

```bash
bindplanectl get source-type host -o yaml > host.custom.yaml
```

> **_NOTE:_**  Build in resource types can be modified, but will be replaced by BindPlane on restart. It is best to create a new resource type based on a built in resource type.

Make the following changes:
- Remove `metadata.id` field
- Change `metadata.name` to something other than `host`
- Modify `metadata.displayName` and `metadata.description`
- Make your desired changes to the remainder of the source type, such as changes to parameters.

Once you have reconfigured your source type, you can apply it to create a new source type:

```bash
bindplanectl apply -f host.custom.yaml
```

Comfirm the new source type exists:

```
bindplanectl get source-type <new source type name>
```

If using the web interface, be sure to refresh your browser before searching for the new source type.

## FAQ

- [Do modified resoures cause config updates?](./usage.md#do-modified-resoures-cause-config-updates)
- [Can built in resources be modified?](./usage.md#can-built-in-resources-be-modified)

### Do modified resoures cause config updates?

Yes. Anytime a resource type is changed, configurations which use the underlying resource type will be updated
automatically.

### Can built in resources be modified?

No. Changes to built in resource types will be overwritten when BindPlane is restarted. It
is recommended that users copy a built in resource type and use it to create a new
unique resource type.
