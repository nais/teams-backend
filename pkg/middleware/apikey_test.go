package middleware

import (
	helpers "github.com/nais/console/pkg/console"
	"github.com/nais/console/pkg/test"
	"testing"

	"github.com/nais/console/pkg/dbmodels"
	"gorm.io/gorm"
)

func setupFixtures(db *gorm.DB) error {
	db.AutoMigrate(&dbmodels.User{})
	return db.Transaction(func(tx *gorm.DB) error {
		tx.Create(&dbmodels.User{
			Model:        dbmodels.Model{},
			SoftDeletes:  dbmodels.SoftDeletes{},
			Email:        helpers.Strp("user@example.com"),
			Name:         helpers.Strp("User Name"),
			Teams:        nil,
			RoleBindings: nil,
		})
		return nil
	})
}

func TestApiKeyAuthentication(t *testing.T) {
	db := test.GetTestDB()
	err := setupFixtures(db)
	if err != nil {
		panic(err)
	}

	// FIXME: Do some actual testing
	_ = db
}
