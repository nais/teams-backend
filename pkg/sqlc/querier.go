// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.15.0

package sqlc

import (
	"context"

	"github.com/google/uuid"
)

type Querier interface {
	AddUserToTeam(ctx context.Context, arg AddUserToTeamParams) error
	CreateAPIKey(ctx context.Context, arg CreateAPIKeyParams) error
	CreateAuditLog(ctx context.Context, arg CreateAuditLogParams) error
	CreateTeam(ctx context.Context, arg CreateTeamParams) (*Team, error)
	CreateUser(ctx context.Context, arg CreateUserParams) (*User, error)
	CreateUserRole(ctx context.Context, arg CreateUserRoleParams) error
	DeleteUser(ctx context.Context, id uuid.UUID) error
	GetRoleAuthorizations(ctx context.Context, roleName RoleName) ([]AuthzName, error)
	GetTeamByID(ctx context.Context, id uuid.UUID) (*Team, error)
	GetTeamBySlug(ctx context.Context, slug string) (*Team, error)
	GetTeamMembers(ctx context.Context, teamID uuid.UUID) ([]*User, error)
	GetTeamMetadata(ctx context.Context, teamID uuid.UUID) ([]*TeamMetadatum, error)
	GetTeamSystemState(ctx context.Context, arg GetTeamSystemStateParams) (*SystemState, error)
	GetTeams(ctx context.Context) ([]*Team, error)
	GetUserByApiKey(ctx context.Context, apiKey string) (*User, error)
	GetUserByEmail(ctx context.Context, email string) (*User, error)
	GetUserByID(ctx context.Context, id uuid.UUID) (*User, error)
	GetUserRole(ctx context.Context, id int32) (*UserRole, error)
	GetUserRoles(ctx context.Context, userID uuid.UUID) ([]*UserRole, error)
	GetUserTeams(ctx context.Context, userID uuid.UUID) ([]*UserTeam, error)
	GetUsers(ctx context.Context) ([]*User, error)
	GetUsersByEmail(ctx context.Context, email string) ([]*User, error)
	RemoveApiKeysFromUser(ctx context.Context, userID uuid.UUID) error
	RemoveUserRoles(ctx context.Context, userID uuid.UUID) error
	SetTeamSystemState(ctx context.Context, arg SetTeamSystemStateParams) error
	SetUserName(ctx context.Context, arg SetUserNameParams) (*User, error)
	UpdateTeam(ctx context.Context, arg UpdateTeamParams) (*Team, error)
}

var _ Querier = (*Queries)(nil)
