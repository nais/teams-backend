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
	QuerierTransactionFunc  func(ctx context.Context, querier Querier) error
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

type Reconciler struct {
	*sqlc.Reconciler
}

type ReconcilerConfig struct {
	*sqlc.GetReconcilerConfigRow
}

type ReconcilerConfigValues struct {
	values map[sqlc.ReconcilerConfigKey]string
}

func (v ReconcilerConfigValues) GetValue(s sqlc.ReconcilerConfigKey) string {
	if v, exists := v.values[s]; exists {
		return v
	}
	return ""
}

type ReconcilerError struct {
	*sqlc.ReconcilerError
}

type Role struct {
	Authorizations         []sqlc.AuthzName
	RoleName               sqlc.RoleName
	TargetServiceAccountID *uuid.UUID
	TargetTeamSlug         *slug.Slug
}

type ServiceAccount struct {
	*sqlc.ServiceAccount
}

type Session struct {
	*sqlc.Session
}

type TeamMetadata struct {
	Key   string
	Value *string
}

type Team struct {
	*sqlc.Team
}

type User struct {
	*sqlc.User
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
	CreateAuditLogEntry(ctx context.Context, correlationID uuid.UUID, systemName sqlc.SystemName, actor *string, targetType sqlc.AuditLogsTargetType, targetIdentifier string, action sqlc.AuditAction, message string) error
	CreateUser(ctx context.Context, name, email, externalID string) (*User, error)
	CreateServiceAccount(ctx context.Context, name string) (*ServiceAccount, error)
	GetServiceAccountByName(ctx context.Context, name string) (*ServiceAccount, error)
	GetUserByID(ctx context.Context, ID uuid.UUID) (*User, error)
	GetUserByEmail(ctx context.Context, email string) (*User, error)
	GetServiceAccountByApiKey(ctx context.Context, APIKey string) (*ServiceAccount, error)
	DeleteUser(ctx context.Context, userID uuid.UUID) error
	GetUsers(ctx context.Context) ([]*User, error)
	GetUserTeams(ctx context.Context, userID uuid.UUID) ([]*Team, error)
	CreateTeam(ctx context.Context, slug slug.Slug, purpose string) (*Team, error)
	SetTeamMetadata(ctx context.Context, slug slug.Slug, metadata []TeamMetadata) error
	GetTeamMetadata(ctx context.Context, slug slug.Slug) ([]*TeamMetadata, error)
	UpdateTeam(ctx context.Context, teamID uuid.UUID, purpose *string) (*Team, error)
	GetTeamByID(ctx context.Context, ID uuid.UUID) (*Team, error)
	GetTeamBySlug(ctx context.Context, slug slug.Slug) (*Team, error)
	GetTeams(ctx context.Context) ([]*Team, error)
	GetTeamMembers(ctx context.Context, teamSlug slug.Slug) ([]*User, error)
	UserIsTeamOwner(ctx context.Context, userID uuid.UUID, teamSlug slug.Slug) (bool, error)
	SetTeamMemberRole(ctx context.Context, userID uuid.UUID, teamSlug slug.Slug, role sqlc.RoleName) error
	GetAuditLogsForTeam(ctx context.Context, slug slug.Slug) ([]*AuditLog, error)
	AssignGlobalRoleToUser(ctx context.Context, userID uuid.UUID, roleName sqlc.RoleName) error
	AssignGlobalRoleToServiceAccount(ctx context.Context, serviceAccountID uuid.UUID, roleName sqlc.RoleName) error
	RemoveUserFromTeam(ctx context.Context, userID uuid.UUID, teamSlug slug.Slug) error
	CreateAPIKey(ctx context.Context, apiKey string, serviceAccountID uuid.UUID) error
	RemoveAllUserRoles(ctx context.Context, userID uuid.UUID) error
	RemoveAllServiceAccountRoles(ctx context.Context, serviceAccountID uuid.UUID) error
	RemoveApiKeysFromServiceAccount(ctx context.Context, serviceAccountID uuid.UUID) error
	GetUserRoles(ctx context.Context, userID uuid.UUID) ([]*Role, error)
	GetServiceAccountRoles(ctx context.Context, serviceAccountID uuid.UUID) ([]*Role, error)
	Transaction(ctx context.Context, fn DatabaseTransactionFunc) error
	LoadReconcilerStateForTeam(ctx context.Context, reconcilerName sqlc.ReconcilerName, slug slug.Slug, state interface{}) error
	SetReconcilerStateForTeam(ctx context.Context, reconcilerName sqlc.ReconcilerName, slug slug.Slug, state interface{}) error
	UpdateUser(ctx context.Context, userID uuid.UUID, name, email, externalID string) (*User, error)
	SetReconcilerErrorForTeam(ctx context.Context, correlationID uuid.UUID, slug slug.Slug, reconcilerName sqlc.ReconcilerName, err error) error
	GetTeamReconcilerErrors(ctx context.Context, slug slug.Slug) ([]*ReconcilerError, error)
	ClearReconcilerErrorsForTeam(ctx context.Context, slug slug.Slug, reconcilerName sqlc.ReconcilerName) error
	GetServiceAccounts(ctx context.Context) ([]*ServiceAccount, error)
	DeleteServiceAccount(ctx context.Context, serviceAccountID uuid.UUID) error
	GetUserByExternalID(ctx context.Context, externalID string) (*User, error)
	CreateSession(ctx context.Context, userID uuid.UUID) (*Session, error)
	DeleteSession(ctx context.Context, sessionID uuid.UUID) error
	GetSessionByID(ctx context.Context, sessionID uuid.UUID) (*Session, error)
	ExtendSession(ctx context.Context, sessionID uuid.UUID) (*Session, error)
	GetReconciler(ctx context.Context, reconcilerName sqlc.ReconcilerName) (*Reconciler, error)
	GetReconcilers(ctx context.Context) ([]*Reconciler, error)
	GetEnabledReconcilers(ctx context.Context) ([]*Reconciler, error)
	ConfigureReconciler(ctx context.Context, reconcilerName sqlc.ReconcilerName, config map[sqlc.ReconcilerConfigKey]string) (*Reconciler, error)
	GetReconcilerConfig(ctx context.Context, reconcilerName sqlc.ReconcilerName) ([]*ReconcilerConfig, error)
	ResetReconcilerConfig(ctx context.Context, reconcilerName sqlc.ReconcilerName) (*Reconciler, error)
	EnableReconciler(ctx context.Context, reconcilerName sqlc.ReconcilerName) (*Reconciler, error)
	DisableReconciler(ctx context.Context, reconcilerName sqlc.ReconcilerName) (*Reconciler, error)
	DangerousGetReconcilerConfigValues(ctx context.Context, reconcilerName sqlc.ReconcilerName) (*ReconcilerConfigValues, error)
	GetAuditLogsForReconciler(ctx context.Context, reconcilerName sqlc.ReconcilerName) ([]*AuditLog, error)
	DisableTeam(ctx context.Context, teamID uuid.UUID) (*Team, error)
	EnableTeam(ctx context.Context, teamID uuid.UUID) (*Team, error)
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
