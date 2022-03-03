# console.nais.io

Console is an API server for team creation and propagation to external systems.

ADR: https://github.com/navikt/pig/blob/master/kubeops/adr/010-console-nais-io.md

## TODO

* Build synchronization core

* Build synchronization modules
  * GCP
    * JITA access for GCP super admin (nais admins customers' clusters)
    * RBAC sync: create groups and add them to gke-security-groups@<domain>
  * GitHub
  * NAIS deploy
  * Kubernetes

* Implement ACLs

* Implement remainder of GraphQL API