# Message options

A message can have the following options available:

| Name                          | Modifier | Description                                              |
|-------------------------------|----------|----------------------------------------------------------|
| [domain](#domain-expansion)   | optional | Options that modify the domain version of the message.   |
| [custom_api](#custom-api)     | optional | Options that adds user custom APIs to messages.          |
| [inbound](#inbound-options)   | optional | Options that modify the inbound version of the message.  |
| [outbound](#outbound-options) | optional | Options that modify the outbound version of the message. |

## Domain expansion

Available options:

| Name                        | Type | Modifier | Description                                                   |
|-----------------------------|------|----------|---------------------------------------------------------------|
| dont_export                 | bool | optional | Sets that the message won't have a domain equivalent message. |
| [naming_mode](#Naming-Mode) | enum | optional | Sets the naming output format.                                |

## Custom Api

Available options:

| Name          | Type    | Modifier | Description                                       |
|---------------|---------|----------|---------------------------------------------------|
| [code](#code) | message | array    | Adds new API for the wire version of the message. |

### Code

| Name              | Type               | Modifier | Description                                                       |
|-------------------|--------------------|----------|-------------------------------------------------------------------|
| signature         | string             | required | Defines the new API signature (name, arguments and return value). |
| body              | string (multiline) | required | Defines the new API function body.                                |
| [import](#import) | message            | array    | Sets the required custom imports for this new API.                |

#### Import

| Name  | Type   | Modifier | Description                       |
|-------|--------|----------|-----------------------------------|
| alias | string | optional | Sets the import alias name.       |
| name  | string | required | The name of the imported package. |

#### Example

```protobuf
message Person {
  option (mikros.extensions.wire_expansion) = {
    custom_code: {
      signature: "(p *Person) IsAlive() bool"
      body: "return p.alive"
    }
    
    custom_code: {
      signature: "(p *Person) SayHello()"
      body: "fmt.Println(\"Hello\")"
      import: {
        name: "fmt"
      }
    }
  };
  
  bool alive = 1;
}
```

```go
// Person struct defined by the golang gRPC plugin
type Person struct {
    Alive bool
}

// wire.go
package person

import (
    "fmt"
)

func (p *Person) IsAlive() bool {
    return p.alive
}

func (p *Person) SayHello() {
    fmt.Println("Hello")
}
```

## Inbound Options

Available options:

| Name                        | Type | Modifier | Description                    |
|-----------------------------|------|----------|--------------------------------|
| [naming_mode](#Naming-Mode) | enum | optional | Sets the naming output format. |

## Outbound Options

Available options:

| Name                        | Type | Modifier | Description                                                     |
|-----------------------------|------|----------|-----------------------------------------------------------------|
| export                      | bool | optional | Sets that the message will have an outbound equivalent message. |
| [naming_mode](#Naming-Mode) | enum | optional | Sets the naming output format.                                  |

## Naming Mode

Available options:

| Name                   | Description                                |
|------------------------|--------------------------------------------|
| NAMING_MODE_SNAKE_CASE | Sets the output name case as `snake_case`. |
| NAMING_MODE_CAMEL_CASE | Sets the output name as lower `camelCase`. |
