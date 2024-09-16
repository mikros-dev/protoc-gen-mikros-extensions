# Examples

This directory holds protobuf and addons examples demonstrating how to use
the extensions annotations with protobuf declaration and how to build custom
addons for the plugin.

## Addons

Addons are a mechanism that allows a user extend even more the plugin
features and add custom templates and outputs according to its needs. In fact,
you can call them as plugins inside plugins.

Following some rules (i.e. an interface) one can add new features to its
environment without the need of building a whole new protoc/buf plugin.

For more details in how to write new addons, follow the documentation [here](../docs/addons.md).

The [addons](addons) directory has the following examples:

* [wire_improve](addons/wire_improve): demonstrates how to add a new template with new API for proto messages.
* [domain_improve](addons/domain_improve): demonstrates how to add new APIs for domain messages using custom addon proto annotations.

> **A little note about custom addon proto annotations.**
> 
> The addon proto file must be compiled, to generate the go source code, before
> compiling the plugin final version. The plugin must "known" all annotation
> extensions to be able to generate the right code when executed over the final
> proto files.

### Building the addons

In order to build the addons examples, execute the script `build-addons.sh`
located in the addons directory.

## Protobuf annotations

Protobuf annotations examples can be found inside the [protobuf](protobuf)
directory. There you'll find some examples showing how to use the plugin
annotations into services, messages and field declarations.

### How to generate the examples

In order to generate the examples, you must have installed [buf](https://buf.build). 
After installed, execute the following command inside the protobuf directory:

```bash
buf generate proto
```