package db

import (
	"context"
	"database/sql"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v4"
	"github.com/nais/console/pkg/sqlc"
)

type database struct {
	querier Querier
	conn    *pgx.Conn
}

type TransactionFunc func(ctx context.Context, dbtx Database) error

type Database interface {
	AddAuditLog(ctx context.Context, correlationID uuid.UUID, actorEmail *string, systemName *sqlc.SystemName, targetTeamSlug, targetUserEmail *string, action sqlc.AuditAction, message string) error
	AddUser(ctx context.Context, name, email string) (*User, error)
	GetUserByID(ctx context.Context, ID uuid.UUID) (*User, error)
	GetUserByEmail(ctx context.Context, email string) (*User, error)
	GetUserByApiKey(ctx context.Context, APIKey string) (*User, error)
	DeleteUser(ctx context.Context, userID uuid.UUID) error

	GetUsersByEmail(ctx context.Context, email string) ([]*User, error)

	AddTeam(ctx context.Context, name, slug string, purpose *string) (*Team, error)
	GetTeamByID(ctx context.Context, ID uuid.UUID) (*Team, error)
	GetTeamBySlug(ctx context.Context, slug string) (*Team, error)
	GetTeams(ctx context.Context) ([]*Team, error)

	AssignGlobalRoleToUser(ctx context.Context, userID uuid.UUID, roleName sqlc.RoleName) error
	AssignTargetedRoleToUser(ctx context.Context, userID uuid.UUID, roleName sqlc.RoleName, targetID uuid.UUID) error

	CreateAPIKey(ctx context.Context, apiKey string, userID uuid.UUID) error

	RemoveUserRoles(ctx context.Context, userID uuid.UUID) error
	RemoveApiKeysFromUser(ctx context.Context, userID uuid.UUID) error

	GetRoleNames() []sqlc.RoleName
	GetSystemNames() []sqlc.SystemName

	Transaction(ctx context.Context, fn TransactionFunc) error

	LoadSystemState(ctx context.Context, systemName sqlc.SystemName, teamID uuid.UUID, state interface{}) error
	SetSystemState(ctx context.Context, systemName sqlc.SystemName, teamID uuid.UUID, state interface{}) error

	SetUserName(ctx context.Context, userID uuid.UUID, name string) (*User, error)
}

func NewDatabase(q Querier, conn *pgx.Conn) Database {
	return &database{querier: q, conn: conn}
}

func nullString(s *string) sql.NullString {
	ns := sql.NullString{}
	ns.Scan(s)
	return ns
}

func nullUUID(ID *uuid.UUID) uuid.NullUUID {
	nu := uuid.NullUUID{}
	nu.Scan(ID)
	return nu
}
