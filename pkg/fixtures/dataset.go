package fixtures

import (
	"github.com/nais/console/pkg/dbmodels"
	"github.com/nais/console/pkg/roles"
	log "github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

const (
	AdminUserName        = "nais-console user"
	AdminUserEmailPrefix = "nais-console" // matches the default nais admin user account in the tenant GCP org
)

// InsertInitialDataset Insert an initial dataset into the database. This will only be executed if there are currently
// no users in the users table.
func InsertInitialDataset(db *gorm.DB, tenantDomain string, adminApiKey string) error {
	return db.Transaction(func(tx *gorm.DB) error {
		// If there are any users in the database, skip creation
		users := make([]*dbmodels.User, 0)
		var numUsers int64
		tx.Find(&users).Count(&numUsers)
		if numUsers > 0 {
			log.Infof("users table not empty, skipping inserts of the initial dataset.")
			return nil
		}

		log.Infof("Inserting initial admin user.")
		adminUser := &dbmodels.User{
			Name:  AdminUserName,
			Email: AdminUserEmailPrefix + "@" + tenantDomain,
		}

		err := tx.Create(adminUser).Error
		if err != nil {
			return err
		}

		if adminApiKey != "" {
			apiKey := &dbmodels.ApiKey{
				APIKey: adminApiKey,
				User:   *adminUser,
			}
			err = tx.Create(apiKey).Error
			if err != nil {
				return err
			}
		}

		err = CreateRolesAndAuthorizations(tx)
		if err != nil {
			return err
		}

		adminRole := &dbmodels.Role{}
		err = tx.Where("name = ?", roles.RoleAdmin).First(adminRole).Error
		if err != nil {
			return err
		}
		err = tx.Create(&dbmodels.UserRole{
			RoleID: *adminRole.ID,
			UserID: *adminUser.ID,
		}).Error
		if err != nil {
			return err
		}

		return nil
	})
}

func CreateRolesAndAuthorizations(tx *gorm.DB) error {
	auditLogsRead := &dbmodels.Authorization{Name: string(roles.AuthorizationAuditLogsRead)}
	serviceAccountsCreate := &dbmodels.Authorization{Name: string(roles.AuthorizationServiceAccountsCreate)}
	serviceAccountsDelete := &dbmodels.Authorization{Name: string(roles.AuthorizationServiceAccountsDelete)}
	serviceAccountsList := &dbmodels.Authorization{Name: string(roles.AuthorizationServiceAccountsList)}
	serviceAccountsRead := &dbmodels.Authorization{Name: string(roles.AuthorizationServiceAccountsRead)}
	serviceAccountsUpdate := &dbmodels.Authorization{Name: string(roles.AuthorizationServiceAccountsUpdate)}
	systemStatesDelete := &dbmodels.Authorization{Name: string(roles.AuthorizationSystemStatesDelete)}
	systemStatesRead := &dbmodels.Authorization{Name: string(roles.AuthorizationSystemStatesRead)}
	systemStatesUpdate := &dbmodels.Authorization{Name: string(roles.AuthorizationSystemStatesUpdate)}
	teamsCreate := &dbmodels.Authorization{Name: string(roles.AuthorizationTeamsCreate)}
	teamsDelete := &dbmodels.Authorization{Name: string(roles.AuthorizationTeamsDelete)}
	teamsList := &dbmodels.Authorization{Name: string(roles.AuthorizationTeamsList)}
	teamsRead := &dbmodels.Authorization{Name: string(roles.AuthorizationTeamsRead)}
	teamsUpdate := &dbmodels.Authorization{Name: string(roles.AuthorizationTeamsUpdate)}
	usersList := &dbmodels.Authorization{Name: string(roles.AuthorizationUsersList)}
	usersUpdate := &dbmodels.Authorization{Name: string(roles.AuthorizationUsersUpdate)}
	authorizations := []*dbmodels.Authorization{
		auditLogsRead,
		serviceAccountsCreate,
		serviceAccountsDelete,
		serviceAccountsList,
		serviceAccountsRead,
		serviceAccountsUpdate,
		systemStatesDelete,
		systemStatesRead,
		systemStatesUpdate,
		teamsCreate,
		teamsDelete,
		teamsList,
		teamsRead,
		teamsUpdate,
		usersList,
		usersUpdate,
	}
	err := tx.Create(authorizations).Error
	if err != nil {
		return err
	}

	roleAdmin := &dbmodels.Role{Name: string(roles.RoleAdmin)}
	serviceAccountCreator := &dbmodels.Role{Name: string(roles.RoleServiceAccountCreator)}
	serviceAccountOwner := &dbmodels.Role{Name: string(roles.RoleServiceAccountOwner)}
	teamCreator := &dbmodels.Role{Name: string(roles.RoleTeamCreator)}
	teamMember := &dbmodels.Role{Name: string(roles.RoleTeamMember)}
	teamOwner := &dbmodels.Role{Name: string(roles.RoleTeamOwner)}
	teamViewer := &dbmodels.Role{Name: string(roles.RoleTeamViewer)}
	userAdmin := &dbmodels.Role{Name: string(roles.RoleUserAdmin)}
	userViewer := &dbmodels.Role{Name: string(roles.RoleUserViewer)}
	err = tx.Create([]*dbmodels.Role{
		roleAdmin,
		serviceAccountCreator,
		serviceAccountOwner,
		teamCreator,
		teamMember,
		teamOwner,
		teamViewer,
		userAdmin,
		userViewer,
	}).Error
	if err != nil {
		return err
	}

	ra := []*dbmodels.RoleAuthorization{
		{Role: *roleAdmin, Authorization: *auditLogsRead},
		{Role: *roleAdmin, Authorization: *serviceAccountsCreate},
		{Role: *roleAdmin, Authorization: *serviceAccountsDelete},
		{Role: *roleAdmin, Authorization: *serviceAccountsList},
		{Role: *roleAdmin, Authorization: *serviceAccountsRead},
		{Role: *roleAdmin, Authorization: *serviceAccountsUpdate},
		{Role: *roleAdmin, Authorization: *systemStatesDelete},
		{Role: *roleAdmin, Authorization: *systemStatesRead},
		{Role: *roleAdmin, Authorization: *systemStatesUpdate},
		{Role: *roleAdmin, Authorization: *teamsCreate},
		{Role: *roleAdmin, Authorization: *teamsDelete},
		{Role: *roleAdmin, Authorization: *teamsList},
		{Role: *roleAdmin, Authorization: *teamsRead},
		{Role: *roleAdmin, Authorization: *teamsUpdate},
		{Role: *roleAdmin, Authorization: *usersList},
		{Role: *roleAdmin, Authorization: *usersUpdate},

		{Role: *serviceAccountCreator, Authorization: *serviceAccountsCreate},

		{Role: *serviceAccountOwner, Authorization: *serviceAccountsDelete},
		{Role: *serviceAccountOwner, Authorization: *serviceAccountsRead},
		{Role: *serviceAccountOwner, Authorization: *serviceAccountsUpdate},

		{Role: *teamCreator, Authorization: *teamsCreate},

		{Role: *teamMember, Authorization: *teamsRead},
		{Role: *teamMember, Authorization: *auditLogsRead},

		{Role: *teamOwner, Authorization: *teamsDelete},
		{Role: *teamOwner, Authorization: *teamsRead},
		{Role: *teamOwner, Authorization: *teamsUpdate},
		{Role: *teamOwner, Authorization: *auditLogsRead},

		{Role: *teamViewer, Authorization: *teamsList},
		{Role: *teamViewer, Authorization: *teamsRead},
		{Role: *teamViewer, Authorization: *auditLogsRead},

		{Role: *userAdmin, Authorization: *usersList},
		{Role: *userAdmin, Authorization: *usersUpdate},

		{Role: *userViewer, Authorization: *usersList},
	}
	err = tx.Create(ra).Error
	if err != nil {
		return err
	}

	return nil
}
