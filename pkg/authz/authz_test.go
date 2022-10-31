package authz_test

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/nais/console/pkg/authz"
	"github.com/nais/console/pkg/db"
	"github.com/nais/console/pkg/sqlc"
	"github.com/stretchr/testify/assert"
)

const (
	authTeamCreateError = `required role: "teams:create"`
	authTeamUpdateError = `required role: "teams:update"`
)

func TestContextWithUser(t *testing.T) {
	ctx := context.Background()
	assert.Nil(t, authz.ActorFromContext(ctx))

	user := &db.User{
		User: &sqlc.User{
			Name:  "User Name",
			Email: "mail@example.com",
		},
	}

	roles := make([]*db.Role, 0)

	ctx = authz.ContextWithActor(ctx, user, roles)
	assert.Equal(t, user, authz.ActorFromContext(ctx).User)
	assert.Equal(t, roles, authz.ActorFromContext(ctx).Roles)
}

func TestRequireGlobalAuthorization(t *testing.T) {
	user := &db.User{
		User: &sqlc.User{
			Name:  "User Name",
			Email: "mail@example.com",
		},
	}

	t.Run("Nil user", func(t *testing.T) {
		assert.ErrorIs(t, authz.RequireGlobalAuthorization(nil, sqlc.AuthzNameTeamsCreate), authz.ErrNotAuthenticated)
	})

	t.Run("User with no roles", func(t *testing.T) {
		contextUser := authz.ActorFromContext(authz.ContextWithActor(context.Background(), user, []*db.Role{}))
		assert.EqualError(t, authz.RequireGlobalAuthorization(contextUser, sqlc.AuthzNameTeamsCreate), authTeamCreateError)
	})

	t.Run("User with insufficient roles", func(t *testing.T) {
		roles := []*db.Role{
			{
				RoleName:       sqlc.RoleNameTeamviewer,
				Authorizations: []sqlc.AuthzName{},
			},
		}
		contextUser := authz.ActorFromContext(authz.ContextWithActor(context.Background(), user, roles))
		assert.EqualError(t, authz.RequireGlobalAuthorization(contextUser, sqlc.AuthzNameTeamsCreate), authTeamCreateError)
	})

	t.Run("User with sufficient role", func(t *testing.T) {
		roles := []*db.Role{
			{
				RoleName:       sqlc.RoleNameTeamcreator,
				Authorizations: []sqlc.AuthzName{sqlc.AuthzNameTeamsCreate},
			},
		}
		contextUser := authz.ActorFromContext(authz.ContextWithActor(context.Background(), user, roles))
		assert.NoError(t, authz.RequireGlobalAuthorization(contextUser, sqlc.AuthzNameTeamsCreate))
	})
}

func TestRequireAuthorizationForTarget(t *testing.T) {
	user := &db.User{
		User: &sqlc.User{
			Name:  "User Name",
			Email: "mail@example.com",
		},
	}
	targetID := uuid.New()

	t.Run("Nil user", func(t *testing.T) {
		assert.ErrorIs(t, authz.RequireAuthorization(nil, sqlc.AuthzNameTeamsCreate, targetID), authz.ErrNotAuthenticated)
	})

	t.Run("User with no roles", func(t *testing.T) {
		contextUser := authz.ActorFromContext(authz.ContextWithActor(context.Background(), user, []*db.Role{}))
		assert.EqualError(t, authz.RequireAuthorization(contextUser, sqlc.AuthzNameTeamsCreate, targetID), authTeamCreateError)
	})

	t.Run("User with insufficient roles", func(t *testing.T) {
		roles := []*db.Role{
			{
				Authorizations: []sqlc.AuthzName{},
			},
		}
		contextUser := authz.ActorFromContext(authz.ContextWithActor(context.Background(), user, roles))
		assert.EqualError(t, authz.RequireAuthorization(contextUser, sqlc.AuthzNameTeamsUpdate, targetID), authTeamUpdateError)
	})

	t.Run("User with targeted role", func(t *testing.T) {
		roles := []*db.Role{
			{
				TargetID:       &targetID,
				Authorizations: []sqlc.AuthzName{sqlc.AuthzNameTeamsUpdate},
			},
		}
		contextUser := authz.ActorFromContext(authz.ContextWithActor(context.Background(), user, roles))
		assert.NoError(t, authz.RequireAuthorization(contextUser, sqlc.AuthzNameTeamsUpdate, targetID))
	})

	t.Run("User with targeted role for wrong target", func(t *testing.T) {
		wrongID := uuid.New()
		roles := []*db.Role{
			{
				TargetID:       &wrongID,
				Authorizations: []sqlc.AuthzName{sqlc.AuthzNameTeamsUpdate},
			},
		}
		contextUser := authz.ActorFromContext(authz.ContextWithActor(context.Background(), user, roles))
		assert.EqualError(t, authz.RequireAuthorization(contextUser, sqlc.AuthzNameTeamsUpdate, targetID), authTeamUpdateError)
	})

	t.Run("User with global role", func(t *testing.T) {
		roles := []*db.Role{
			{
				Authorizations: []sqlc.AuthzName{sqlc.AuthzNameTeamsUpdate},
			},
		}
		contextUser := authz.ActorFromContext(authz.ContextWithActor(context.Background(), user, roles))
		assert.NoError(t, authz.RequireAuthorization(contextUser, sqlc.AuthzNameTeamsUpdate, targetID))
	})
}
