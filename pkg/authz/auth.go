package authz

import (
	"context"
	"fmt"
	"github.com/nais/console/pkg/dbmodels"
)

const (
	userContextKey = "user"

	AccessLevelCreate AccessLevel = "C"
	AccessLevelRead   AccessLevel = "R"
	AccessLevelUpdate AccessLevel = "U"
	AccessLevelDelete AccessLevel = "D"

	PermissionAllow = "allow"
	PermissionDeny  = "deny"

	ResourceTeams Resource = "teams"
)

type AccessLevel string
type Resource string

func (r Resource) Format(args ...interface{}) Resource {
	return Resource(fmt.Sprintf(string(r), args...))
}

// ContextWithUser Return a context with a user module stored.
func ContextWithUser(ctx context.Context, user *dbmodels.User) context.Context {
	return context.WithValue(ctx, userContextKey, user)
}

// UserFromContext Finds any authenticated user from the context. Requires that a middleware has stored a user in the first place.
func UserFromContext(ctx context.Context) *dbmodels.User {
	user, _ := ctx.Value(userContextKey).(*dbmodels.User)
	return user
}
