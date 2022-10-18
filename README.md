# console.nais.io

Console is an API server for team creation and propagation to external systems.

ADR: https://github.com/navikt/pig/blob/master/kubeops/adr/010-console-nais-io.md

## Configuration

Console is configured using environment variables:

### `CONSOLE_AUTO_LOGIN_USER`

Auto login a specific user on all requests against the GraphQL API. This setting should **NEVER** be used in a production environment.

### `CONSOLE_DATABASE_URL`

The URL for the database. Defaults to `postgres://console:console@localhost:3002/console?sslmode=disable` which works for local development purposes.

### `CONSOLE_FRONTEND_URL`

URL to the console frontend. Defaults to `http://localhost:3001`. The frontend is in a [separate repository](https://github.com/nais/console-frontend).

### `CONSOLE_LISTEN_ADDRESS`

The host:port combination used by the http server. Defaults to `127.0.0.1:3000`.

### `CONSOLE_LOG_FORMAT`

Customize the log format. Defaults to `text`. Can be set to `json`.

### `CONSOLE_LOG_LEVEL`

The log level used in console. Defaults to `DEBUG`.

### `CONSOLE_TENANT_DOMAIN`

The domain for the tenant. Defaults to `example.com`.

### `CONSOLE_ADMIN_API_KEY`

Can be used to create an API key for the initial admin user. Used for local development when user sync is not enabled, and will only be used for the initial dataset.

### `CONSOLE_STATIC_SERVICE_ACCOUNTS`

Can be used to create a set of service accounts with roles and API keys. The value must be JSON-encoded. Example:

```json
[
  {
    "name": "nais-service-account-1",
    "apiKey": "key1",
    "roles": ["Team viewer", "User viewer"]
  },
  {
    "name": "nais-service-account-2",
    "apiKey": "key2",
    "roles": ["Team creator"]
  }
]
```
The names **must** begin with `nais-`. This is a reserved prefix, and service accounts with this prefix can not be created
using the GraphQL API.

The roles specified must be valid role names. See [pkg/roles/roles.go](pkg/roles/roles.go) for all role names.

The service accounts will be re-created every time the application starts. Use the API to delete one or more of 
the service accounts, it is not sufficient to simply remove the service account from the JSON structure in the environment 
variable.
## Reconcilers

Console uses reconcilers to sync team information to external systems, for instance GitHub or Azure AD. All reconcilers
must be enabled via environment variables, and require different settings to work as expected. All configuration values
is mentioned below.

### Google Workspace

To create groups in Google Workspace and sync members you will need the following environment variables set:

#### `CONSOLE_GOOGLE_DELEGATED_USER`

A user account (email address) that has admin rights in the Google Workspace account.

#### `CONSOLE_GOOGLE_CREDENTIALS_FILE`

JSON file that contains the private key of the service account.

### GCP Projects

To create projects for the team in GCP you will need to set the following environment variables:

#### `CONSOLE_GCP_CLUSTERS`

JSON-encoded object with info about the clusters:

```json
{
  "<env>": {
    "team_folder_id": 123456789012,
    "cluster_project_id": "<id>"
  },
  ...
}
```

where `<env>` can be for example `prod`, and `<id>` is for example `nais-dev-abcd`.

#### `CONSOLE_GCP_CNRM_ROLE`

The name of the custom CNRM role. The value also contains the org ID. Example:

`organizations/<org_id>/roles/CustomCNRMRole`

where `<org_id>` is a numeric ID.

#### `CONSOLE_GCP_BILLING_ACCOUNT`

The ID of the billing account that each team project will use.

### NAIS namespace

To generate NAIS namespaces for a team the following environment variables must be set:

#### `CONSOLE_NAIS_NAMESPACE_PROJECT_ID`

The ID of the NAIS management project in the tenant organization in GCP.

## Local development

Console needs Go 1.18, and depends on a PostgreSQL database.
For convenience, a Docker Compose configuration is provided.

Running the compiled binary without any arguments will start an instance that
does not touch any external systems. The API server runs GraphQL at http://localhost:3000.

In order to make any request to the API server, your requests must be authenticated
with a `Authorization: Bearer <APIKEY>` header. This API key lives in the database table `api_keys`.
Future work involves setting these credentials up automatically.

```sh
docker-compose up
make
bin/console
```


## Bootstrapping other systems

### GCP
* Create a service account (automated via [nais/nais-terraform-modules](https://github.com/nais/nais-terraform-modules))
* Enable Workspace Admin API (automated via [nais/nais-terraform-modules](https://github.com/nais/nais-terraform-modules))
* Set up OAuth2 consent screen in the nais-management project in the tenant org:
  ```
  gcloud alpha iap oauth-brands create \
  --application_title=Console \
  --support_email=SUPPORT_EMAIL \
  --project=PROJECT_ID
  ```
* Create OAuth2 client ID:
  ```
  gcloud alpha iap oauth-clients create \
  projects/PROJECT_ID/brands/BRAND-ID \
  --display_name=Console
  ``` 
* 
  * Type: Web application
  * Name: Console
  * Authorized redirect URIs:
    * http://localhost:3000/oauth2/callback

### Google Workspace
* Enable `Security -> API Controls -> Domain-wide Delegation`
* Add service account as API client with scopes:
  * `https://www.googleapis.com/auth/admin.directory.group`
  * `https://www.googleapis.com/auth/admin.directory.user.readonly`

### Github
* Set up single sign-on against tenant's IdP. SCIM is recommended, but not required.
* Create GitHub application in your organization, with all features disabled, and with scopes:
  * Organization Administration: `read`
  * Organization Members: `readwrite`
* Install GitHub application on organization and obtain private key, application ID, and installation ID

Important: do not share the same GitHub application between tenants!