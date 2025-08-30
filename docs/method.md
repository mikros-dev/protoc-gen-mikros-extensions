# RPC Method options

An RPC method has the following options available to be used:

| Name          | Type    | Modifier | Description                          |
|---------------|---------|----------|--------------------------------------|
| [http](#http) | message | optional | Sets HTTP information about the RPC. |

## HTTP

Available options:

| Name                     | Type   | Modifier | Description                                                                                                                                |
|--------------------------|--------|----------|--------------------------------------------------------------------------------------------------------------------------------------------|
| auth_arg                 | string | array    | Sets authorization values for the RPC.                                                                                                     |
| header                   | string | array    | Sets header variables that the RPC will have.                                                                                              |
| parse_request_in_service | bool   |          | Enables or disables the generated code for parsing the request message<br> in the handler, i.e, it will be client responsibility to parse. |
