package fixtures

import (
	"fmt"
	"github.com/nais/console/pkg/authz"
	helpers "github.com/nais/console/pkg/console"
	"github.com/nais/console/pkg/dbmodels"
	console_reconciler "github.com/nais/console/pkg/reconcilers/console"
	log "github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

var (
	allAccessLevels = string(authz.AccessLevelCreate) + string(authz.AccessLevelRead) + string(authz.AccessLevelUpdate) + string(authz.AccessLevelDelete)
)

const (
	AdminUserName = "admin"
	defaultApiKey = "secret"
)

func InsertInitialDataset(db *gorm.DB, partnerDomain string) error {
	return db.Transaction(func(tx *gorm.DB) error {
		// If there are any users in the database, skip creation
		users := make([]*dbmodels.User, 0)
		var numUsers int64
		tx.Find(&users).Count(&numUsers)
		if numUsers > 0 {
			return nil
		}

		console := &dbmodels.System{}
		err := tx.First(console, "name = ?", console_reconciler.Name).Error
		if err != nil {
			return fmt.Errorf("system fixtures not in database")
		}

		log.Infof("Inserting initial root user into database...")

		rootUser := &dbmodels.User{
			Name:  AdminUserName,
			Email: helpers.ServiceAccountEmail(AdminUserName, partnerDomain),
		}

		err = tx.Create(rootUser).Error
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
