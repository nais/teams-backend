package fixtures

import (
	"context"

	"github.com/nais/console/pkg/dbmodels"
	"github.com/nais/console/pkg/reconcilers/registry"
	"gorm.io/gorm"

	// load modules to have them show up in the registry
	_ "github.com/nais/console/pkg/reconcilers/github/team"
	_ "github.com/nais/console/pkg/reconcilers/google/workspace_admin"
	_ "github.com/nais/console/pkg/reconcilers/nais/deploy"
)

func EnsureSystemsExistInDatabase(ctx context.Context, db *gorm.DB) error {
	systemNames := registry.Reconcilers()

	for systemName := range systemNames {
		sys := &dbmodels.System{
			Name: systemName,
		}

		tx := db.WithContext(ctx).FirstOrCreate(sys, "name = ?", systemName)

		if tx.Error != nil {
			return tx.Error
		}
	}

	return nil
}
