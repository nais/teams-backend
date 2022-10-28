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
	t.Run("empty string", func(t *testing.T) {
		ctx := context.Background()
		database := db.NewMockDatabase(t)
		err := fixtures.SetupStaticServiceAccounts(ctx, database, "")
		assert.EqualError(t, err, "EOF")
	})

	t.Run("service account with no roles", func(t *testing.T) {
		ctx := context.Background()
		database := db.NewMockDatabase(t)
		json := `[
					{
						"name": "nais-service-account",
						"apiKey": "some key",
						"roles": []
					}
				]`
		err := fixtures.SetupStaticServiceAccounts(ctx, database, json)
		assert.EqualError(t, err, `service account must have at least one role: "nais-service-account"`)
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
		err := fixtures.SetupStaticServiceAccounts(ctx, database, json)
		assert.EqualError(t, err, `service account is missing an API key: "nais-service-account"`)
	})

	t.Run("service account with invalid name", func(t *testing.T) {
		ctx := context.Background()
		database := db.NewMockDatabase(t)
		json := `[
				{
					"name": "service-account",
					"apiKey": "some key",
					"roles": ["Team viewer"]
				}
			]`
		err := fixtures.SetupStaticServiceAccounts(ctx, database, json)
		assert.EqualError(t, err, `service account is missing required "nais-" prefix: "service-account"`)
	})

	t.Run("service account with invalid role", func(t *testing.T) {
		ctx := context.Background()
		database := db.NewMockDatabase(t)
		json := `[
				{
					"name": "nais-service-account",
					"apiKey": "some key",
					"roles": ["role"]
				}
			]`
		err := fixtures.SetupStaticServiceAccounts(ctx, database, json)
		assert.EqualError(t, err, `invalid role name: "role" for service account "nais-service-account"`)
	})

	t.Run("create multiple service accounts and delete old one", func(t *testing.T) {
		ctx := context.Background()
		txCtx := context.Background()
		database := db.NewMockDatabase(t)
		dbtx := db.NewMockDatabase(t)

		sa1 := serviceAccountWithName("nais-service-account-1")
		sa2 := serviceAccountWithName("nais-service-account-2")
		sa3 := serviceAccountWithName("nais-service-account-3")

		database.
			On("Transaction", ctx, mock.AnythingOfType("db.DatabaseTransactionFunc")).
			Run(func(args mock.Arguments) {
				fn := args.Get(1).(db.DatabaseTransactionFunc)
				fn(txCtx, dbtx)
			}).
			Return(nil).
			Once()

		// First service account
		dbtx.
			On("GetServiceAccountByName", txCtx, "nais-service-account-1").
			Return(nil, errors.New("service account not found")).
			Once()
		dbtx.
			On("CreateServiceAccount", txCtx, "nais-service-account-1").
			Return(sa1, nil).
			Once()
		dbtx.
			On("RemoveAllServiceAccountRoles", txCtx, sa1.ID).
			Return(nil).
			Once()
		dbtx.
			On("RemoveApiKeysFromServiceAccount", txCtx, sa1.ID).
			Return(nil).
			Once()
		dbtx.
			On("AssignGlobalRoleToServiceAccount", txCtx, sa1.ID, sqlc.RoleNameTeamcreator).
			Return(nil).
			Once()
		dbtx.
			On("AssignGlobalRoleToServiceAccount", txCtx, sa1.ID, sqlc.RoleNameTeamviewer).
			Return(nil).
			Once()
		dbtx.
			On("CreateAPIKey", txCtx, "key-1", sa1.ID).
			Return(nil).
			Once()

		// Second service account
		dbtx.
			On("GetServiceAccountByName", txCtx, "nais-service-account-2").
			Return(sa2, nil).
			Once()
		dbtx.
			On("RemoveAllServiceAccountRoles", txCtx, sa2.ID).
			Return(nil).
			Once()
		dbtx.
			On("RemoveApiKeysFromServiceAccount", txCtx, sa2.ID).
			Return(nil).
			Once()
		dbtx.
			On("AssignGlobalRoleToServiceAccount", txCtx, sa2.ID, sqlc.RoleNameAdmin).
			Return(nil).
			Once()
		dbtx.
			On("CreateAPIKey", txCtx, "key-2", sa2.ID).
			Return(nil).
			Once()

		// Delete old service account
		dbtx.
			On("GetServiceAccounts", txCtx).
			Return([]*db.ServiceAccount{sa1, sa2, sa3}, nil).
			Once()
		dbtx.
			On("DeleteServiceAccount", txCtx, sa3.ID).
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
		err := fixtures.SetupStaticServiceAccounts(ctx, database, json)
		assert.NoError(t, err)
	})
}

func serviceAccountWithName(name string) *db.ServiceAccount {
	return &db.ServiceAccount{
		ServiceAccount: &sqlc.ServiceAccount{
			ID:   uuid.New(),
			Name: name,
		},
	}
}
