# Addons

This documentation shows how to build and use `addons` inside the plugin to
extend its features.

## Introduction

There may be cases where a new custom code needs to be generated from your
protobuf files, or a new type of message needs to be created to be used by
your services. These scenarios can be resolved adding new features for the
plugin without changing its source code.

We call these new features as `addons` here.

They provide the developer an easy way to add custom templates (for generated
code) and annotations (protobuf annotations).

## Creating a new addon

To create a new plugin addon you simply need to create a new go module somewhere
that will need to be compiled as a golang [plugin](https://pkg.go.dev/plugin).

This addon must export a structure named `Addon` that implements the
[Addon](../pkg/addon/addon.go) interface.

For some examples of how to create an addon, you can check the addons created
in the [examples](../examples) directory. There you'll find different source
code examples and scripts showing how to build them.

### Beware of new protobuf annotations

Custom protobuf annotations are really helpful when one wants to add a more
detailed functionality, both at the .proto file and the template syntax.

But, in order to improve the plugin annotations and to use these new custom
ones inside the protobuf files, the new annotations must be declared in its
own .proto file, following the official [extension declarations](https://protobuf.dev/programming-guides/extension_declarations/).

Also, this new .proto file must be put inside the `mikros/extensions`
directory so that, when building the plugin, all .proto files, both the plugin
and the custom ones, are compiled together. This is required for these new
annotations to be found inside the addon code when used.

## Context

The context is the place where information that need to be accessed inside
templates are put. The plugin already shares a context with its templates that
can also be used inside addon templates.

This structure provides access to everything declared inside the protobuf files:
messages, RPCs, service, etc. If you want to know more about what is available,
you can take a look directly in the [context](../pkg/context) package itself.
