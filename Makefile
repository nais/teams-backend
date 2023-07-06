BUILDTIME = $(shell date "+%s")
DATE = $(shell date "+%Y-%m-%d")
LAST_COMMIT = $(shell git rev-parse --short HEAD)
LDFLAGS := -X github.com/nais/teams-backend/pkg/version.Revision=$(LAST_COMMIT) -X github.com/nais/teams-backend/pkg/version.Date=$(DATE) -X github.com/nais/teams-backend/pkg/version.BuildUnixTime=$(BUILDTIME)
TEST_POSTGRES_CONTAINER_NAME = teams-backend-postgres-integration-test
TEST_POSTGRES_CONTAINER_PORT = 5666

.PHONY: static teams-backend test generate

all: generate teams-backend

teams-backend:
	go build -o bin/teams-backend -ldflags "-s $(LDFLAGS)" cmd/teams-backend/*.go

local:
	go run ./cmd/teams-backend/main.go

test:
	go test ./...

stop-integration-test-db:
	docker stop $(TEST_POSTGRES_CONTAINER_NAME) || true && docker rm $(TEST_POSTGRES_CONTAINER_NAME) || true

start-integration-test-db: stop-integration-test-db
	docker run -d -e POSTGRES_PASSWORD=postgres --name $(TEST_POSTGRES_CONTAINER_NAME) -p $(TEST_POSTGRES_CONTAINER_PORT):5432 postgres:14-alpine

integration-test: start-integration-test-db
	go test ./... -tags=db_integration_test

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
	go run mvdan.cc/gofumpt -w ./pkg/sqlc/

generate-mocks:
	go run github.com/vektra/mockery/v2 --inpackage --case snake --srcpkg ./pkg/azureclient --name Client
	go run github.com/vektra/mockery/v2 --inpackage --case snake --srcpkg ./pkg/reconcilers/dependencytrack --name Client
	go run github.com/vektra/mockery/v2 --inpackage --case snake --srcpkg ./pkg/reconcilers/github/team --name TeamsService
	go run github.com/vektra/mockery/v2 --inpackage --case snake --srcpkg ./pkg/reconcilers/github/team --name GraphClient
	go run github.com/vektra/mockery/v2 --inpackage --case snake --srcpkg ./pkg/teamsync --name Queue
	go run github.com/vektra/mockery/v2 --inpackage --case snake --srcpkg ./pkg/teamsync --name Handler
	go run github.com/vektra/mockery/v2 --inpackage --case snake --srcpkg ./pkg/auditlogger --name AuditLogger --with-expecter
	go run github.com/vektra/mockery/v2 --inpackage --case snake --srcpkg ./pkg/db --name AuthenticatedUser
	go run github.com/vektra/mockery/v2 --inpackage --case snake --srcpkg ./pkg/db --name Database
	go run github.com/vektra/mockery/v2 --inpackage --case snake --srcpkg ./pkg/authn --name Handler
	go run github.com/vektra/mockery/v2 --inpackage --case snake --srcpkg ./pkg/logger --name Logger
	go run github.com/vektra/mockery/v2 --inpackage --case snake --srcpkg ./pkg/deployproxy --name Proxy
	go run github.com/vektra/mockery/v2 --inpackage --case snake --srcpkg ./pkg/reconcilers --name Reconciler
	go run mvdan.cc/gofumpt -w ./pkg/graph/

static:
	CGO_ENABLED=0 go build -a -installsuffix cgo -o bin/teams-backend -ldflags "-s $(LDFLAGS)" cmd/teams-backend/main.go

docker:
	docker build -t ghcr.io/nais/teams-backend:latest .

rollback:
	echo "TODO: command that git checkouts code that works, leave migrations at HEAD, specify migration version i pkg/db/database.Migrate()"

seed:
	go run cmd/database-seeder/main.go -users 1000 -teams 100 -owners 2 -members 10
