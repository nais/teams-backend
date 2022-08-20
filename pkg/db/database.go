package db

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v4"
	"github.com/nais/console/pkg/sqlc"
)

type database struct {
	querier Querier
	conn    *pgx.Conn
}

type Database interface {
	AddAuditLog(ctx context.Context, auditLog AuditLog) (*AuditLog, error)
	AddUser(ctx context.Context, user User) (*User, error)
	GetUserByID(ctx context.Context, id uuid.UUID) (*User, error)
	GetUserByEmail(ctx context.Context, email string) (*User, error)
	GetUserByApiKey(ctx context.Context, apiKey string) (*User, error)

	AddTeam(ctx context.Context, team Team) (*Team, error)
	GetTeamByID(ctx context.Context, id uuid.UUID) (*Team, error)
	GetTeamBySlug(ctx context.Context, slug string) (*Team, error)
	GetTeams(ctx context.Context) ([]*Team, error)

	AddUserRole(ctx context.Context, userID uuid.UUID, roleName sqlc.RoleName) error
	CreateAPIKey(ctx context.Context, apiKey string, userID uuid.UUID) error
}

func NewDatabase(q Querier, conn *pgx.Conn) Database {
	return &database{querier: q, conn: conn}
}
