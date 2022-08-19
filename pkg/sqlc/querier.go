// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.15.0

package sqlc

import (
	"context"

	"github.com/google/uuid"
	"github.com/nais/console/pkg/roles"
)

type Querier interface {
	AddRoleToUser(ctx context.Context, arg AddRoleToUserParams) error
	AddUserToTeam(ctx context.Context, arg AddUserToTeamParams) error
	CreateCorrelation(ctx context.Context, id uuid.UUID) (*Correlation, error)
	CreateSystem(ctx context.Context, arg CreateSystemParams) (*System, error)
	CreateTeam(ctx context.Context, arg CreateTeamParams) (*Team, error)
	GetRole(ctx context.Context, id uuid.UUID) (*Role, error)
	GetRoleByName(ctx context.Context, name roles.Role) (*Role, error)
	GetRoles(ctx context.Context) ([]*Role, error)
	GetSystem(ctx context.Context, id uuid.UUID) (*System, error)
	GetSystemByName(ctx context.Context, name string) (*System, error)
	GetSystems(ctx context.Context) ([]*System, error)
	GetTeam(ctx context.Context, id uuid.UUID) (*Team, error)
	GetTeamMembers(ctx context.Context, teamID uuid.UUID) ([]*User, error)
	GetTeamMetadata(ctx context.Context, teamID uuid.UUID) (*TeamMetadatum, error)
	GetTeams(ctx context.Context) ([]*Team, error)
	GetUser(ctx context.Context, id uuid.UUID) (*User, error)
	GetUserRole(ctx context.Context, id uuid.UUID) (*UserRole, error)
	GetUserRoles(ctx context.Context, userID uuid.UUID) ([]*UserRole, error)
	GetUserTeams(ctx context.Context, userID uuid.UUID) ([]*UserTeam, error)
}

var _ Querier = (*Queries)(nil)
