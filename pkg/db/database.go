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

type Database interface {
	AddAuditLog(ctx context.Context, correlationId uuid.UUID, actorEmail *string, systemName *sqlc.SystemName, targetTeamSlug, targetUserEmail *string, action sqlc.AuditAction, message string) error
	AddUser(ctx context.Context, name, email string) (*User, error)
	GetUserByID(ctx context.Context, id uuid.UUID) (*User, error)
	GetUserByEmail(ctx context.Context, email string) (*User, error)
	GetUserByApiKey(ctx context.Context, apiKey string) (*User, error)

	AddTeam(ctx context.Context, name, slug string, purpose *string) (*Team, error)
	GetTeamByID(ctx context.Context, id uuid.UUID) (*Team, error)
	GetTeamBySlug(ctx context.Context, slug string) (*Team, error)
	GetTeams(ctx context.Context) ([]*Team, error)

	AddUserRole(ctx context.Context, userID uuid.UUID, roleName sqlc.RoleName) error
	CreateAPIKey(ctx context.Context, apiKey string, userID uuid.UUID) error

	RemoveUserRoles(ctx context.Context, userID uuid.UUID) error
	RemoveApiKeysFromUser(ctx context.Context, userID uuid.UUID) error

	LoadSystemState(ctx context.Context, systemName sqlc.SystemName, teamId uuid.UUID, state interface{}) error
	SetSystemState(ctx context.Context, systemName sqlc.SystemName, teamId uuid.UUID, state interface{}) error
}

func NewDatabase(q Querier, conn *pgx.Conn) Database {
	return &database{querier: q, conn: conn}
}

func nullString(s *string) sql.NullString {
	ns := sql.NullString{}
	ns.Scan(s)
	return ns
}
