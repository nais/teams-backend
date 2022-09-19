package db

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/nais/console/pkg/slug"
	"github.com/nais/console/pkg/sqlc"
)

type (
	QuerierTransactionFunc  func(querier Querier) error
	DatabaseTransactionFunc func(ctx context.Context, dbtx Database) error
)

type AuthenticatedUser interface {
	GetID() uuid.UUID
	Identity() string
	IsServiceAccount() bool
}

type AuditLog struct {
	*sqlc.AuditLog
}

type ReconcileError struct {
	*sqlc.ReconcileError
}

type Role struct {
	*sqlc.UserRole
	Name           sqlc.RoleName
	Authorizations []sqlc.AuthzName
}

type ServiceAccount struct {
	ID   uuid.UUID
	Name string
}

type Session struct {
	*sqlc.Session
}

type TeamMetadata map[string]string

type Team struct {
	*sqlc.Team
}

type User struct {
	ID         uuid.UUID
	Email      string
	ExternalID string
	Name       string
}

type Querier interface {
	sqlc.Querier
	Transaction(ctx context.Context, callback QuerierTransactionFunc) error
}

type Queries struct {
	*sqlc.Queries
	connPool *pgxpool.Pool
	tx       pgx.Tx
}

type database struct {
	querier Querier
}

type Database interface {
	CreateAuditLogEntry(ctx context.Context, correlationID uuid.UUID, systemName sqlc.SystemName, actor *string, targetTeamSlug *slug.Slug, targetUser *string, action sqlc.AuditAction, message string) error
	CreateUser(ctx context.Context, name, email, externalID string) (*User, error)
	CreateServiceAccount(ctx context.Context, name string) (*ServiceAccount, error)
	GetServiceAccountByName(ctx context.Context, name string) (*ServiceAccount, error)
	GetUserByID(ctx context.Context, ID uuid.UUID) (*User, error)
	GetUserByEmail(ctx context.Context, email string) (*User, error)
	GetServiceAccountByApiKey(ctx context.Context, APIKey string) (*ServiceAccount, error)
	DeleteUser(ctx context.Context, userID uuid.UUID) error
	GetUsers(ctx context.Context) ([]*User, error)
	GetUserTeams(ctx context.Context, userID uuid.UUID) ([]*Team, error)
	CreateTeam(ctx context.Context, name string, slug slug.Slug, purpose *string) (*Team, error)
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
	CreateAPIKey(ctx context.Context, apiKey string, serviceAccountID uuid.UUID) error
	RemoveAllUserRoles(ctx context.Context, userID uuid.UUID) error
	RemoveApiKeysFromServiceAccount(ctx context.Context, serviceAccountID uuid.UUID) error
	GetUserRoles(ctx context.Context, userID uuid.UUID) ([]*Role, error)
	Transaction(ctx context.Context, fn DatabaseTransactionFunc) error
	LoadSystemState(ctx context.Context, systemName sqlc.SystemName, teamID uuid.UUID, state interface{}) error
	SetSystemState(ctx context.Context, systemName sqlc.SystemName, teamID uuid.UUID, state interface{}) error
	UpdateUser(ctx context.Context, userID uuid.UUID, name, email, externalID string) (*User, error)
	SetTeamReconcileErrorForSystem(ctx context.Context, correlationID uuid.UUID, teamID uuid.UUID, systemName sqlc.SystemName, err error) error
	GetTeamReconcileErrors(ctx context.Context, teamID uuid.UUID) ([]*ReconcileError, error)
	ClearTeamReconcileErrorForSystem(ctx context.Context, teamID uuid.UUID, systemName sqlc.SystemName) error
	GetServiceAccounts(ctx context.Context) ([]*ServiceAccount, error)
	DeleteServiceAccount(ctx context.Context, serviceAccountID uuid.UUID) error
	GetUserByExternalID(ctx context.Context, externalID string) (*User, error)
	CreateSession(ctx context.Context, userID uuid.UUID) (*Session, error)
	DeleteSession(ctx context.Context, sessionID uuid.UUID) error
	GetSessionByID(ctx context.Context, sessionID uuid.UUID) (*Session, error)
	ExtendSession(ctx context.Context, sessionID uuid.UUID) (*Session, error)
}

func (u User) GetID() uuid.UUID {
	return u.ID
}

func (u User) Identity() string {
	return u.Email
}

func (u User) IsServiceAccount() bool {
	return false
}

func (s ServiceAccount) GetID() uuid.UUID {
	return s.ID
}

func (s ServiceAccount) Identity() string {
	return s.Name
}

func (s ServiceAccount) IsServiceAccount() bool {
	return true
}
