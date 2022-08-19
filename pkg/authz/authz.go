package authz

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/nais/console/pkg/db"
	"github.com/nais/console/pkg/dbmodels"
	"github.com/nais/console/pkg/roles"
	"github.com/nais/console/pkg/sqlc"
)

const userContextKey = "user"

// ContextWithUser Return a context with a user module stored.
func ContextWithUser(ctx context.Context, user *db.User) context.Context {
	return context.WithValue(ctx, userContextKey, user)
}

// UserFromContext Finds any authenticated user from the context. Requires that a middleware has stored a user in the
// first place.
func UserFromContext(ctx context.Context) *db.User {
	user, _ := ctx.Value(userContextKey).(*db.User)
	return user
}

var ErrNotAuthorized = errors.New("not authorized")

// RequireGlobalAuthorization Require an actor to have a specific authorization through a globally assigned role. The
// role bindings must already be attached to the actor.
func RequireGlobalAuthorization(actor *sqlc.User, userRoles []*sqlc.UserRole, requiredAuthorization roles.Authorization) error {
	if actor == nil {
		return ErrNotAuthorized
	}

	authorizations := make(map[dbmodels.Authorization]struct{})

	for _, roleBinding := range actor.RoleBindings {
		if roleBinding.TargetID == nil {
			for _, authorization := range roleBinding.Role.Authorizations {
				authorizations[authorization] = struct{}{}
			}
		}
	}

	return authorized(authorizations, requiredAuthorization)
}

// RequireAuthorization Require an actor to have a specific authorization through a globally assigned or a correctly
// targetted role. The role bindings must already be attached to the actor.
func RequireAuthorization(actor *dbmodels.User, requiredAuthorization roles.Authorization, target uuid.UUID) error {
	if actor == nil {
		return ErrNotAuthorized
	}

	authorizations := make(map[dbmodels.Authorization]struct{})

	for _, roleBinding := range actor.RoleBindings {
		if roleBinding.TargetID == nil || *roleBinding.TargetID == target {
			for _, authorization := range roleBinding.Role.Authorizations {
				authorizations[authorization] = struct{}{}
			}
		}
	}

	return authorized(authorizations, requiredAuthorization)
}

// RequireAuthorizationOrTargetMatch Require an actor to have a specific authorization through a globally assigned or a
// correctly targetted role. The role bindings must already be attached to the actor. If the actor matches the target,
// the action will be allowed.
func RequireAuthorizationOrTargetMatch(actor *dbmodels.User, requiredAuthorization roles.Authorization, target uuid.UUID) error {
	if actor != nil && *actor.ID == target {
		return nil
	}

	return RequireAuthorization(actor, requiredAuthorization, target)
}

// authorized Check if one of the authorizations in the map matches the required authorization.
func authorized(authorizations map[dbmodels.Authorization]struct{}, requiredAuthorization roles.Authorization) error {
	for authorization := range authorizations {
		if roles.Authorization(authorization.Name) == requiredAuthorization {
			return nil
		}
	}

	return ErrNotAuthorized
}
