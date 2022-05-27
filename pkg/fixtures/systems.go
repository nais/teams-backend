package fixtures

import (
	"context"

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

func EnsureSystemsExistInDatabase(ctx context.Context, db *gorm.DB) error {
	recs := registry.Reconcilers()
	names := make([]string, 0)
	for name := range recs {
		names = append(names, name)
	}

	for _, systemName := range names {
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
