package authz

import (
	"context"
	"github.com/nais/console/pkg/dbmodels"
)

const userContextKey = "user"

// ContextWithUser Return a context with a user module stored.
func ContextWithUser(ctx context.Context, user *dbmodels.User) context.Context {
	return context.WithValue(ctx, userContextKey, user)
}

// UserFromContext Finds any authenticated user from the context. Requires that a middleware has stored a user in the
// first place.
func UserFromContext(ctx context.Context) *dbmodels.User {
	user, _ := ctx.Value(userContextKey).(*dbmodels.User)
	return user
}
