// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.15.0

package sqlc

import (
	"context"

	"github.com/google/uuid"
	"github.com/nais/console/pkg/slug"
)

type Querier interface {
	AssignGlobalRoleToUser(ctx context.Context, arg AssignGlobalRoleToUserParams) error
	AssignTargetedRoleToUser(ctx context.Context, arg AssignTargetedRoleToUserParams) error
	ClearReconcilerErrorsForTeam(ctx context.Context, arg ClearReconcilerErrorsForTeamParams) error
	ConfigureReconciler(ctx context.Context, arg ConfigureReconcilerParams) error
	CreateAPIKey(ctx context.Context, arg CreateAPIKeyParams) error
	CreateAuditLog(ctx context.Context, arg CreateAuditLogParams) error
	CreateServiceAccount(ctx context.Context, name string) (*User, error)
	CreateSession(ctx context.Context, arg CreateSessionParams) (*Session, error)
	CreateTeam(ctx context.Context, arg CreateTeamParams) (*Team, error)
	CreateUser(ctx context.Context, arg CreateUserParams) (*User, error)
	DeleteServiceAccount(ctx context.Context, id uuid.UUID) error
	DeleteSession(ctx context.Context, id uuid.UUID) error
	DeleteUser(ctx context.Context, id uuid.UUID) error
	DisableReconciler(ctx context.Context, name ReconcilerName) (*Reconciler, error)
	EnableReconciler(ctx context.Context, name ReconcilerName) (*Reconciler, error)
	GetAuditLogsForTeam(ctx context.Context, targetTeamSlug *slug.Slug) ([]*AuditLog, error)
	GetEnabledReconcilers(ctx context.Context) ([]*Reconciler, error)
	GetReconciler(ctx context.Context, name ReconcilerName) (*Reconciler, error)
	GetReconcilerConfig(ctx context.Context, reconciler ReconcilerName) ([]*GetReconcilerConfigRow, error)
	GetReconcilerStateForTeam(ctx context.Context, arg GetReconcilerStateForTeamParams) (*ReconcilerState, error)
	GetReconcilers(ctx context.Context) ([]*Reconciler, error)
	GetRoleAuthorizations(ctx context.Context, roleName RoleName) ([]AuthzName, error)
	GetServiceAccountByApiKey(ctx context.Context, apiKey string) (*User, error)
	GetServiceAccountByName(ctx context.Context, name string) (*User, error)
	GetServiceAccounts(ctx context.Context) ([]*User, error)
	GetSessionByID(ctx context.Context, id uuid.UUID) (*Session, error)
	GetTeamByID(ctx context.Context, id uuid.UUID) (*Team, error)
	GetTeamBySlug(ctx context.Context, slug slug.Slug) (*Team, error)
	GetTeamMembers(ctx context.Context, teamID uuid.UUID) ([]*User, error)
	GetTeamMetadata(ctx context.Context, teamID uuid.UUID) ([]*TeamMetadatum, error)
	GetTeamReconcilerErrors(ctx context.Context, teamID uuid.UUID) ([]*ReconcilerError, error)
	GetTeams(ctx context.Context) ([]*Team, error)
	GetUserByEmail(ctx context.Context, email string) (*User, error)
	GetUserByExternalID(ctx context.Context, externalID string) (*User, error)
	GetUserByID(ctx context.Context, id uuid.UUID) (*User, error)
	GetUserRoles(ctx context.Context, userID uuid.UUID) ([]*UserRole, error)
	GetUserTeams(ctx context.Context, userID uuid.UUID) ([]*Team, error)
	GetUsers(ctx context.Context) ([]*User, error)
	RemoveAllUserRoles(ctx context.Context, userID uuid.UUID) error
	RemoveApiKeysFromServiceAccount(ctx context.Context, userID uuid.UUID) error
	RemoveGlobalUserRole(ctx context.Context, arg RemoveGlobalUserRoleParams) error
	ResetReconcilerConfig(ctx context.Context, reconciler ReconcilerName) error
	RevokeGlobalRoleFromUser(ctx context.Context, arg RevokeGlobalRoleFromUserParams) error
	RevokeTargetedRoleFromUser(ctx context.Context, arg RevokeTargetedRoleFromUserParams) error
	SetReconcilerErrorForTeam(ctx context.Context, arg SetReconcilerErrorForTeamParams) error
	SetReconcilerStateForTeam(ctx context.Context, arg SetReconcilerStateForTeamParams) error
	SetSessionExpires(ctx context.Context, arg SetSessionExpiresParams) (*Session, error)
	SetTeamMetadata(ctx context.Context, arg SetTeamMetadataParams) error
	UpdateTeam(ctx context.Context, arg UpdateTeamParams) (*Team, error)
	UpdateUser(ctx context.Context, arg UpdateUserParams) (*User, error)
}

var _ Querier = (*Queries)(nil)
