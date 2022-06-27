package fixtures

import (
	"github.com/nais/console/pkg/reconcilers/azure/group"
	"github.com/nais/console/pkg/reconcilers/console"
	"github.com/nais/console/pkg/reconcilers/github/team"
	"github.com/nais/console/pkg/reconcilers/google/gcp"
	"github.com/nais/console/pkg/reconcilers/google/workspace_admin"
	"github.com/nais/console/pkg/reconcilers/nais/namespace"
	log "github.com/sirupsen/logrus"

	"github.com/nais/console/pkg/dbmodels"
	"github.com/nais/console/pkg/reconcilers/registry"
	"gorm.io/gorm"
)

// registerReconcilers Register reconcilers in the registry. The order the reconcilers are registered in, is the same as
// the order they will execute in. Reconcilers should be able to run in any order, but they might benefit from a
// specific order if they have soft dependencies (for instance a group created by one reconciler is used by another
// reconciler).
func registerReconcilers() {
	registry.Register(console_reconciler.Name, console_reconciler.NewFromConfig)
	registry.Register(azure_group_reconciler.Name, azure_group_reconciler.NewFromConfig)
	registry.Register(github_team_reconciler.Name, github_team_reconciler.NewFromConfig)
	registry.Register(google_workspace_admin_reconciler.Name, google_workspace_admin_reconciler.NewFromConfig)
	registry.Register(google_gcp_reconciler.Name, google_gcp_reconciler.NewFromConfig)
	registry.Register(nais_namespace_reconciler.Name, nais_namespace_reconciler.NewFromConfig)
}

// CreateReconcilerSystems Ensure system entries exists in the database for all reconcilers
func CreateReconcilerSystems(db *gorm.DB) (map[string]*dbmodels.System, error) {
	registerReconcilers()

	recs := registry.Reconcilers()
	systems := make(map[string]*dbmodels.System)
	for _, rec := range recs {
		system := &dbmodels.System{
			Name: rec.Name,
		}

		log.Infof("Ensure system '%s' exists in the database...", rec.Name)
		err := db.Where("name = ?", rec.Name).FirstOrCreate(system).Error
		if err != nil {
			return nil, err
		}

		systems[rec.Name] = system
	}

	log.Infof("All systems have been added to the database.")

	return systems, nil
}
