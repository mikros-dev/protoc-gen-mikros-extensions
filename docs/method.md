# RPC Method options

An RPC method has the following options available to be used:

| Name          | Type    | Modifier | Description                           |
|---------------|---------|----------|---------------------------------------|
| [http](#http) | message | optional | Sets HTTP informations about the RPC. |

## HTTP

Available options:

| Name   | Type   | Modifier | Description                                   |
|--------|--------|----------|-----------------------------------------------|
| scope  | string | array    | Sets authorization scopes for the RPC.        |
| header | string | array    | Sets header variables that the RPC will have. |
