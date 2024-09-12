# protoc-gen-mikros-extensions

A mikros framework protoc/buf plugin to extend features for services and applications.

## Building and installing

In order to compile and install the plugin locally you'll need to follow the steps:

* Install the go compiler;
* Execute the commands:
  * `go generate`
  * `go build && go install`

## Extensions available

* [Enum](docs/enum.md)
* [RPC Methods](docs/method.md)
* [Fields](docs/field.md)
* [Messages](docs/message.md)

## Usage example

Using the extensions inside a .proto file:

```protobuf
syntax = "proto3";

package services.person;

import "protoc-gen-mikros-extensions/mikros/extensions/extensions.proto";

option go_package = "github.com/example/services/gen/go/services/person;person;"

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