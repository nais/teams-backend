# console.nais.io

Console is an API server for team creation and propagation to external systems.

ADR: https://github.com/navikt/pig/blob/master/kubeops/adr/010-console-nais-io.md

## TODO

* Build synchronization core

* Build synchronization modules
  * GCP
    * Project ID: `hashtrunc(PREFIX-TEAM-CLUSTER)` (6-30 chars, lowercase, `[a-z][a-z0-9-]+[a-z0-9]`)
      * Prefix example: `nais-tenantname`?
    * Project name: `TEAM-CLUSTER`? Human-readable, no limits.
    * JITA access for GCP super admin (nais admins customers' clusters)
    * RBAC sync: create groups and add them to `gke-security-groups@<domain>`. Ends up in rolebinding "groups" field.
  * GitHub
    * Teams, members, and ACL
  * NAIS deploy
    * How do we get our self-provisioned API key?
      * Provisioned by naisd, together with other credentials we need (gcp, github, etc.)
  * Kubernetes
    * Connect team group into "nais:developer"
    * Deploy using NAIS deploy

* Implement ACLs

* Implement remainder of GraphQL API