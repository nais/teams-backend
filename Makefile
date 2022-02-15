BUILDTIME = $(shell date "+%s")
DATE = $(shell date "+%Y-%m-%d")
LAST_COMMIT = $(shell git rev-parse --short HEAD)
LDFLAGS := -X github.com/nais/console/pkg/version.Revision=$(LAST_COMMIT) -X github.com/nais/console/pkg/version.Date=$(DATE) -X github.com/nais/console/pkg/version.BuildUnixTime=$(BUILDTIME)

.PHONY: alpine console test migration

console:
	go build -o bin/console -ldflags "-s $(LDFLAGS)" cmd/console/*.go

test:
	go test ./...

migration:
	go generate ./...

alpine:
	go build -a -installsuffix cgo -o bin/console -ldflags "-s $(LDFLAGS)" cmd/console/main.go

docker:
	docker build -t ghcr.io/nais/console:latest .
