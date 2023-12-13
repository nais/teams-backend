// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.20.0

package sqlc

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgtype"
	"github.com/nais/teams-backend/pkg/slug"
)

type Querier interface {
	AddReconcilerOptOut(ctx context.Context, arg AddReconcilerOptOutParams) error
	AssignGlobalRoleToServiceAccount(ctx context.Context, arg AssignGlobalRoleToServiceAccountParams) error
	AssignGlobalRoleToUser(ctx context.Context, arg AssignGlobalRoleToUserParams) error
	AssignTeamRoleToServiceAccount(ctx context.Context, arg AssignTeamRoleToServiceAccountParams) error
	AssignTeamRoleToUser(ctx context.Context, arg AssignTeamRoleToUserParams) error
	ClearReconcilerErrorsForTeam(ctx context.Context, arg ClearReconcilerErrorsForTeamParams) error
	ConfigureReconciler(ctx context.Context, arg ConfigureReconcilerParams) error
	ConfirmTeamDeleteKey(ctx context.Context, key uuid.UUID) error
	CreateAPIKey(ctx context.Context, arg CreateAPIKeyParams) error
	CreateAuditLog(ctx context.Context, arg CreateAuditLogParams) error
	CreateRepositoryAuthorization(ctx context.Context, arg CreateRepositoryAuthorizationParams) error
	CreateServiceAccount(ctx context.Context, name string) (*ServiceAccount, error)
	CreateSession(ctx context.Context, arg CreateSessionParams) (*Session, error)
	CreateTeam(ctx context.Context, arg CreateTeamParams) (*Team, error)
	CreateTeamDeleteKey(ctx context.Context, arg CreateTeamDeleteKeyParams) (*TeamDeleteKey, error)
	CreateUser(ctx context.Context, arg CreateUserParams) (*User, error)
	DangerousGetReconcilerConfigValues(ctx context.Context, reconciler ReconcilerName) ([]*DangerousGetReconcilerConfigValuesRow, error)
	DeleteServiceAccount(ctx context.Context, id uuid.UUID) error
	DeleteSession(ctx context.Context, id uuid.UUID) error
	DeleteTeam(ctx context.Context, argSlug slug.Slug) error
	DeleteUser(ctx context.Context, id uuid.UUID) error
	DisableReconciler(ctx context.Context, name ReconcilerName) (*Reconciler, error)
	EnableReconciler(ctx context.Context, name ReconcilerName) (*Reconciler, error)
	FirstRunComplete(ctx context.Context) error
	GetActiveTeamBySlug(ctx context.Context, argSlug slug.Slug) (*Team, error)
	GetActiveTeams(ctx context.Context) ([]*Team, error)
	GetAllTeamMembers(ctx context.Context, targetTeamSlug *slug.Slug) ([]*User, error)
	GetAllUserRoles(ctx context.Context) ([]*UserRole, error)
	GetAllUsers(ctx context.Context) ([]*User, error)
	GetAuditLogsForCorrelationID(ctx context.Context, arg GetAuditLogsForCorrelationIDParams) ([]*AuditLog, error)
	GetAuditLogsForCorrelationIDCount(ctx context.Context, correlationID uuid.UUID) (int64, error)
	GetAuditLogsForReconciler(ctx context.Context, arg GetAuditLogsForReconcilerParams) ([]*AuditLog, error)
	GetAuditLogsForReconcilerCount(ctx context.Context, targetIdentifier string) (int64, error)
	GetAuditLogsForTeam(ctx context.Context, arg GetAuditLogsForTeamParams) ([]*AuditLog, error)
	GetAuditLogsForTeamCount(ctx context.Context, targetIdentifier string) (int64, error)
	GetEnabledReconcilers(ctx context.Context) ([]*Reconciler, error)
	GetReconciler(ctx context.Context, name ReconcilerName) (*Reconciler, error)
	GetReconcilerConfig(ctx context.Context, reconciler ReconcilerName) ([]*GetReconcilerConfigRow, error)
	GetReconcilerStateForTeam(ctx context.Context, arg GetReconcilerStateForTeamParams) (*ReconcilerState, error)
	GetReconcilers(ctx context.Context) ([]*Reconciler, error)
	GetRepositoryAuthorizations(ctx context.Context, arg GetRepositoryAuthorizationsParams) ([]RepositoryAuthorizationEnum, error)
	GetServiceAccountByApiKey(ctx context.Context, apiKey string) (*ServiceAccount, error)
	GetServiceAccountByName(ctx context.Context, name string) (*ServiceAccount, error)
	GetServiceAccountRoles(ctx context.Context, serviceAccountID uuid.UUID) ([]*ServiceAccountRole, error)
	GetServiceAccounts(ctx context.Context) ([]*ServiceAccount, error)
	GetSessionByID(ctx context.Context, id uuid.UUID) (*Session, error)
	GetSlackAlertsChannels(ctx context.Context, teamSlug slug.Slug) ([]*SlackAlertsChannel, error)
	GetTeamBySlug(ctx context.Context, argSlug slug.Slug) (*Team, error)
	GetTeamDeleteKey(ctx context.Context, key uuid.UUID) (*TeamDeleteKey, error)
	GetTeamMember(ctx context.Context, arg GetTeamMemberParams) (*User, error)
	GetTeamMemberOptOuts(ctx context.Context, arg GetTeamMemberOptOutsParams) ([]*GetTeamMemberOptOutsRow, error)
	GetTeamMembers(ctx context.Context, arg GetTeamMembersParams) ([]*User, error)
	GetTeamMembersCount(ctx context.Context, targetTeamSlug *slug.Slug) (int64, error)
	GetTeamMembersForReconciler(ctx context.Context, arg GetTeamMembersForReconcilerParams) ([]*User, error)
	GetTeamReconcilerErrors(ctx context.Context, teamSlug slug.Slug) ([]*ReconcilerError, error)
	GetTeams(ctx context.Context) ([]*Team, error)
	GetTeamsCount(ctx context.Context) (int64, error)
	GetTeamsPaginated(ctx context.Context, arg GetTeamsPaginatedParams) ([]*Team, error)
	GetTeamsWithPermissionInGitHubRepo(ctx context.Context, arg GetTeamsWithPermissionInGitHubRepoParams) ([]*Team, error)
	GetTeamsWithPermissionInGitHubRepoCount(ctx context.Context, state pgtype.JSONB) (int64, error)
	GetUserByEmail(ctx context.Context, email string) (*User, error)
	GetUserByExternalID(ctx context.Context, externalID string) (*User, error)
	GetUserByID(ctx context.Context, id uuid.UUID) (*User, error)
	GetUserRoles(ctx context.Context, userID uuid.UUID) ([]*UserRole, error)
	GetUserTeams(ctx context.Context, arg GetUserTeamsParams) ([]*GetUserTeamsRow, error)
	GetUsers(ctx context.Context, arg GetUsersParams) ([]*User, error)
	GetUsersCount(ctx context.Context) (int64, error)
	GetUsersWithGloballyAssignedRole(ctx context.Context, roleName RoleName) ([]*User, error)
	IsFirstRun(ctx context.Context) (bool, error)
	RemoveAllServiceAccountRoles(ctx context.Context, serviceAccountID uuid.UUID) error
	RemoveApiKeysFromServiceAccount(ctx context.Context, serviceAccountID uuid.UUID) error
	RemoveReconcilerOptOut(ctx context.Context, arg RemoveReconcilerOptOutParams) error
	RemoveReconcilerStateForTeam(ctx context.Context, arg RemoveReconcilerStateForTeamParams) error
	RemoveRepositoryAuthorization(ctx context.Context, arg RemoveRepositoryAuthorizationParams) error
	RemoveSlackAlertsChannel(ctx context.Context, arg RemoveSlackAlertsChannelParams) error
	RemoveUserFromTeam(ctx context.Context, arg RemoveUserFromTeamParams) error
	ResetReconcilerConfig(ctx context.Context, reconciler ReconcilerName) error
	RevokeGlobalUserRole(ctx context.Context, arg RevokeGlobalUserRoleParams) error
	SetLastSuccessfulSyncForTeam(ctx context.Context, argSlug slug.Slug) error
	SetReconcilerErrorForTeam(ctx context.Context, arg SetReconcilerErrorForTeamParams) error
	SetReconcilerStateForTeam(ctx context.Context, arg SetReconcilerStateForTeamParams) error
	SetSessionExpires(ctx context.Context, arg SetSessionExpiresParams) (*Session, error)
	SetSlackAlertsChannel(ctx context.Context, arg SetSlackAlertsChannelParams) error
	UpdateTeam(ctx context.Context, arg UpdateTeamParams) (*Team, error)
	UpdateUser(ctx context.Context, arg UpdateUserParams) (*User, error)
}

var _ Querier = (*Queries)(nil)
