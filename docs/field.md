# Field options

A field message can be customized with the following available options:

| Name                  | Modifier | Description                                                                               |
|-----------------------|----------|-------------------------------------------------------------------------------------------|
| [domain](#domain)     | optional | Options available that affects behavior for the domain version of the field.              |
| [database](#database) | optional | Options available that reflects at the domain version but related to database operations. |
| [inbound](#inbound)   | optional | Options available that reflects at the inbound version of the field.                      |
| [outbound](#outbound) | optional | Options available that reflects at the outbound version of the field.                     |
| [validate](#validate) | optional | Validation options for the field.                                                         |

## domain

Available options:

| Name                      | Type    | Modifier | Description                                                                  |
|---------------------------|---------|----------|------------------------------------------------------------------------------|
| name                      | string  | optional | Defines the field name in the domain message.                                |
| allow_empty               | bool    | optional | Sets that the field will be optional in the domain message (i.e. a pointer). |
| [struct_tag](#struct_tag) | message | repeated | Sets optional struct tags for the Domain structure.                          |

### Example

```protobuf
message Person {
  string full_name = 1 [(mikros.extensions.domain) = {
    name: "CompleteName"
    allow_empty: true
  }];
  
  string address = 2 [(mikros.extensions.domain) = {
    name: "FullAddress"
  }];
}
```
```go
type PersonDomain struct {
    CompleteName *string `json:"complete_name"`
    FullAddress  string  `json:"full_address,omitempty"` 
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

## inbound

Available options:

| Name        | Type   | Modifier | Description                                          |
|-------------|--------|----------|------------------------------------------------------|
| name        | string | optional | Defines the field name inside the inbound structure. |

## outbound

Available options:

| Name                                 | Type    | Modifier | Description                                                                   |
|--------------------------------------|---------|----------|-------------------------------------------------------------------------------|
| name                                 | string  | optional | Defines the field name inside the outbound structure.                         |
| hide                                 | bool    | optional | Hide the field in the outbound structure.                                     |
| [bitflag](#bitflag)                  | message | optional | Sets bitflag details about the field.                                         |
| allow_empty                          | bool    | optional | Sets that the field will exist in the outbound structure even if it is empty. |
| [struct_tag](#struct_tag)            | message | repeated | Sets optional struct tags for the Domain structure.                           |
| custom_bind                          | bool    | optional | Sets that the field will have a custom bind API call.                         |
| custom_type                          | string  | optional | Lets the user set the outbound type of field.                                 |
| [custom_import](./message.md#import) | message | optional | An import package required for the custom type.                               |

### bitflag

Available options:

| Name   | Type   | Modifier | Description                                                                          |
|--------|--------|----------|--------------------------------------------------------------------------------------|
| values | string | required | Should point to an enum name which holds all the values that the bitflag represents. |
| prefix | string | required | An prefix string that is present in all enum values.                                 |

## validation

Available options:

| Name             | Type   | Modifier | Description                                                        |
|------------------|--------|----------|--------------------------------------------------------------------|
| [rule](#rules)   | enum   | optional | Sets the validation rule.                                          |
| rule_args        | string | array    | Optional arguments for the rule validator.                         |
| custom_rule      | string | optional | The rule name if `rule` is `FIELD_VALIDATOR_RULE_CUSTOM`.          |
| required         | bool   | optional | Sets that the field is required or not.                            |
| min              | number | optional | Defines the minimum value of the field.                            |
| max              | number | optional | Defines the maximum value of the field.                            |
| max_length       | number | optional | Defines the maximum characters of a string.                        |
| dive             | bool   | optional | Enables validation for array fields.                               |
| required_if      | string | optional | Field is required if another field has a specific value.           |
| required_if_not  | string | optional | Field is required if another field does not have a specific value. |
| required_with    | string | optional | Field is required if other field(s) exists.                        |
| required_without | string | optional | Field is required if other field(s) does not exist.                |
| required_all     | string | optional | Field is required if all fields exist.                             |
| required_any     | string | optional | Field is required if any field exists.                             |
| error_message    | string | optional | Custom error validation message.                                   |
| skip             | bool   | optional | Sets the field to not be validated.                                |

### rules

| Name                        | Description                                                                                               |
|-----------------------------|-----------------------------------------------------------------------------------------------------------|
| FIELD_VALIDATOR_RULE_REGEX  | Uses a regex rule to validate the field.                                                                  |
| FIELD_VALIDATOR_RULE_CUSTOM | Uses a custom validator for the field. Check the [validations](validations.md) documentation for details. |

## struct_tag

Available options:

| Name  | Type   | Modifier | Description         |
|-------|--------|----------|---------------------|
| name  | string | required | Sets the tag name.  |
| value | string | required | Sets the tag value. |
