package fixtures

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/nais/console/pkg/console"
	"github.com/nais/console/pkg/dbmodels"
	"gorm.io/gorm"
)

/*
ENV example

STATIC_SERVICE_ACCOUNTS="[
	{
		"name": "service-account-1",
		"apiKey": "key1",
		"roles": ["role1", "role2"]
	},
	{
		"name": "service-account-2",
		"apiKey": "key2",
		"roles": ["role2", "role3"]
	}
]
*/

type ServiceAccount struct {
	Name   string   `json:"name"`
	Roles  []string `json:"roles"`
	APIKey string   `json:"apiKey"`
}

func SetupStaticServiceAccounts(db *gorm.DB, serviceAccountsRaw, tenantDomain string) error {
	return db.Transaction(func(tx *gorm.DB) error {
		var serviceAccounts []ServiceAccount
		err := json.NewDecoder(strings.NewReader(serviceAccountsRaw)).Decode(&serviceAccounts)
		if err != nil {
			return err
		}

		for _, serviceAccount := range serviceAccounts {
			roles := []*dbmodels.Role{}
			err := tx.Where("name in (?)", serviceAccount.Roles).Find(&roles).Error
			if err != nil {
				return err
			}

			if len(roles) != len(serviceAccount.Roles) {
				return fmt.Errorf("could not find roles %v", serviceAccount.Roles)
			}

			user := &dbmodels.User{
				Name:  serviceAccount.Name,
				Email: console.ServiceAccountEmail(dbmodels.Slug(serviceAccount.Name), tenantDomain),
			}

			err = tx.Where("email = ?", user.Email).FirstOrCreate(user).Error
			if err != nil {
				return err
			}

			// First clean up from previous runs
			err = tx.Where("user_id = ?", user.ID).Delete(&dbmodels.UserRole{}).Error
			if err != nil {
				return err
			}
			err = tx.Where("user_id = ?", user.ID).Delete(&dbmodels.ApiKey{}).Error
			if err != nil {
				return err
			}

			userRoles := []*dbmodels.UserRole{}
			for _, role := range roles {
				userRoles = append(userRoles, &dbmodels.UserRole{
					RoleID: *role.ID,
					UserID: *user.ID,
				})
			}

			err = tx.Create(userRoles).Error
			if err != nil {
				return err
			}

			err = tx.Create(&dbmodels.ApiKey{
				UserID: *user.ID,
				APIKey: serviceAccount.APIKey,
			}).Error
			if err != nil {
				return err
			}
		}

		return nil
	})
}
