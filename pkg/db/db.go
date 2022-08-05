package db

import (
	"context"
	"errors"
	"fmt"
	"github.com/google/uuid"
	"github.com/jackc/pgconn"
	"github.com/nais/console/pkg/authz"
	"github.com/nais/console/pkg/dbmodels"
	"github.com/nais/console/pkg/roles"
	"gorm.io/gorm"
	"strings"
)

const (
	DuplicateErrorCode = "23505"
)

// Model Enables abstracted access to CreatedBy and UpdatedBy for generic database models.
type Model interface {
	GetModel() *dbmodels.Model
}

func CreateTrackedObject(ctx context.Context, db *gorm.DB, newObject Model) error {
	user := authz.UserFromContext(ctx)
	if user == nil {
		return fmt.Errorf("context has no user")
	}

	model := newObject.GetModel()
	model.CreatedBy = user
	model.UpdatedBy = user
	return db.Create(newObject).Error
}

func CreateTrackedObjectIgnoringDuplicates(ctx context.Context, db *gorm.DB, obj Model) error {
	err := CreateTrackedObject(ctx, db, obj)

	if err == nil {
		return nil
	}

	switch t := err.(type) {
	case *pgconn.PgError:
		if t.Code == DuplicateErrorCode {
			return nil
		}
	}

	return err
}

func UpdateTrackedObject(ctx context.Context, db *gorm.DB, updatedObject Model) error {
	user := authz.UserFromContext(ctx)
	if user == nil {
		return fmt.Errorf("context has no user")
	}

	model := updatedObject.GetModel()
	model.UpdatedBy = user
	return db.Updates(updatedObject).Error
}

func UserIsTeamOwner(db *gorm.DB, userId, teamId uuid.UUID) (bool, error) {
	teamOwnerRole := &dbmodels.Role{}
	err := db.Where("name = ?", roles.RoleTeamOwner).First(teamOwnerRole).Error
	if err != nil {
		return false, err
	}

	roleBinding := &dbmodels.UserRole{}
	err = db.Where("role_id = ? AND user_id = ? AND target_id = ?", teamOwnerRole.ID, userId, teamId).First(roleBinding).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return false, nil
		}

		return false, err
	}

	return true, nil
}

func GetUserByEmail(db *gorm.DB, email string) *dbmodels.User {
	user := &dbmodels.User{}
	err := db.Where("email = ?", strings.ToLower(email)).First(user).Error
	if err != nil {
		return nil
	}
	return user
}
