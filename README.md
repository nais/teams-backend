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

### `CONSOLE_PARTNER_DOMAIN`

The domain for the partner. Defaults to `example.com`.

## Reconcilers

Console uses reconcilers to sync team information to external systems, for instance GitHub or Azure AD. All reconcilers
must be enabled via environment variables, and require different settings to work as expected. All configuration values
is mentioned below.

### GitHub

To create teams on GitHub and sync members you will need the following environment variables set:

#### `CONSOLE_GITHUB_ENABLED`

Set to `true` to enable the reconciler.

#### `CONSOLE_GITHUB_ORGANIZATION`

The slug of the organization the app is installed on.

#### `CONSOLE_GITHUB_APP_ID`

The application ID of the GitHub Application that Console will use when communicating with the GitHub APIs. The 
application will need the following permissions:

| Permission                  | Access         |
|-----------------------------|----------------|
| Organization administration | Read-only      |
| Organization members        | Read and write |

#### `CONSOLE_GITHUB_APP_INSTALLATION_ID`

The installation ID for the application when installed to the org.

#### `CONSOLE_GITHUB_PRIVATE_KEY_PATH`

Path to the private key file (PEM format).

### Azure AD

To create groups in Azure AD and sync members you will need the following environment variables set:

#### `CONSOLE_AZURE_ENABLED`

Set to `true` to enable the reconciler.

#### `CONSOLE_AZURE_CLIENT_ID`

The client ID of the application registration. The app needs the following API permissions:

| Permission                | Type        |
|---------------------------|-------------|
| Group.Create              | Application |
| GroupMember.ReadWrite.All | Application |

#### `CONSOLE_AZURE_CLIENT_SECRET`

The client secret.

#### `CONSOLE_AZURE_TENANT_ID`

The tenant ID.

### Google Workspace

To create groups in Google Workspace and sync members you will need the following environment variables set:

#### `CONSOLE_GOOGLE_ENABLED`

Set to `true` to enable the reconciler.

#### `CONSOLE_GOOGLE_DELEGATED_USER`

A user account (email address) that has admin rights in the Google Workspace account.

#### `CONSOLE_GOOGLE_CREDENTIALS_FILE`

JSON file that contains the private key of the service account.

### GCP Projects

To create projects for the team in GCP you will need to set the following environment variables:

#### `CONSOLE_GCP_ENABLED`

Set to `true` to enable the reconciler.

#### `CONSOLE_GCP_PROJECT_PARENT_IDS`

Comma-separated list of `environment:parent_folder_id` values, where environment is appended to the project name for the
team.

### NAIS namespace

To generate NAIS namespaces for a team the following environment variables must be set:

#### `CONSOLE_NAIS_NAMESPACE_ENABLED`

Set to `true` to enable the reconciler.

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
* Set up single sign-on against tenant's IDP. SCIM is optional, but not required.
* Create GitHub application and obtain: private key, application ID, installation ID
* Install GitHub application on organization and give scopes:
  * Organization Administration: `read`
  * Organization Members: `readwrite`

Important: do not share the same GitHub application between tenants.

## ACL

Within a team, users are either _owners_ or _members_. This maps somewhat accurately
to our target systems.

* Every user has a set of basic rights

User roles:

* Global
  * `nais:console:teams:admin` -> manage all teams
  * `nais:console:teams:create` -> create team
* Per team
  * `nais:console:teams:<team>:admin` -> manage team

Team roles:

* Per reconciler: disabled, read, readwrite
  * Only implement "readwrite", and make it look like a boolean option


# What should we test?

* reconcilers
* user synchronizer
* authentication middleware
