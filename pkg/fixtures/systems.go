package fixtures

import (
	"context"
	log "github.com/sirupsen/logrus"

	"github.com/nais/console/pkg/dbmodels"
	"github.com/nais/console/pkg/reconcilers/registry"
	"gorm.io/gorm"

	// load modules to have them show up in the registry
	_ "github.com/nais/console/pkg/reconcilers/azure/group"
	_ "github.com/nais/console/pkg/reconcilers/console"
	_ "github.com/nais/console/pkg/reconcilers/github/team"
	_ "github.com/nais/console/pkg/reconcilers/google/gcp"
	_ "github.com/nais/console/pkg/reconcilers/google/workspace_admin"
	_ "github.com/nais/console/pkg/reconcilers/nais/deploy"
	_ "github.com/nais/console/pkg/reconcilers/nais/namespace"
)

func EnsureSystemsExistInDatabase(ctx context.Context, db *gorm.DB) (map[string]*dbmodels.System, error) {
	recs := registry.Reconcilers()
	systems := make(map[string]*dbmodels.System)
	for name := range recs {
		system := &dbmodels.System{
			Name: name,
		}

		log.Infof(`Ensure system "%s" exists in the database...`, name)
		tx := db.WithContext(ctx).FirstOrCreate(system, "name = ?", name)

		if tx.Error != nil {
			return nil, tx.Error
		}

		systems[name] = system
	}

	log.Infof("All systems have been added to the database.")

	return systems, nil
}