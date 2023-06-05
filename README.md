# NAIS teams-backend

`teams-backend` is an API server for NAIS team management, along with propagation to external systems.

ADR: https://github.com/nais/core/blob/main/adr/010-console-nais-io.md

## Local development

`teams-backend` needs a go version as per the go.mod file, and depends on a running PostgreSQL database. For convenience, a [Docker Compose](https://docs.docker.com/compose/) configuration is provided.

```sh
docker compose up -d
make teams-backend
./bin/teams-backend
```

Running the compiled binary without any arguments will start an instance that does not touch any external systems. The API server runs GraphiQL at http://localhost:3000.

In order to make any request to the API server, your requests must be authenticated with a `Authorization: Bearer <api-key>` header. For local development the `TEAMS_BACKEND_STATIC_SERVICE_ACCOUNTS` environment variable can be used to create a service account with the necessary permissions. Refer to the configuration documentation below for more information.

For a combination of more tools running locally ([hookd](https://github.com/nais/deploy), [Teams frontend](https://github.com/nais/teams-frontend) and more), check out the [nais/features-dev](https://github.com/nais/features-dev) repo.

## Configuration

`teams-backend` can be configured using environment variables. For a complete list of possible configuration values along with documentation refer to the [pkg/config](pkg/config/config.go) package. Some configuration values that require more in depth documentation can be found below.

### GCP clusters (`TEAMS_BACKEND_GCP_CLUSTERS`)

JSON-encoded object with information about the GCP clusters to use with `teams-backend`.

Example:

```json
{
  "dev": {
    "teams_folder_id": "123456789012",
    "project_id": "project-id-123"
  },
  "prod": {
    "teams_folder_id": "123456789013",
    "project_id": "project-id-456"
  }
}
```

The keys in the object refer to the environment names to use. In the example above we have two environments, `dev` and `prod`. Each environment maps to a JSON-object with two keys:

- `teams_folder_id`: The numeric ID of the `teams` folder in the given environment, where all team projects will be created.
- `project_id`: The ID of the GCP project for the environment/cluster.

### Static service accounts (`TEAMS_BACKEND_STATIC_SERVICE_ACCOUNTS`)

`teams-backend` can create a list of service accounts with predefined API keys and roles on start up.

Example:

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

The service account names **must** begin with `nais-`. Each role in the `roles` list must be valid role names. Role names can be fetched from the GraphQL API:

```graphql
query {
    roles
}
```

`teams-backend` will, on each start up of the application, ensure that the service accounts specified in the JSON value exists. If a service account is removed from the JSON value, `teams-backend` will remove it from the database as well.

### Running locally
A config file for minimal working local setup with usersync / login (dev-nais.io user required) enabled can be downloaded like this:
```sh
gcloud --project="nais-management-7178" secrets versions access --secret="teams-backend-local-config" latest # (nais.io users have access to this secret)
```

## GraphQL API

`teams-backend` runs [GraphiQL](https://github.com/graphql/graphiql) by default, and this can be used to explore and use the GraphQL API. After you have started `teams-backend` you will find graphiql on http://localhost:3000. Use the `TEAMS_BACKEND_STATIC_SERVICE_ACCOUNTS` configuration parameter to create a service accounut to use with the API.

Some common queries is listed below.

### Fetch teams:

```graphql
{
  teams {
    slug
    purpose
    members {
      user {
        email
        name
      }
      role
    }
  }
}
```

### Fetch roles:

```graphql
{
  roles
}
```

## Reconcilers

`teams-backend` uses reconcilers to sync team information to external systems, for instance GitHub or Azure AD. The supported reconcilers can be configured with a combination of environment variables and configuration options set through the GraphQL API. By default all reconcilers are disabled when `teams-backend` starts up. To enable a reconciler, the `enableReconciler` mutation in the GraphQL API can be used. Keep in mind that the reconciler enabled status is persisted in the database, so if you enable one or more reconcilers they will still be enabled the next time you start up the `teams-backend` application, unless you start up with an empty database.

The implemented reconcilers and their purpose are documented below.

### Google Workspace

To create groups in a Google Workspace organization and sync members the `google:workspace-admin` reconciler can be enabled. Once a team is created in `teams-backend` the reconciler will create a group for the team in the connected Google Workspace. Given a `teams-backend` team with a slug set to `my-team` the Google workspace group will end up with:

- Email: `nais-team-my-team@<domain>`
- Name: `nais-team-my-team`

When a user is added / removed to the team in `teams-backend` the reconciler will make the same change in the Google Workspace group.

### GCP Projects

Each team in `teams-backend` can get a GCP project by using the `google:gcp:project` reconciler. The reconciler will create a project in each cluster that `teams-backend` is configured to use. When creating a project the team group will be set as the owner of the project.

### NAIS namespace

To generate NAIS namespaces for a team in the configured cluster the `nais:namespace` reconciler can be used.

### Azure AD groups

The `azure:group` reconciler works in a similar fashion as the Google Workspace one, but instead it will create a security group in Azure AD. The Azure AD tenant must share the same domain as the Google Workspace, and the email address of the users must match up for `teams-backend` to correctly identify the users.

### GitHub teams

The `github:team` reconciler can create a GitHub team for each `teams-backend` team, and maintain team memberships based on the information found in `teams-backend`. To use this reconciler a [GitHub App](https://docs.github.com/en/developers/apps/getting-started-with-apps/about-apps) must exist. The app requires the following scopes:

- Organization Administration: `read`
- Organization Members: `readwrite`

Install the application on the organization and obtain the private key, application ID, and installation ID.

### NAIS deploy key

To generate NAIS deploy keys for each `teams-backend` team the `nais:deploy` reconciler can be used.

## Bootstrapping other systems

Refer to the [NAIS docs](https://naas.nais.io/technical/tenant-setup/) for botstrapping other systems to work with `teams-backend`.

## Verifying the `teams-backend` image and its contents

The image is signed "keylessly" (is that a word?) using [Sigstore cosign](https://github.com/sigstore/cosign).
To verify its authenticity run
```
cosign verify \
--certificate-identity "https://github.com/nais/teams-backend/.github/workflows/build_and_push_image.yaml@refs/heads/main" \
--certificate-oidc-issuer "https://token.actions.githubusercontent.com" \
europe-north1-docker.pkg.dev/nais-io/nais/images/teams-backend@sha256:<shasum> 
```

The images are also attested with SBOMs in the [CycloneDX](https://cyclonedx.org/) format.
You can verify these by running
```
cosign verify-attestation --type cyclonedx \
--certificate-identity "https://github.com/nais/teams-backend/.github/workflows/build_and_push_image.yaml@refs/heads/main" \
--certificate-oidc-issuer "https://token.actions.githubusercontent.com" \
europe-north1-docker.pkg.dev/nais-io/nais/images/teams-backend@sha256:<shasum>
```

## License

`teams-backend` is licensed under the MIT License, see [LICENSE.md](LICENSE.md).
