# Service options

A service has the following options available:

| Name                            | Modifier | Description                             |
|---------------------------------|----------|-----------------------------------------|
| [authorization](#authorization) | optional | Sets authorization details for the RPC. |

### authorization

| Name           | Type | Modifier | Description                              |
|----------------|------|----------|------------------------------------------|
| [mode](#modes) | enum | required | Sets the authorization mode for the RPC. |

#### modes

| Name                       | Description                                   |
|----------------------------|-----------------------------------------------|
| AUTHORIZATION_MODE_NO_AUTH | Service without authorization checking.       |
| AUTHORIZATION_MODE_SCOPED  | Service with scopes authorization validation. |      
