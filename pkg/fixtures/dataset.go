package fixtures

import (
	"context"
	"fmt"
	helpers "github.com/nais/console/pkg/console"
	console_reconciler "github.com/nais/console/pkg/reconcilers/console"

	"github.com/google/uuid"
	"github.com/nais/console/pkg/authz"
	"github.com/nais/console/pkg/dbmodels"
	"github.com/nais/console/pkg/graph"
	default_roles "github.com/nais/console/pkg/roles"
	log "github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

func serialuuid(serial byte) *uuid.UUID {
	return &uuid.UUID{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, serial}
}

func pk(serial byte) dbmodels.Model {
	return dbmodels.Model{
		ID: serialuuid(serial),
	}
}

const (
	idRootUser        = 0xa0
	idRootRoleBinding = 0xa2

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
			Model: pk(idRootUser),
			Email: helpers.Strp(emailRootUser),
			Name:  helpers.Strp(nameRootUser),
		}

		roles := []*dbmodels.Role{
			{
				Model:       dbmodels.Model{ID: default_roles.TeamManagerID},
				SystemID:    console.ID,
				Name:        "Team manager",
				Description: "Allows a user to update the contents of a team.",
				Resource:    string(graph.ResourceTeams),
				AccessLevel: authz.AccessReadWrite,
				Permission:  authz.PermissionAllow,
			},
			{
				Model:       dbmodels.Model{ID: default_roles.TeamReaderID},
				SystemID:    console.ID,
				Name:        "Team reader",
				Description: "Allows a user to read the contents of a team.",
				Resource:    string(graph.ResourceTeams),
				AccessLevel: authz.AccessRead,
				Permission:  authz.PermissionAllow,
			},
		}

		roleBinding := &dbmodels.RoleBinding{
			Model:  pk(idRootRoleBinding),
			UserID: serialuuid(idRootUser),
			RoleID: default_roles.TeamManagerID,
		}

		apikey := &dbmodels.ApiKey{
			APIKey: defaultApiKey,
			UserID: *rootUser.ID,
		}

		tx.Create(rootUser)
		tx.Create(roles)
		tx.Create(roleBinding)
		tx.Create(apikey)

		return tx.Error
	})
}
