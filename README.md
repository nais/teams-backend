# console.nais.io

Console is an API server for team creation and propagation to external systems.

## TODO

* Determine which API documentation framework to use; candidates are
  * [OpenAPI (with re-doc)](https://redocly.github.io)
  * [RAML](https://raml.org/)
  * [API Blueprint](https://apiblueprint.org/)

* Build walking skeleton (hello world, CI, deploy to production)

* Build minimal API that can be replicated across components without duplicating too much

* Build synchronization core

* Build synchronization modules
  * GCP
  * GitHub
  * NAIS deploy
  * Kubernetes