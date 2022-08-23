package fixtures

import (
	"context"

	"github.com/nais/console/pkg/sqlc"

	"github.com/nais/console/pkg/db"
)

const (
	AdminUserName        = "nais-console user"
	AdminUserEmailPrefix = "nais-console" // matches the default nais admin user account in the tenant GCP org
)

// InsertInitialDataset Insert an initial dataset into the database
func InsertInitialDataset(ctx context.Context, database db.Database, tenantDomain string, adminApiKey string) error {
	if adminApiKey != "" {
		_, err := database.GetUserByEmail(ctx, AdminUserEmailPrefix+"@"+tenantDomain)
		if err != nil {
			admin, err := database.AddUser(ctx, AdminUserName, AdminUserEmailPrefix+"@"+tenantDomain)
			if err != nil {
				return err
			}

			err = database.CreateAPIKey(ctx, adminApiKey, admin.ID)
			if err != nil {
				return err
			}

			err = database.AssignGlobalRoleToUser(ctx, admin.ID, sqlc.RoleNameAdmin)
			if err != nil {
				return err
			}
		}
	}

	return nil
}
