package fixtures

import (
	"context"

	"github.com/nais/console/pkg/db"
)

const (
	AdminUserName        = "nais-console user"
	AdminUserEmailPrefix = "nais-console" // matches the default nais admin user account in the tenant GCP org
)

// InsertInitialDataset Insert an initial dataset into the database. This will only be executed if there are currently
// no users in the users table.
func InsertInitialDataset(ctx context.Context, database db.Database, tenantDomain string, adminApiKey string) error {
	if adminApiKey != "" {
		admin, err := database.GetUserByEmail(ctx, AdminUserEmailPrefix+"@"+tenantDomain)
		if err != nil {
			return err
		}

		err = database.CreateAPIKey(ctx, adminApiKey, admin.ID)
		if err != nil {
			return err
		}

	}

	return nil
}
