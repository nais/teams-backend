package authz

import (
	"context"
	"errors"

	"github.com/nais/console/pkg/sqlc"

	"github.com/google/uuid"
	"github.com/nais/console/pkg/db"
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
// roles must already be attached to the actor.
func RequireGlobalAuthorization(actor *db.User, requiredAuthzName sqlc.AuthzName) error {
	if actor == nil {
		return ErrNotAuthorized
	}

	authorizations := make(map[sqlc.AuthzName]struct{})

	for _, role := range actor.Roles {
		if role.IsGlobal() {
			for _, authorization := range role.Authorizations {
				authorizations[authorization] = struct{}{}
			}
		}
	}

	return authorized(authorizations, requiredAuthzName)
}

// RequireAuthorization Require an actor to have a specific authorization through a globally assigned or a correctly
// targeted role. The roles must already be attached to the actor.
func RequireAuthorization(actor *db.User, requiredAuthzName sqlc.AuthzName, target uuid.UUID) error {
	if actor == nil {
		return ErrNotAuthorized
	}

	authorizations := make(map[sqlc.AuthzName]struct{})

	for _, role := range actor.Roles {
		if role.IsGlobal() || role.Targets(target) {
			for _, authorization := range role.Authorizations {
				authorizations[authorization] = struct{}{}
			}
		}
	}

	return authorized(authorizations, requiredAuthzName)
}

// RequireAuthorizationOrTargetMatch Require an actor to have a specific authorization through a globally assigned or a
// correctly targeted role. The roles must already be attached to the actor. If the actor matches the target,
// the action will be allowed.
func RequireAuthorizationOrTargetMatch(actor *db.User, requiredAuthzName sqlc.AuthzName, target uuid.UUID) error {
	if actor != nil && actor.ID == target {
		return nil
	}

	return RequireAuthorization(actor, requiredAuthzName, target)
}

// authorized Check if one of the authorizations in the map matches the required authorization.
func authorized(authorizations map[sqlc.AuthzName]struct{}, requiredAuthzName sqlc.AuthzName) error {
	for authorization := range authorizations {
		if authorization == requiredAuthzName {
			return nil
		}
	}

	return ErrNotAuthorized
}
