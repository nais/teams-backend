package auth

import (
	"context"

	"github.com/nais/console/pkg/dbmodels"
)

var userCtxKey = &contextKey{"user"}

type contextKey struct {
	name string
}

// Finds any authenticated user from the context. Requires that a middleware authenticator has set the user.
func UserFromContext(ctx context.Context) *dbmodels.User {
	user, _ := ctx.Value(userCtxKey).(*dbmodels.User)
	return user
}

// Insert a user object into a context.
func ContextWithUser(ctx context.Context, user *dbmodels.User) context.Context {
	return context.WithValue(ctx, userCtxKey, user)
}
