package fixtures_test

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/nais/console/pkg/db"
	"github.com/nais/console/pkg/fixtures"
	"github.com/nais/console/pkg/sqlc"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestSetupStaticServiceAccounts(t *testing.T) {
	tenantDomain := "example.com"

	t.Run("empty string", func(t *testing.T) {
		ctx := context.Background()
		database := db.NewMockDatabase(t)
		err := fixtures.SetupStaticServiceAccounts(ctx, database, "", tenantDomain)
		assert.EqualError(t, err, "EOF")
	})

	t.Run("user with no roles", func(t *testing.T) {
		ctx := context.Background()
		database := db.NewMockDatabase(t)
		json := `[
					{
						"name": "nais-service-account",
						"apiKey": "some key",
						"roles": []
					}
				]`
		err := fixtures.SetupStaticServiceAccounts(ctx, database, json, tenantDomain)
		assert.EqualError(t, err, "service account must have at least one role: 'nais-service-account'")
	})

	t.Run("missing API key", func(t *testing.T) {
		ctx := context.Background()
		database := db.NewMockDatabase(t)
		json := `[
				{
					"name": "nais-service-account",
					"roles": ["Admin"]
				}
			]`
		err := fixtures.SetupStaticServiceAccounts(ctx, database, json, tenantDomain)
		assert.EqualError(t, err, "service account is missing an API key: 'nais-service-account'")
	})

	t.Run("user with invalid name", func(t *testing.T) {
		ctx := context.Background()
		database := db.NewMockDatabase(t)
		json := `[
				{
					"name": "service-account",
					"apiKey": "some key",
					"roles": ["Team viewer", "invalid name"]
				}
			]`
		err := fixtures.SetupStaticServiceAccounts(ctx, database, json, tenantDomain)
		assert.EqualError(t, err, "service account is missing required 'nais-' prefix: 'service-account'")
	})

	t.Run("user with invalid role", func(t *testing.T) {
		ctx := context.Background()
		database := db.NewMockDatabase(t)
		json := `[
				{
					"name": "nais-service-account",
					"apiKey": "some key",
					"roles": ["role"]
				}
			]`
		err := fixtures.SetupStaticServiceAccounts(ctx, database, json, tenantDomain)
		assert.EqualError(t, err, "invalid role name: 'role' for service account 'nais-service-account'")
	})

	t.Run("create multiple service accounts", func(t *testing.T) {
		ctx := context.Background()
		txCtx := context.Background()
		database := db.NewMockDatabase(t)
		dbtx := db.NewMockDatabase(t)

		sa1 := userWithName("nais-service-account-1", tenantDomain)
		sa2 := userWithName("nais-service-account-2", tenantDomain)

		database.
			On("Transaction", ctx, mock.AnythingOfType("db.TransactionFunc")).
			Run(func(args mock.Arguments) {
				fn := args.Get(1).(db.TransactionFunc)
				fn(txCtx, dbtx)
			}).
			Return(nil).
			Once()

		// First service account
		dbtx.
			On("GetUserByEmail", txCtx, "nais-service-account-1@example.com.serviceaccounts.nais.io").
			Return(nil, errors.New("user not found")).
			Once()
		dbtx.
			On("AddUser", txCtx, "nais-service-account-1", "nais-service-account-1@example.com.serviceaccounts.nais.io").
			Return(sa1, nil).
			Once()
		dbtx.
			On("RemoveAllUserRoles", txCtx, sa1.ID).
			Return(nil).
			Once()
		dbtx.
			On("RemoveApiKeysFromUser", txCtx, sa1.ID).
			Return(nil).
			Once()
		dbtx.
			On("AssignGlobalRoleToUser", txCtx, sa1.ID, sqlc.RoleNameTeamcreator).
			Return(nil).
			Once()
		dbtx.
			On("AssignGlobalRoleToUser", txCtx, sa1.ID, sqlc.RoleNameTeamviewer).
			Return(nil).
			Once()
		dbtx.
			On("CreateAPIKey", txCtx, "key-1", sa1.ID).
			Return(nil).
			Once()

		// Second service account
		dbtx.
			On("GetUserByEmail", txCtx, "nais-service-account-2@example.com.serviceaccounts.nais.io").
			Return(sa2, nil).
			Once()
		dbtx.
			On("RemoveAllUserRoles", txCtx, sa2.ID).
			Return(nil).
			Once()
		dbtx.
			On("RemoveApiKeysFromUser", txCtx, sa2.ID).
			Return(nil).
			Once()
		dbtx.
			On("AssignGlobalRoleToUser", txCtx, sa2.ID, sqlc.RoleNameAdmin).
			Return(nil).
			Once()
		dbtx.
			On("CreateAPIKey", txCtx, "key-2", sa2.ID).
			Return(nil).
			Once()

		json := `[
				{
					"name": "nais-service-account-1",
					"apiKey": "key-1",
					"roles": ["Team creator", "Team viewer"]
				},
				{
					"name": "nais-service-account-2",
					"apiKey": "key-2",
					"roles": ["Admin"]
				}
			]`
		err := fixtures.SetupStaticServiceAccounts(ctx, database, json, tenantDomain)
		assert.NoError(t, err)
	})
}

func userWithName(name, tenantDomain string) *db.User {
	return &db.User{
		User: &sqlc.User{
			ID:    uuid.New(),
			Email: name + "@" + tenantDomain + ".serviceaccounts.nais.io",
			Name:  name,
		},
	}
}
