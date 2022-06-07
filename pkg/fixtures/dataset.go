package fixtures

import (
	"context"
	"fmt"
	helpers "github.com/nais/console/pkg/console"
	console_reconciler "github.com/nais/console/pkg/reconcilers/console"

	"github.com/google/uuid"
	"github.com/nais/console/pkg/authz"
	"github.com/nais/console/pkg/dbmodels"
	log "github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

func modelWithID(id *uuid.UUID) dbmodels.Model {
	return dbmodels.Model{
		ID: id,
	}
}

var (
	allAccessLevels = string(authz.AccessLevelCreate) + string(authz.AccessLevelRead) + string(authz.AccessLevelUpdate) + string(authz.AccessLevelDelete)
)

const (
	emailRootUser = "root@localhost"
	nameRootUser  = "admin"

	defaultApiKey = "secret"
)

func InsertRootUser(ctx context.Context, db *gorm.DB) error {
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
			Model: modelWithID(dbmodels.AdminUserID),
			Email: helpers.Strp(emailRootUser),
			Name:  helpers.Strp(nameRootUser),
		}

		roles := []*dbmodels.Role{
			{
				Model:       modelWithID(dbmodels.TeamEditorRoleID),
				SystemID:    console.ID,
				Name:        "Team editor",
				Description: "Gives the user full access to the team. If given on a global scale, this role gives full access to all teams.",
				Resource:    string(authz.ResourceTeams),
				AccessLevel: allAccessLevels,
				Permission:  authz.PermissionAllow,
			},
			{
				Model:       modelWithID(dbmodels.TeamViewerRoleID),
				SystemID:    console.ID,
				Name:        "Team viewer",
				Description: "Allows a user to view the contents of a team.",
				Resource:    string(authz.ResourceTeams),
				AccessLevel: string(authz.AccessLevelRead),
				Permission:  authz.PermissionAllow,
			},
			{
				Model:       modelWithID(dbmodels.TeamCreatorRoleID),
				SystemID:    console.ID,
				Name:        "Team creator",
				Description: "Allows a user to create new teams.",
				Resource:    string(authz.ResourceTeams),
				AccessLevel: string(authz.AccessLevelCreate),
				Permission:  authz.PermissionAllow,
			},

			{
				Model:       modelWithID(dbmodels.RoleEditorRoleID),
				SystemID:    console.ID,
				Name:        "Role editor",
				Description: "Gives the user role administration access.",
				Resource:    string(authz.ResourceRoles),
				AccessLevel: allAccessLevels,
				Permission:  authz.PermissionAllow,
			},
		}

		roleBindings := []*dbmodels.RoleBinding{
			{
				UserID: dbmodels.AdminUserID,
				RoleID: dbmodels.TeamEditorRoleID,
			},
			{
				UserID: dbmodels.AdminUserID,
				RoleID: dbmodels.RoleEditorRoleID,
			},
		}

		apikey := &dbmodels.ApiKey{
			APIKey: defaultApiKey,
			UserID: *rootUser.ID,
		}

		tx.Create(rootUser)
		tx.Create(roles)
		tx.Create(roleBindings)
		tx.Create(apikey)

		return tx.Error
	})
}
