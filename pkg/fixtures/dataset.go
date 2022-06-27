package fixtures

import (
	"github.com/nais/console/pkg/authz"
	helpers "github.com/nais/console/pkg/console"
	"github.com/nais/console/pkg/dbmodels"
	log "github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

var (
	allAccessLevels = string(authz.AccessLevelCreate) + string(authz.AccessLevelRead) + string(authz.AccessLevelUpdate) + string(authz.AccessLevelDelete)
)

const (
	AdminUserName = "admin"
	defaultApiKey = "secret" // FIXME: Get from env var
)

// InsertInitialDataset Insert an initial dataset into the database. This will only be executed if there are currently
// no users in the users table.
func InsertInitialDataset(db *gorm.DB, partnerDomain string) error {
	return db.Transaction(func(tx *gorm.DB) error {
		// If there are any users in the database, skip creation
		users := make([]*dbmodels.User, 0)
		var numUsers int64
		tx.Find(&users).Count(&numUsers)
		if numUsers > 0 {
			log.Infof("users table not empty, skipping inserts of the initial dataset.")
			return nil
		}

		log.Infof("Inserting initial root user into database.")
		rootUser := &dbmodels.User{
			Name:  AdminUserName,
			Email: helpers.ServiceAccountEmail(AdminUserName, partnerDomain),
		}

		err := tx.Create(rootUser).Error
		if err != nil {
			return err
		}

		apiKey := &dbmodels.ApiKey{
			APIKey: defaultApiKey,
			User:   *rootUser,
		}

		return tx.Create(apiKey).Error
	})
}
