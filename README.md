# console.nais.io

Console is an API server for team creation and propagation to external systems.

ADR: https://github.com/navikt/pig/blob/master/kubeops/adr/010-console-nais-io.md

## TODO

* console.nais.io kan gjøre JITA access for kunde-clustere

* RBAC sync på GCP: legg brukere inn i partner gke-security-groups@kunde.tld

* Determine which API documentation framework to use; candidates are
  * [OpenAPI (with re-doc)](https://redocly.github.io)
  * [RAML](https://raml.org/)
  * [API Blueprint](https://apiblueprint.org/)

* Build minimal API that can be replicated across components without duplicating too much

* Build synchronization core

* Build synchronization modules
  * GCP
  * GitHub
  * NAIS deploy
  * Kubernetes

Finne en vei å gå for å sørge for at vi gjør mest mulig av følgende samtidig som vi integrerer med dokumentasjon:

- Routing (med URL, osv)
- Middleware (authn, authz?)
- Deserialisering av request-objekt
- Funksjonalitet (authz, CRUD, DB)
- Serialisering av riktig response-objekt