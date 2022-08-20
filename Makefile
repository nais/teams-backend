BUILDTIME = $(shell date "+%s")
DATE = $(shell date "+%Y-%m-%d")
LAST_COMMIT = $(shell git rev-parse --short HEAD)
LDFLAGS := -X github.com/nais/console/pkg/version.Revision=$(LAST_COMMIT) -X github.com/nais/console/pkg/version.Date=$(DATE) -X github.com/nais/console/pkg/version.BuildUnixTime=$(BUILDTIME)

.PHONY: alpine console test generate

all: generate console

console:
	go build -o bin/console -ldflags "-s $(LDFLAGS)" cmd/console/*.go

test:
	go test ./...

generate:
	go run github.com/99designs/gqlgen generate --verbose

generate-sqlc:
	go run github.com/kyleconroy/sqlc/cmd/sqlc generate

alpine:
	go build -a -installsuffix cgo -o bin/console -ldflags "-s $(LDFLAGS)" cmd/console/main.go

docker:
	docker build -t ghcr.io/nais/console:latest .

mocks:
	go run github.com/vektra/mockery/v2 --inpackage --case snake --srcpkg ./pkg/azureclient --name Client
	go run github.com/vektra/mockery/v2 --inpackage --case snake --srcpkg ./pkg/reconcilers/github/team --name TeamsService
	go run github.com/vektra/mockery/v2 --inpackage --case snake --srcpkg ./pkg/reconcilers/github/team --name GraphClient
	go run github.com/vektra/mockery/v2 --inpackage --case snake --srcpkg ./pkg/auditlogger --name AuditLogger
	go run github.com/vektra/mockery/v2 --inpackage --case snake --srcpkg ./pkg/db --name Database
