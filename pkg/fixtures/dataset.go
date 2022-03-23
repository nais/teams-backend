package fixtures

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/nais/console/pkg/auth"
	"github.com/nais/console/pkg/dbmodels"
	"github.com/nais/console/pkg/graph"
	console_reconciler "github.com/nais/console/pkg/reconcilers/console"
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

func strp(s string) *string {
	return &s
}

const (
	idRootUser = 0xa0
	idRootRole = 0xa1

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
			Email: strp("root@localhost"),
			Name:  strp("the administrator"),
		}

		role := &dbmodels.Role{
			Model:       pk(idRootRole),
			SystemID:    console.ID,
			Resource:    string(graph.ResourceTeams),
			AccessLevel: auth.AccessReadWrite,
			Permission:  auth.PermissionAllow,
			UserID:      serialuuid(idRootUser),
		}

		apikey := &dbmodels.ApiKey{
			APIKey: defaultApiKey,
			UserID: *rootUser.ID,
		}

		tx.Save(rootUser)
		tx.Save(role)
		tx.Save(apikey)

		return tx.Error
	})
}
