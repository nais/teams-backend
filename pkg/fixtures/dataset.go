package fixtures

import (
	"context"
	"fmt"
	helpers "github.com/nais/console/pkg/console"
	console_reconciler "github.com/nais/console/pkg/reconcilers/console"
	role_names "github.com/nais/console/pkg/roles"

	"github.com/nais/console/pkg/authz"
	"github.com/nais/console/pkg/dbmodels"
	log "github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

var (
	allAccessLevels = string(authz.AccessLevelCreate) + string(authz.AccessLevelRead) + string(authz.AccessLevelUpdate) + string(authz.AccessLevelDelete)
)

const (
	emailRootUser = "root@localhost"
	nameRootUser  = "admin"

	defaultApiKey = "secret"
)

func InsertInitialDataset(ctx context.Context, db *gorm.DB) error {
	return db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// If there are any users in the database, skip creation
		users := make([]*dbmodels.User, 0)
		var numUsers int64
		tx.Find(&users).Count(&numUsers)
		if numUsers > 0 {
			return nil
		}

		console := &dbmodels.System{}
		tx.First(console, "name = ?", console_reconciler.Name)
		if console.ID == nil {
			return fmt.Errorf("system fixtures not in database")
		}

		log.Infof("Inserting initial root user into database...")

		rootUser := &dbmodels.User{
			Email: helpers.Strp(emailRootUser),
			Name:  helpers.Strp(nameRootUser),
		}

		tx.Create(rootUser)

		apikey := &dbmodels.ApiKey{
			APIKey: defaultApiKey,
			UserID: *rootUser.ID,
		}

		tx.Create(apikey)

		roles := map[string]*dbmodels.Role{
			role_names.TeamEditor: {
				SystemID:    console.ID,
				Name:        role_names.TeamEditor,
				Description: "Gives the user full access to the team. If given on a global scale, this role gives full access to all teams.",
				Resource:    string(authz.ResourceTeams),
				AccessLevel: allAccessLevels,
				Permission:  authz.PermissionAllow,
			},
			role_names.TeamViewer: {
				SystemID:    console.ID,
				Name:        role_names.TeamViewer,
				Description: "Allows a user to view the contents of a team.",
				Resource:    string(authz.ResourceTeams),
				AccessLevel: string(authz.AccessLevelRead),
				Permission:  authz.PermissionAllow,
			},
			role_names.TeamCreator: {
				SystemID:    console.ID,
				Name:        role_names.TeamCreator,
				Description: "Allows a user to create new teams.",
				Resource:    string(authz.ResourceTeams),
				AccessLevel: string(authz.AccessLevelCreate),
				Permission:  authz.PermissionAllow,
			},

			role_names.RoleEditor: {
				SystemID:    console.ID,
				Name:        role_names.RoleEditor,
				Description: "Gives the user role administration access.",
				Resource:    string(authz.ResourceRoles),
				AccessLevel: allAccessLevels,
				Permission:  authz.PermissionAllow,
			},
		}

		for _, role := range roles {
			tx.Create(role)
		}

		roleBindings := []*dbmodels.RoleBinding{
			{
				UserID: rootUser.ID,
				RoleID: roles[role_names.TeamEditor].ID,
			},
			{
				UserID: rootUser.ID,
				RoleID: roles[role_names.RoleEditor].ID,
			},
		}

		tx.Create(roleBindings)

		return tx.Error
	})
}
