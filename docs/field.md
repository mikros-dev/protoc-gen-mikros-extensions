# Field options

A field message can be customized with the following available options:

| Name                  | Modifier | Description                                                                               |
|-----------------------|----------|-------------------------------------------------------------------------------------------|
| [domain](#domain)     | optional | Options available that affects behavior for the domain version of the field.              |
| [database](#database) | optional | Options available that reflects in the domain version but related to database operations. |

## domain

Available options:

| Name     | Type   | Modifier | Description                                                                  |
|----------|--------|----------|------------------------------------------------------------------------------|
| name     | string | optional | Defines the field name in the domain message.                                |
| tag      | string | optional | Defines the JSON field name tag in the domain message.                       |
| optional | bool   | optional | Sets that the field will be optional in the domain message (i.e. a pointer). |

### Example

```protobuf
message Person {
  string full_name = 1 [(mikros.extensions.domain) = {
    name: "complete_name"
    tag: "Complete_name"
    optional: true
  }];
  
  string address = 2 [(mikros.extensions.domain) = {
    tag: "full_address"
  }];
}
```
```go
type PersonDomain struct {
    CompleteName *string `json:"Complete_name,omitempty"`
    Address      string  `json:"full_address,omitempty"` 
}
```
## database

Available options:

| Name        | Type   | Modifier | Description                                                         |
|-------------|--------|----------|---------------------------------------------------------------------|
| name        | string | optional | Defines the field name inside the database.                         |
| allow_empty | bool   | optional | Sets that the field will exist in the database even if it is empty. |

### Example

```protobuf
message Person {
  string id = 1;
  string name = 2 [(mikros.extensions.database) = {
    name: "person_name"
  }];
  
  int32 age = 3 [(mikros.extensions.database) = {
    allow_empty: true
  }];
}
```
```go
type PersonDomain struct {
    Id   string `json:"id,omitempty"   bson:"_id,omitempty"`
    Name string `json:"name,omitempty" bson:"person_name,omitempty"`
    Age  int32  `json:"age,omitempty", bson:"age"`
}
```