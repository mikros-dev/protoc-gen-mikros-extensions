# Settings

Some plugin behaviour can be customized by using different options adjusted
using a TOML file. These custom settings can be enabled using the `settings`
plugin option when executing it.

The repository has a default [settings](../configs/protoc-gen-mikros-extensions.toml)
file that can be used to create a new one according your needs.

## Options available

The settings file offers a bunch of options for you to customize the plugin
execution according your environment.

You can set:

* if debug messages are going to be displayed when running or not;
* where your addons reside;
* the database driver chosen for domain messages;
* the HTTP driver chosen for HTTP services API.

In addition to these options, the following ones are also available.

### Messages kind

The `suffix` section lets you define the suffixes you use in your messages
so that the plugin can recognize them and assume the right type of message
when dealing with them.

### Templates

The `templates` section provides settings to customize how the templates
will be handled by the plugin.

#### Converters

By default, the plugin generates some converters APIs (functions to convert
to pointers, protobuf timestamp to stdlib time.Time structures, among others)
that can be replaced by your own API.

In order to use a custom converters API you need to set the `templates.common.converters`
to `true` and use the `templates.common.api.converters` to set your API details,
so the plugin can reference it.

### Validations

Another feature that can be expanded is the validation for fields generated
by the plugin.

The `FIELD_VALIDATOR_RULE_CUSTOM` validator rule type lets you reference a
custom validator declared inside the settings file. Where you can, declare
its package name and the custom validation rules names. 

The plugin uses [ozzo](https://github.com/go-ozzo/ozzo-validation) validation
for its rules and the new custom validation rules must implement its validate
interface:

```golang
// Validate validates a value and returns an error if validation fails.
Validate(value interface{}) error
```
