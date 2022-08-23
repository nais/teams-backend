package db

import (
	"context"
	"database/sql"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/nais/console/pkg/sqlc"
)

type database struct {
	querier  Querier
	connPool *pgxpool.Pool
}

type TransactionFunc func(ctx context.Context, dbtx Database) error

var (
	ErrNoRows = pgx.ErrNoRows
)

type Database interface {
	AddAuditLog(ctx context.Context, correlationID uuid.UUID, systemName sqlc.SystemName, actorEmail, targetTeamSlug, targetUserEmail *string, action sqlc.AuditAction, message string) error
	AddUser(ctx context.Context, name, email string) (*User, error)
	GetUserByID(ctx context.Context, ID uuid.UUID) (*User, error)
	GetUserByEmail(ctx context.Context, email string) (*User, error)
	GetUserByApiKey(ctx context.Context, APIKey string) (*User, error)
	DeleteUser(ctx context.Context, userID uuid.UUID) error

	GetUsersByEmail(ctx context.Context, email string) ([]*User, error)
	GetUsers(ctx context.Context) ([]*User, error)
	GetUserTeams(ctx context.Context, userID uuid.UUID) ([]*Team, error)

	AddTeam(ctx context.Context, name, slug string, purpose *string, userID uuid.UUID) (*Team, error)
	UpdateTeam(ctx context.Context, teamID uuid.UUID, name, purpose *string) (*Team, error)
	GetTeamByID(ctx context.Context, ID uuid.UUID) (*Team, error)
	GetTeamBySlug(ctx context.Context, slug string) (*Team, error)
	GetTeams(ctx context.Context) ([]*Team, error)
	GetTeamMembers(ctx context.Context, teamID uuid.UUID) ([]*User, error)
	UserIsTeamOwner(ctx context.Context, userID, teamID uuid.UUID) (bool, error)

	GetAuditLogsForTeam(ctx context.Context, slug string) ([]*AuditLog, error)

	AssignGlobalRoleToUser(ctx context.Context, userID uuid.UUID, roleName sqlc.RoleName) error
	AssignTargetedRoleToUser(ctx context.Context, userID uuid.UUID, roleName sqlc.RoleName, targetID uuid.UUID) error

	AddUserToTeam(ctx context.Context, userID uuid.UUID, teamID uuid.UUID) error

	CreateAPIKey(ctx context.Context, apiKey string, userID uuid.UUID) error

	RemoveUserRoles(ctx context.Context, userID uuid.UUID) error
	RemoveApiKeysFromUser(ctx context.Context, userID uuid.UUID) error

	GetRoleNames() []sqlc.RoleName
	GetSystemNames() []sqlc.SystemName

	GetUserRoles(ctx context.Context, userID uuid.UUID) ([]*Role, error)

	Transaction(ctx context.Context, fn TransactionFunc) error

	LoadSystemState(ctx context.Context, systemName sqlc.SystemName, teamID uuid.UUID, state interface{}) error
	SetSystemState(ctx context.Context, systemName sqlc.SystemName, teamID uuid.UUID, state interface{}) error

	SetUserName(ctx context.Context, userID uuid.UUID, name string) (*User, error)
}

func NewDatabase(q Querier, conn *pgxpool.Pool) Database {
	return &database{querier: q, connPool: conn}
}

func NullStringToStringP(ns sql.NullString) *string {
	var strP *string
	if ns.String != "" {
		strP = &ns.String
	}
	return strP
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
