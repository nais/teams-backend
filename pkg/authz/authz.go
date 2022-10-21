package authz

import (
	"context"
	"errors"

	"github.com/nais/console/pkg/sqlc"

	"github.com/google/uuid"
	"github.com/nais/console/pkg/db"
)

type ContextKey string

type Actor struct {
	User  db.AuthenticatedUser
	Roles []*db.Role
}

var (
	ErrNotAuthenticated = errors.New("not authenticated")
)

func (u *Actor) Authenticated() bool {
	if u == nil || u.User == nil {
		return false
	}

	return true
}

const contextKeyUser ContextKey = "actor"

// ContextWithActor Return a context with an actor attached to it.
func ContextWithActor(ctx context.Context, user db.AuthenticatedUser, roles []*db.Role) context.Context {
	return context.WithValue(ctx, contextKeyUser, &Actor{
		User:  user,
		Roles: roles,
	})
}

// RequireRole Check if an actor has a required role
func RequireRole(actor *Actor, requiredRoleName sqlc.RoleName) error {
	for _, role := range actor.Roles {
		if role.RoleName == requiredRoleName {
			return nil
		}
	}

	return ErrNotAuthorized{role: string(requiredRoleName)}
}

// ActorFromContext Get the actor stored in the context. Requires that a middleware has stored an actor in the first
// place.
func ActorFromContext(ctx context.Context) *Actor {
	actor, _ := ctx.Value(contextKeyUser).(*Actor)
	return actor
}

// RequireGlobalAuthorization Require an actor to have a specific authorization through a globally assigned role.
func RequireGlobalAuthorization(actor *Actor, requiredAuthzName sqlc.AuthzName) error {
	if !actor.Authenticated() {
		return ErrNotAuthenticated
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
// targeted role.
func RequireAuthorization(actor *Actor, requiredAuthzName sqlc.AuthzName, target uuid.UUID) error {
	if !actor.Authenticated() {
		return ErrNotAuthenticated
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
// correctly targeted role. If the actor matches the target the action will be allowed.
func RequireAuthorizationOrTargetMatch(actor *Actor, requiredAuthzName sqlc.AuthzName, target uuid.UUID) error {
	if !actor.Authenticated() {
		return ErrNotAuthenticated
	}

	if actor.User.GetID() == target {
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

	return ErrNotAuthorized{role: string(requiredAuthzName)}
}
