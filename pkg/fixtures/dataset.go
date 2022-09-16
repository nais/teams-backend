package fixtures

import (
	"context"

	"github.com/nais/console/pkg/sqlc"

	"github.com/nais/console/pkg/db"
)

const (
	AdminServiceAccountName = "admin"
)

// CreateAdminServiceAccount Create an admin service account with a specific API key
func CreateAdminServiceAccount(ctx context.Context, database db.Database, adminApiKey string) error {
	if adminApiKey == "" {
		return nil
	}

	return database.Transaction(ctx, func(ctx context.Context, dbtx db.Database) error {
		admin, err := dbtx.GetServiceAccountByName(ctx, AdminServiceAccountName)
		if err != nil {
			admin, err = dbtx.CreateServiceAccount(ctx, AdminServiceAccountName)
			if err != nil {
				return err
			}
		}

		err = dbtx.RemoveApiKeysFromServiceAccount(ctx, admin.ID)
		if err != nil {
			return err
		}

		err = dbtx.CreateAPIKey(ctx, adminApiKey, admin.ID)
		if err != nil {
			return err
		}

		err = dbtx.AssignGlobalRoleToUser(ctx, admin.ID, sqlc.RoleNameAdmin)
		if err != nil {
			return err
		}

		return nil
	})
}
