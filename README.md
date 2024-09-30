# protoc-gen-mikros-extensions

A mikros framework protoc/buf plugin to extend features for services and applications.

## Features

The plugin adds the following features for generated source code:

* New entities for domain, inbound and outbound messages.
* Validation for wire input messages.
* HTTP server routes for HTTP services.
* Unit test helpers API for creating entities.
* The possibility of extending the plugin using [addons](docs/addons.md).

## Building and installing

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

import "protoc-gen-mikros-extensions/mikros/extensions/extensions.proto";

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
