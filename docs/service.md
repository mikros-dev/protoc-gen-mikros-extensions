# Service options

A service has the following options available:

| Name                            | Modifier | Description                             |
|---------------------------------|----------|-----------------------------------------|
| [authorization](#authorization) | optional | Sets authorization details for the RPC. |

### authorization

| Name             | Type   | Modifier | Description                                                         |
|------------------|--------|----------|---------------------------------------------------------------------|
| [mode](#modes)   | enum   | required | Sets the authorization mode for the RPC.                            |
| custom_auth_name | string | optional | Sets the authorization name when mode is AUTHORIZATION_MODE_CUSTOM. |

#### modes

| Name                       | Description                                   |
|----------------------------|-----------------------------------------------|
| AUTHORIZATION_MODE_NO_AUTH | Service without authorization checking.       |
| AUTHORIZATION_MODE_CUSTOM  | Service with scopes authorization validation. |
