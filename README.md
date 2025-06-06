# protoc-gen-mikros-extensions

A mikros framework protoc/buf plugin to extend features for services and applications.

## Features

The plugin adds the following features for generated source code:

* New entities for domain, inbound and outbound messages.
* Validation for wire input messages.
* HTTP server routes for HTTP services.
* Unit test helpers API for creating entities.
* The possibility of extending the plugin using [addons](docs/addons.md).

## Installation and usage into projects

To install the plugin latest version and use it in your projects, use the command:
```bash
go install github.com/mikros-dev/protoc-gen-mikros-extensions@latest
```

Assuming that a project is using [buf](https://buf.build/docs/) tool to compile
and manage protobuf files, this plugin can be used the following way:

> Note: We assume buf version 2 here, if you're using version 1, use buf docs
> to check how to set a local plugin (or to migrate your settings to version 2).

* Edit your **buf.gen.yaml** file, in the `plugins` section and add the following
excerpt:
```yaml
plugins:
  - local: protoc-gen-mikros-extensions
    out: gen # Where your generated files will be
    opt:
      - settings=extensions_settings.toml # The file name of your plugin settings
```

* Edit the **buf.yaml** file, in the `deps` section, add the following excerpt:
```yaml
deps:
  - buf.build/mikros-dev/protoc-gen-mikros-extensions
```

* Execute the command:
```bash
buf dep update
```

## Building and installing locally

In order to compile and install the plugin locally you'll need to follow the steps:

* Install the go compiler;
* Execute the commands:
  * `go generate`
  * `go build && go install`

## Protobuf extensions available

* [Service](docs/service.md)
* [Enum](docs/enum.md)
* [RPC Methods](docs/method.md)
* [Fields](docs/field.md)
* [Messages](docs/message.md)

## Usage

The plugin should be used according your environment, if you're using `buf` or
`protoc`. It does not need any mandatory option to execute over your proto files.

But, if you want to use custom settings, the option named **settings** must point
to a valid TOML file in your system. Its syntax details are described [here](docs/settings.md).

## Example

Using the extensions inside a .proto file:

```protobuf
syntax = "proto3";

package services.person;

import "mikros_extensions.proto";

option go_package = "github.com/example/services/gen/go/services/person;person;";

service PersonService {
  rpc CreatePerson(CreatePersonRequest) returns (CreatePersonResponse);
}

message CreatePersonRequest {
  string name = 1;
  int32 age = 2;
}

message CreatePersonResponse {
  PersonWire person = 1;
}

message PersonWire {
  string id = 1;
  string name = 2;
  int32 age = 3;
}
```

For a more complete example, check the [examples](examples/README.md) directory.

## License

[Apache License 2.0](LICENSE)
