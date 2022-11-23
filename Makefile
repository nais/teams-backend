BUILDTIME = $(shell date "+%s")
DATE = $(shell date "+%Y-%m-%d")
LAST_COMMIT = $(shell git rev-parse --short HEAD)
LDFLAGS := -X github.com/nais/console/pkg/version.Revision=$(LAST_COMMIT) -X github.com/nais/console/pkg/version.Date=$(DATE) -X github.com/nais/console/pkg/version.BuildUnixTime=$(BUILDTIME)

.PHONY: alpine console test generate

all: generate console migrate

console:
	go build -o bin/console -ldflags "-s $(LDFLAGS)" cmd/console/*.go

migrate:
	go build -o bin/migrate -ldflags "-s $(LDFLAGS)" cmd/legacy-migration/*.go

test:
	go test ./...

fmt:
	go run mvdan.cc/gofumpt -w ./

generate: generate-sqlc generate-gql generate-mocks

check:
	go run honnef.co/go/tools/cmd/staticcheck ./...
	go run golang.org/x/vuln/cmd/govulncheck -v ./...

generate-gql:
	go run github.com/99designs/gqlgen generate --verbose
	go run mvdan.cc/gofumpt -w ./pkg/graph/

generate-sqlc:
	go run github.com/kyleconroy/sqlc/cmd/sqlc generate

generate-mocks:
	go run github.com/vektra/mockery/v2 --inpackage --case snake --srcpkg ./pkg/azureclient --name Client
	go run github.com/vektra/mockery/v2 --inpackage --case snake --srcpkg ./pkg/reconcilers/github/team --name TeamsService
	go run github.com/vektra/mockery/v2 --inpackage --case snake --srcpkg ./pkg/reconcilers/github/team --name GraphClient
	go run github.com/vektra/mockery/v2 --inpackage --case snake --srcpkg ./pkg/auditlogger --name AuditLogger
	go run github.com/vektra/mockery/v2 --inpackage --case snake --srcpkg ./pkg/db --name Database
	go run github.com/vektra/mockery/v2 --inpackage --case snake --srcpkg ./pkg/authn --name Handler

alpine:
	go build -a -installsuffix cgo -o bin/console -ldflags "-s $(LDFLAGS)" cmd/console/main.go

docker:
	docker build -t ghcr.io/nais/console:latest .

rollback:
	echo "TODO: command that git checkouts code that works, leave migrations at HEAD, specify migration version i pkg/db/database.Migrate()"
