# Enum options

The following options can be used inside an enum declaration:

| Name        | Modifier | Description                                               |
|-------------|----------|-----------------------------------------------------------|
| [api](#api) | optional | Set which APIs extensions the enum will have implemented. |

## api

Available options:

| Name                        | Type | Modifier | Description                                                                                                                                 |
|-----------------------------|------|----------|---------------------------------------------------------------------------------------------------------------------------------------------|
| [bitflag](#bitflag-example) | bool | optional | Set that the enum will act as a bit flag mask value set.<br>The field that uses this enum must be a **uint64** rather than the proper enum. |

### bitflag example

```protobuf
enum SomeSimpleExample {
  option (mikros.extensions.api) = {
    bitflag: true
  };

  SOME_SIMPLE_EXAMPLE_UNSPECIFIED = 0;
  SOME_SIMPLE_EXAMPLE_ON = 1;
  SOME_SIMPLE_EXAMPLE_OFF = 2;
  SOME_SIMPLE_EXAMPLE_READ = 3;
  SOME_SIMPLE_EXAMPLE_WRITE = 4;
}
```
```go
// Using the generated API code
func Example() {
    values := []pb.SomeSipleExample{
        pb.SomeSimpleExample_SOME_SIMPLE_EXAMPLE_ON,
        pb.SomeSimpleExample_SOME_SIMPLE_EXAMPLE_OFF,
        pb.SomeSimpleExample_SOME_SIMPLE_EXAMPLE_READ,
        pb.SomeSimpleExample_SOME_SIMPLE_EXAMPLE_WRITE,
    }
    
    for _, e := range values {
        fmt.Println(e.Bitflag())
    }

    // Should print
    // 1
    // 2
    // 4 
    // 8
}
```

# Enum values

Enum values have the following properties available:

| Name            | Modifier | Description               |
|-----------------|----------|---------------------------|
| [entry](#entry) | optional | Customize the enum entry. |

## entry

Available options:

| Name | Type   | Modifier | Description                                                                                               |
|------|--------|----------|-----------------------------------------------------------------------------------------------------------|
| name | string | optional | Provide an alternative name for the enum entry. It can be accessed <br> using the `EntryName()` API call. |

### Entry example

```protobuf
enum SomeExample {
  SOME_EXAMPLE_UNSPECIFIED = 0 [(mikros.extensions.entry) = {name:"unknown"}];
}
```
```go
func Example() {
    e := pb.SomeExample_SOME_EXAMPLE_UNSPECIFIED
    fmt.Println(e.EntryName())

    // Should print:
    // unknown
}
```