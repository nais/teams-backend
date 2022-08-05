package fixtures

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/nais/console/pkg/console"
	"github.com/nais/console/pkg/dbmodels"
	"gorm.io/gorm"
)

type ServiceAccount struct {
	Name   string   `json:"name"`
	Roles  []string `json:"roles"`
	APIKey string   `json:"apiKey"`
}

const ServiceAccountPrefix = "nais-"

// SetupStaticServiceAccounts Create a set of service accounts with roles and API keys
func SetupStaticServiceAccounts(db *gorm.DB, serviceAccountsRaw, tenantDomain string) error {
	return db.Transaction(func(tx *gorm.DB) error {
		serviceAccounts := make([]ServiceAccount, 0)
		err := json.NewDecoder(strings.NewReader(serviceAccountsRaw)).Decode(&serviceAccounts)
		if err != nil {
			return err
		}

		for _, serviceAccount := range serviceAccounts {
			if !strings.HasPrefix(serviceAccount.Name, ServiceAccountPrefix) {
				return fmt.Errorf("service account is missing required '%s' prefix: '%s'", ServiceAccountPrefix, serviceAccount.Name)
			}

			if len(serviceAccount.Roles) == 0 {
				return fmt.Errorf("service account must have at least one role: '%s'", serviceAccount.Name)
			}

			if serviceAccount.APIKey == "" {
				return fmt.Errorf("service account is missing an API key: '%s'", serviceAccount.Name)
			}

			roles := make([]*dbmodels.Role, 0)
			err = tx.Where("name in (?)", serviceAccount.Roles).Find(&roles).Error
			if err != nil {
				return err
			}

			if len(roles) != len(serviceAccount.Roles) {
				return fmt.Errorf("one or more roles could not be found: %s", serviceAccount.Roles)
			}

			user := &dbmodels.User{
				Name:  serviceAccount.Name,
				Email: console.ServiceAccountEmail(dbmodels.Slug(serviceAccount.Name), tenantDomain),
			}

			err = tx.Where("email = ?", user.Email).FirstOrCreate(user).Error
			if err != nil {
				return err
			}

			err = tx.Where("user_id = ?", user.ID).Delete(&dbmodels.UserRole{}).Error
			if err != nil {
				return err
			}
			err = tx.Where("user_id = ?", user.ID).Delete(&dbmodels.ApiKey{}).Error
			if err != nil {
				return err
			}

			userRoles := make([]*dbmodels.UserRole, len(roles))
			for idx, role := range roles {
				userRoles[idx] = &dbmodels.UserRole{
					RoleID: *role.ID,
					UserID: *user.ID,
				}
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
