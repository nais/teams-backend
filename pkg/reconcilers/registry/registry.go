package registry

import (
	"github.com/nais/console/pkg/auditlogger"
	"github.com/nais/console/pkg/config"
	"github.com/nais/console/pkg/reconcilers"
	"github.com/nais/console/pkg/sqlc"
	log "github.com/sirupsen/logrus"
	"gorm.io/gorm"

	azure_group_reconciler "github.com/nais/console/pkg/reconcilers/azure/group"
	console_reconciler "github.com/nais/console/pkg/reconcilers/console"
	github_team_reconciler "github.com/nais/console/pkg/reconcilers/github/team"
	google_gcp_reconciler "github.com/nais/console/pkg/reconcilers/google/gcp"
	google_workspace_admin_reconciler "github.com/nais/console/pkg/reconcilers/google/workspace_admin"
	nais_namespace_reconciler "github.com/nais/console/pkg/reconcilers/nais/namespace"
)

type ReconcilerFactory func(sqlc.Querier, *gorm.DB, *config.Config, sqlc.System, auditlogger.AuditLogger) (reconcilers.Reconciler, error)

type ReconcilerInitializer struct {
	Name    string
	Factory ReconcilerFactory
}

var (
	recs     = make([]ReconcilerInitializer, 0)
	recNames = make(map[string]bool)
)

func Register(name string, init ReconcilerFactory) {
	if _, exists := recNames[name]; exists {
		log.Warnf("reconciler '%s' has already been registered", name)
		return
	}

	recs = append(recs, ReconcilerInitializer{
		Name:    name,
		Factory: init,
	})
	recNames[name] = true
}

func Reconcilers() []ReconcilerInitializer {
	return recs
}

// registerReconcilers Register reconcilers in the registry. The order the reconcilers are registered in, is the same as
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
