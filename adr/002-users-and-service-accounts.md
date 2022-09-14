# ADR - Users and service accounts
**TL;DR** *A service account is a type of user that is used by other systems to interact with the Console GraphQL API, while a user is a developer that interacts with Console through the frontend UI. Service accounts authenticate using API keys, while users authenticate using OAuth 2.0.*

## Background
Console exposes a [GraphQL API](https://graphql.org/) that can be used to do team-related interactions. To be able to access this API in a controlled manner we need a concept surrounding users and authentication. Authorization is beyond the scope of this ADR.

## Solution
Console has two different classes of users:

- Regular users (developers)
- Service accounts (machines)

### Users
A user corresponds to a developer, and is identified through a unique email address. Users will not typically interact with Console through the GraphQL API directly, but instead through the [Console frontend UI](https://github.com/nais/console-frontend-elm). For a user to exist in Console it needs to be synced from the tenants GCP organization. Console stores the name and email address of each user entry.

#### User sync
Console considers the tenants GCP organization to be the single source of truth when it comes to user accounts, and will continuously syncronize all users from GCP to Console. It is not possible to manually create Console user accounts through the GraphQL API, so the only way to add users to Console is through the user sync. It is not possible to update existing user information in Console through the API either, this must be done in the GCP organization, and changes to user accounts will be mirrored to the user accounts in Console through the user sync.

When a user is removed from the GCP organization, the user sync will remove the user from the Console database, along with all existing team connections and other relations.

### Service accounts
A service account corresponds to one (or several) external systems, and is identified through a unique name. The name must start with a lowercase letter, and can consist of lowercase letters, numbers and hyppens. The name can not end with a hyphen. Service accounts must use API keys to interact with the GraphQL API, and are not able to sign in using the frontend.

Service accounts can only be created via the `CONSOLE_STATIC_SERVICE_ACCOUNTS` environment variable that must be passed to Console on startup. The environment variable must contain a JSON string that matches the following schema:

```json
{
  "$schema": "https://json-schema.org/draft/2020-12/schema",
  "$id": "https://github.com/nais/console/service-accounts-schema.json",
  "title": "Service accounts",
  "description": "JSON array for creating service accounts during Console startup",
  "type": "array",
  "items": [
    {
      "type": "object",
      "properties": {
        "name": {
          "description": "The unique name of the service account. Must have the `nais-` prefix.",
          "type": "string"
        },
        "apiKey": {
          "description": "The API key of the service account.",
          "type": "string"
        },
        "roles": {
          "description": "Roles that will be granted to the service account.",
          "type": "array",
          "items": {
            "type": "string"
          }
        }
      },
      "required": [
        "name",
        "apiKey",
        "roles"
      ]
    }
  ]
}
```

Following is an example value that would end up creating two different service accounts, with different API keys and different roles:

```json
[
  {
    "name": "nais-service-account-1",
    "apiKey": "api-key-1",
    "roles": ["Team viewer"]
  },
  {
    "name": "nais-service-account-2",
    "apiKey": "api-key-2",
    "roles": ["Team owner"]
  }
]
```

The role names must be one of the supported roles in Console. A complete list of available roles can be fetched from the GraphQL API. When a service account is removed from the environment variable it will be removed when Console restarts.

#### Limitations
Service accounts are currently only meant to be used by the NAIS team, and service accounts are not allowed as team members / owners.