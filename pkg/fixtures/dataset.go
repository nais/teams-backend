package fixtures

import (
	"github.com/nais/console/pkg/authz"
	helpers "github.com/nais/console/pkg/console"
	"github.com/nais/console/pkg/dbmodels"
	"github.com/nais/console/pkg/roles"
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
func InsertInitialDataset(db *gorm.DB, tenantDomain string) error {
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
			Email: helpers.ServiceAccountEmail(AdminUserName, tenantDomain),
		}

		err := tx.Create(rootUser).Error
		if err != nil {
			return err
		}

		apiKey := &dbmodels.ApiKey{
			APIKey: defaultApiKey,
			User:   *rootUser,
		}
		err = tx.Create(apiKey).Error
		if err != nil {
			return err
		}

		err = createRolesAndAuthorizations(tx)
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
			UserID: *rootUser.ID,
		}).Error
		if err != nil {
			return err
		}

		return nil
	})
}

func createRolesAndAuthorizations(tx *gorm.DB) error {
	auditLogsRead := &dbmodels.Authorization{Name: string(roles.AuthorizationAuditLogsRead)}
	serviceAccountsCreate := &dbmodels.Authorization{Name: string(roles.AuthorizationServiceAccountsCreate)}
	serviceAccountsDelete := &dbmodels.Authorization{Name: string(roles.AuthorizationServiceAccountsDelete)}
	serviceAccountList := &dbmodels.Authorization{Name: string(roles.AuthorizationServiceAccountList)}
	serviceAccountsUpdate := &dbmodels.Authorization{Name: string(roles.AuthorizationServiceAccountsUpdate)}
	systemStatesDelete := &dbmodels.Authorization{Name: string(roles.AuthorizationSystemStatesDelete)}
	systemStatesRead := &dbmodels.Authorization{Name: string(roles.AuthorizationSystemStatesRead)}
	systemStatesUpdate := &dbmodels.Authorization{Name: string(roles.AuthorizationSystemStatesUpdate)}
	teamsCreate := &dbmodels.Authorization{Name: string(roles.AuthorizationTeamsCreate)}
	teamsDelete := &dbmodels.Authorization{Name: string(roles.AuthorizationTeamsDelete)}
	teamsList := &dbmodels.Authorization{Name: string(roles.AuthorizationTeamsList)}
	teamsRead := &dbmodels.Authorization{Name: string(roles.AuthorizationTeamsRead)}
	teamsUpdate := &dbmodels.Authorization{Name: string(roles.AuthorizationTeamsUpdate)}
	authorizations := []*dbmodels.Authorization{
		auditLogsRead,
		serviceAccountsCreate,
		serviceAccountsDelete,
		serviceAccountList,
		serviceAccountsUpdate,
		systemStatesDelete,
		systemStatesRead,
		systemStatesUpdate,
		teamsCreate,
		teamsDelete,
		teamsList,
		teamsRead,
		teamsUpdate,
	}
	err := tx.Create(authorizations).Error
	if err != nil {
		return err
	}

	roleAdmin := &dbmodels.Role{Name: string(roles.RoleAdmin)}
	serviceAccountCreator := &dbmodels.Role{Name: string(roles.RoleServiceAccountCreaetor)}
	serviceAccountOwner := &dbmodels.Role{Name: string(roles.RoleServiceAccountOwner)}
	teamCreator := &dbmodels.Role{Name: string(roles.RoleTeamCreator)}
	teamMember := &dbmodels.Role{Name: string(roles.RoleTeamMember)}
	teamOwner := &dbmodels.Role{Name: string(roles.RoleTeamOwner)}
	teamViewer := &dbmodels.Role{Name: string(roles.RoleTeamViewer)}

	err = tx.Create([]*dbmodels.Role{roleAdmin, serviceAccountCreator, serviceAccountOwner, teamCreator, teamMember, teamOwner, teamViewer}).Error
	if err != nil {
		return err
	}

	ra := []*dbmodels.RoleAuthorization{
		{Role: *roleAdmin, Authorization: *auditLogsRead},
		{Role: *roleAdmin, Authorization: *serviceAccountsCreate},
		{Role: *roleAdmin, Authorization: *serviceAccountsDelete},
		{Role: *roleAdmin, Authorization: *serviceAccountList},
		{Role: *roleAdmin, Authorization: *serviceAccountsUpdate},
		{Role: *roleAdmin, Authorization: *systemStatesDelete},
		{Role: *roleAdmin, Authorization: *systemStatesRead},
		{Role: *roleAdmin, Authorization: *systemStatesUpdate},
		{Role: *roleAdmin, Authorization: *teamsCreate},
		{Role: *roleAdmin, Authorization: *teamsDelete},
		{Role: *roleAdmin, Authorization: *teamsList},
		{Role: *roleAdmin, Authorization: *teamsRead},
		{Role: *roleAdmin, Authorization: *teamsUpdate},

		{Role: *serviceAccountCreator, Authorization: *serviceAccountsCreate},

		{Role: *serviceAccountOwner, Authorization: *serviceAccountsDelete},
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
	}
	err = tx.Create(ra).Error
	if err != nil {
		return err
	}

	return nil
}
