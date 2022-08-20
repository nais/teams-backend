package registry

import (
	"github.com/nais/console/pkg/auditlogger"
	"github.com/nais/console/pkg/config"
	"github.com/nais/console/pkg/db"
	"github.com/nais/console/pkg/reconcilers"
	azure_group_reconciler "github.com/nais/console/pkg/reconcilers/azure/group"
	console_reconciler "github.com/nais/console/pkg/reconcilers/console"
	github_team_reconciler "github.com/nais/console/pkg/reconcilers/github/team"
	google_gcp_reconciler "github.com/nais/console/pkg/reconcilers/google/gcp"
	google_workspace_admin_reconciler "github.com/nais/console/pkg/reconcilers/google/workspace_admin"
	nais_namespace_reconciler "github.com/nais/console/pkg/reconcilers/nais/namespace"
	"github.com/nais/console/pkg/sqlc"
	log "github.com/sirupsen/logrus"
)

type ReconcilerFactory func(db.Database, *config.Config, auditlogger.AuditLogger) (reconcilers.Reconciler, error)

type ReconcilerInitializer struct {
	Name    sqlc.SystemName
	Factory ReconcilerFactory
}

var (
	recs     = make([]ReconcilerInitializer, 0)
	recNames = make(map[sqlc.SystemName]bool)
)

func Register(systemName sqlc.SystemName, init ReconcilerFactory) {
	if _, exists := recNames[systemName]; exists {
		log.Warnf("reconciler '%s' has already been registered", systemName)
		return
	}

	recs = append(recs, ReconcilerInitializer{
		Name:    systemName,
		Factory: init,
	})
	recNames[systemName] = true
}

func Reconcilers() []ReconcilerInitializer {
	return recs
}

// RegisterReconcilers Register reconcilers in the registry. The order the reconcilers are registered in, is the same as
// the order they will execute in. Reconcilers should be able to run in any order, but they might benefit from a
// specific order if they have soft dependencies (for instance a group created by one reconciler is used by another
// reconciler).
func RegisterReconcilers() {
	Register(console_reconciler.Name, console_reconciler.NewFromConfig)
	Register(azure_group_reconciler.Name, azure_group_reconciler.NewFromConfig)
	Register(github_team_reconciler.Name, github_team_reconciler.NewFromConfig)
	Register(google_workspace_admin_reconciler.Name, google_workspace_admin_reconciler.NewFromConfig)
	Register(google_gcp_reconciler.Name, google_gcp_reconciler.NewFromConfig)
	Register(nais_namespace_reconciler.Name, nais_namespace_reconciler.NewFromConfig)
}
