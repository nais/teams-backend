package fixtures_test

import (
	"github.com/nais/console/pkg/dbmodels"
	"github.com/nais/console/pkg/fixtures"
	"github.com/nais/console/pkg/test"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
	"testing"
)

func TestSetupStaticServiceAccounts(t *testing.T) {
	db := test.GetTestDB()
	db.AutoMigrate(&dbmodels.ApiKey{}, &dbmodels.Role{}, &dbmodels.UserRole{})
	fixtures.CreateRolesAndAuthorizations(db)

	t.Run("empty string", func(t *testing.T) {
		err := fixtures.SetupStaticServiceAccounts(db, "", "example.com")
		assert.EqualError(t, err, "EOF")
	})

	t.Run("user with no roles", func(t *testing.T) {
		env := `[
			{
				"name": "nais-service-account",
				"apiKey": "some key",
				"roles": []
			}	
		]`
		err := fixtures.SetupStaticServiceAccounts(db, env, "example.com")
		assert.EqualError(t, err, "service account must have at least one role: 'nais-service-account'")
	})

	t.Run("missing API key", func(t *testing.T) {
		env := `[
			{
				"name": "nais-service-account",
				"roles": ["Admin"]
			}	
		]`
		err := fixtures.SetupStaticServiceAccounts(db, env, "example.com")
		assert.EqualError(t, err, "service account is missing an API key: 'nais-service-account'")
	})

	t.Run("user with invalid name", func(t *testing.T) {
		env := `[
			{
				"name": "service-account",
				"apiKey": "some key",
				"roles": ["Team viewer", "invalid name"]
			}	
		]`
		err := fixtures.SetupStaticServiceAccounts(db, env, "example.com")
		assert.EqualError(t, err, "service account is missing required 'nais-' prefix: 'service-account'")
	})

	t.Run("user with invalid role", func(t *testing.T) {
		env := `[
			{
				"name": "nais-service-account",
				"apiKey": "some key",
				"roles": ["Team viewer", "invalid name"]
			}	
		]`
		err := fixtures.SetupStaticServiceAccounts(db, env, "example.com")
		assert.EqualError(t, err, "one or more roles could not be found: [Team viewer invalid name]")
	})

	t.Run("create multiple service accounts", func(t *testing.T) {
		env := `[
			{
				"name": "nais-service-account-1",
				"apiKey": "key-1",
				"roles": ["Team creator", "Team viewer"]
			},
			{
				"name": "nais-service-account-2",
				"apiKey": "key-2",
				"roles": ["User viewer"]
			},
			{
				"name": "nais-service-account-3",
				"apiKey": "key-3",
				"roles": ["Admin"]
			}	
		]`
		err := fixtures.SetupStaticServiceAccounts(db, env, "example.com")
		assert.NoError(t, err)

		users := make([]*dbmodels.User, 0)
		db.Preload("RoleBindings").Preload("RoleBindings.Role", func(db *gorm.DB) *gorm.DB {
			return db.Order("name ASC")
		}).Order("name ASC").Find(&users)
		assert.Len(t, users, 3)

		apiKey := &dbmodels.ApiKey{}
		assert.Len(t, users[0].RoleBindings, 2)
		assert.Equal(t, "Team creator", users[0].RoleBindings[0].Role.Name)
		assert.Equal(t, "Team viewer", users[0].RoleBindings[1].Role.Name)
		db.Where("user_id = ?", users[0].ID).First(apiKey)
		assert.Equal(t, "key-1", apiKey.APIKey)

		apiKey = &dbmodels.ApiKey{}
		assert.Len(t, users[1].RoleBindings, 1)
		assert.Equal(t, "User viewer", users[1].RoleBindings[0].Role.Name)
		db.Where("user_id = ?", users[1].ID).First(apiKey)
		assert.Equal(t, "key-2", apiKey.APIKey)

		apiKey = &dbmodels.ApiKey{}
		assert.Len(t, users[2].RoleBindings, 1)
		assert.Equal(t, "Admin", users[2].RoleBindings[0].Role.Name)
		db.Where("user_id = ?", users[2].ID).First(apiKey)
		assert.Equal(t, "key-3", apiKey.APIKey)
	})
}
