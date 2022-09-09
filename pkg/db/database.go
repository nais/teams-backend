package db

import (
	"context"
	"database/sql"
	"errors"

	"github.com/golang-migrate/migrate/v4"

	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	"github.com/google/uuid"
	"github.com/nais/console/pkg/slug"
	"github.com/nais/console/pkg/sqlc"
	"github.com/nais/console/sqlc/schemas"
)

type database struct {
	querier Querier
}

type TransactionFunc func(ctx context.Context, dbtx Database) error

type Database interface {
	AddAuditLog(ctx context.Context, correlationID uuid.UUID, systemName sqlc.SystemName, actorEmail *string, targetTeamSlug *slug.Slug, targetUserEmail *string, action sqlc.AuditAction, message string) error
	AddUser(ctx context.Context, name, email string) (*User, error)
	AddServiceAccount(ctx context.Context, name string) (*ServiceAccount, error)
	GetServiceAccount(ctx context.Context, name string) (*ServiceAccount, error)
	GetUserByID(ctx context.Context, ID uuid.UUID) (*User, error)
	GetUserByEmail(ctx context.Context, email string) (*User, error)
	GetUserByApiKey(ctx context.Context, APIKey string) (*User, error)
	DeleteUser(ctx context.Context, userID uuid.UUID) error

	GetUsersByEmail(ctx context.Context, email string) ([]*User, error)
	GetUsers(ctx context.Context) ([]*User, error)
	GetUserTeams(ctx context.Context, userID uuid.UUID) ([]*Team, error)

	AddTeam(ctx context.Context, name string, slug slug.Slug, purpose *string, userID uuid.UUID) (*Team, error)
	SetTeamMetadata(ctx context.Context, teamID uuid.UUID, metadata TeamMetadata) error
	GetTeamMetadata(ctx context.Context, teamID uuid.UUID) (TeamMetadata, error)
	UpdateTeam(ctx context.Context, teamID uuid.UUID, name, purpose *string) (*Team, error)
	GetTeamByID(ctx context.Context, ID uuid.UUID) (*Team, error)
	GetTeamBySlug(ctx context.Context, slug slug.Slug) (*Team, error)
	GetTeams(ctx context.Context) ([]*Team, error)
	GetTeamMembers(ctx context.Context, teamID uuid.UUID) ([]*User, error)
	UserIsTeamOwner(ctx context.Context, userID, teamID uuid.UUID) (bool, error)
	SetTeamMemberRole(ctx context.Context, userID uuid.UUID, teamID uuid.UUID, role sqlc.RoleName) error

	GetAuditLogsForTeam(ctx context.Context, slug slug.Slug) ([]*AuditLog, error)

	AssignGlobalRoleToUser(ctx context.Context, userID uuid.UUID, roleName sqlc.RoleName) error
	RevokeGlobalRoleFromUser(ctx context.Context, userID uuid.UUID, roleName sqlc.RoleName) error
	AssignTargetedRoleToUser(ctx context.Context, userID uuid.UUID, roleName sqlc.RoleName, targetID uuid.UUID) error

	RemoveUserFromTeam(ctx context.Context, userID uuid.UUID, teamID uuid.UUID) error

	CreateAPIKey(ctx context.Context, apiKey string, userID uuid.UUID) error

	RemoveAllUserRoles(ctx context.Context, userID uuid.UUID) error
	RemoveApiKeysFromUser(ctx context.Context, userID uuid.UUID) error

	GetUserRoles(ctx context.Context, userID uuid.UUID) ([]*Role, error)

	Transaction(ctx context.Context, fn TransactionFunc) error

	LoadSystemState(ctx context.Context, systemName sqlc.SystemName, teamID uuid.UUID, state interface{}) error
	SetSystemState(ctx context.Context, systemName sqlc.SystemName, teamID uuid.UUID, state interface{}) error

	SetUserName(ctx context.Context, userID uuid.UUID, name string) (*User, error)

	SetTeamReconcileErrorForSystem(ctx context.Context, correlationID uuid.UUID, teamID uuid.UUID, systemName sqlc.SystemName, err error) error
	GetTeamReconcileErrors(ctx context.Context, teamID uuid.UUID) ([]*ReconcileError, error)
	ClearTeamReconcileErrorForSystem(ctx context.Context, teamID uuid.UUID, systemName sqlc.SystemName) error
}

func NewDatabase(q Querier) Database {
	return &database{querier: q}
}

func NullStringToStringP(ns sql.NullString) *string {
	var strP *string
	if ns.String != "" {
		strP = &ns.String
	}
	return strP
}

func Migrate(connString string) error {
	d, err := iofs.New(schemas.FS, ".")
	if err != nil {
		return err
	}
	defer d.Close()

	m, err := migrate.NewWithSourceInstance("iofs", d, connString)
	if err != nil {
		return err
	}

	err = m.Migrate(6)
	if err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return err
	}

	return nil
}

func nullString(s *string) sql.NullString {
	if s == nil {
		return sql.NullString{}
	}
	return sql.NullString{
		String: *s,
		Valid:  true,
	}
}

func nullUUID(ID *uuid.UUID) uuid.NullUUID {
	if ID == nil {
		return uuid.NullUUID{}
	}
	return uuid.NullUUID{
		UUID:  *ID,
		Valid: true,
	}
}
