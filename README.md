# console.nais.io

Console is an API server for team creation and propagation to external systems.

ADR: https://github.com/navikt/pig/blob/master/kubeops/adr/010-console-nais-io.md


## Local development

Console needs Go 1.17, and depends on a PostgreSQL database.
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
* Create a service account
* Enable Workspace Admin API
* Enable `Security -> API Controls -> Domain-wide Delegation`
* Add service account as API client with scopes:
  * `https://www.googleapis.com/auth/admin.directory.group`
  * `https://www.googleapis.com/auth/admin.directory.user.readonly`
* Set up OAuth2 consent screen
  * Internal
* Create OAuth2 client ID
  * Web application
  * 

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

## TODO

* Build synchronization modules
  * GCP
    * Project ID: `hashtrunc(PREFIX-TEAM-CLUSTER)` (6-30 chars, lowercase, `[a-z][a-z0-9-]+[a-z0-9]`)
      * Prefix example: `nais-tenantname`?
    * Project name: `TEAM-CLUSTER`? Human-readable, no limits.
    * JITA access for GCP super admin (nais admins customers' clusters)
  * Kubernetes
    * Connect team group into "nais:developer"
    * Deploy using NAIS deploy

* Implement remainder of GraphQL API
  * Profile endpoint
  * Audit log
  * ACL management

* Better error messages in GraphQL API
  * Don't expose database errors directly