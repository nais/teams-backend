// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.16.0

package sqlc

import (
	"context"

	"github.com/google/uuid"
	"github.com/nais/console/pkg/slug"
)

type Querier interface {
	AssignGlobalRoleToServiceAccount(ctx context.Context, arg AssignGlobalRoleToServiceAccountParams) error
	AssignGlobalRoleToUser(ctx context.Context, arg AssignGlobalRoleToUserParams) error
	AssignTeamRoleToUser(ctx context.Context, arg AssignTeamRoleToUserParams) error
	ClearReconcilerErrorsForTeam(ctx context.Context, arg ClearReconcilerErrorsForTeamParams) error
	ConfigureReconciler(ctx context.Context, arg ConfigureReconcilerParams) error
	CreateAPIKey(ctx context.Context, arg CreateAPIKeyParams) error
	CreateAuditLog(ctx context.Context, arg CreateAuditLogParams) error
	CreateServiceAccount(ctx context.Context, name string) (*ServiceAccount, error)
	CreateSession(ctx context.Context, arg CreateSessionParams) (*Session, error)
	CreateTeam(ctx context.Context, arg CreateTeamParams) (*Team, error)
	CreateUser(ctx context.Context, arg CreateUserParams) (*User, error)
	DangerousGetReconcilerConfigValues(ctx context.Context, reconciler ReconcilerName) ([]*DangerousGetReconcilerConfigValuesRow, error)
	DeleteServiceAccount(ctx context.Context, id uuid.UUID) error
	DeleteSession(ctx context.Context, id uuid.UUID) error
	DeleteUser(ctx context.Context, id uuid.UUID) error
	DisableReconciler(ctx context.Context, name ReconcilerName) (*Reconciler, error)
	DisableTeam(ctx context.Context, slug slug.Slug) (*Team, error)
	EnableReconciler(ctx context.Context, name ReconcilerName) (*Reconciler, error)
	EnableTeam(ctx context.Context, slug slug.Slug) (*Team, error)
	FirstRunComplete(ctx context.Context) error
	GetAuditLogsForReconciler(ctx context.Context, targetIdentifier string) ([]*AuditLog, error)
	GetAuditLogsForTeam(ctx context.Context, targetIdentifier string) ([]*AuditLog, error)
	GetEnabledReconcilers(ctx context.Context) ([]*Reconciler, error)
	GetReconciler(ctx context.Context, name ReconcilerName) (*Reconciler, error)
	GetReconcilerConfig(ctx context.Context, reconciler ReconcilerName) ([]*GetReconcilerConfigRow, error)
	GetReconcilerStateForTeam(ctx context.Context, arg GetReconcilerStateForTeamParams) (*ReconcilerState, error)
	GetReconcilers(ctx context.Context) ([]*Reconciler, error)
	GetRoleAuthorizations(ctx context.Context, roleName RoleName) ([]AuthzName, error)
	GetServiceAccountByApiKey(ctx context.Context, apiKey string) (*ServiceAccount, error)
	GetServiceAccountByName(ctx context.Context, name string) (*ServiceAccount, error)
	GetServiceAccountRoles(ctx context.Context, serviceAccountID uuid.UUID) ([]*ServiceAccountRole, error)
	GetServiceAccounts(ctx context.Context) ([]*ServiceAccount, error)
	GetSessionByID(ctx context.Context, id uuid.UUID) (*Session, error)
	GetTeamBySlug(ctx context.Context, slug slug.Slug) (*Team, error)
	GetTeamMembers(ctx context.Context, targetTeamSlug *slug.Slug) ([]*User, error)
	GetTeamMetadata(ctx context.Context, teamSlug slug.Slug) ([]*TeamMetadatum, error)
	GetTeamReconcilerErrors(ctx context.Context, teamSlug slug.Slug) ([]*ReconcilerError, error)
	GetTeams(ctx context.Context) ([]*Team, error)
	GetUserByEmail(ctx context.Context, email string) (*User, error)
	GetUserByExternalID(ctx context.Context, externalID string) (*User, error)
	GetUserByID(ctx context.Context, id uuid.UUID) (*User, error)
	GetUserRoles(ctx context.Context, userID uuid.UUID) ([]*UserRole, error)
	GetUserTeams(ctx context.Context, userID uuid.UUID) ([]*Team, error)
	GetUsers(ctx context.Context) ([]*User, error)
	GetUsersWithGloballyAssignedRole(ctx context.Context, roleName RoleName) ([]*User, error)
	IsFirstRun(ctx context.Context) (bool, error)
	RemoveAllServiceAccountRoles(ctx context.Context, serviceAccountID uuid.UUID) error
	RemoveApiKeysFromServiceAccount(ctx context.Context, serviceAccountID uuid.UUID) error
	RemoveUserFromTeam(ctx context.Context, arg RemoveUserFromTeamParams) error
	ResetReconcilerConfig(ctx context.Context, reconciler ReconcilerName) error
	RevokeGlobalUserRole(ctx context.Context, arg RevokeGlobalUserRoleParams) error
	SetLastSuccessfulSyncForTeam(ctx context.Context, slug slug.Slug) error
	SetReconcilerErrorForTeam(ctx context.Context, arg SetReconcilerErrorForTeamParams) error
	SetReconcilerStateForTeam(ctx context.Context, arg SetReconcilerStateForTeamParams) error
	SetSessionExpires(ctx context.Context, arg SetSessionExpiresParams) (*Session, error)
	SetTeamMetadata(ctx context.Context, arg SetTeamMetadataParams) error
	UpdateTeam(ctx context.Context, arg UpdateTeamParams) (*Team, error)
	UpdateUser(ctx context.Context, arg UpdateUserParams) (*User, error)
}

var _ Querier = (*Queries)(nil)
