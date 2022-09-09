package graph_test

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/nais/console/pkg/auditlogger"
	"github.com/nais/console/pkg/authz"
	"github.com/nais/console/pkg/db"
	"github.com/nais/console/pkg/graph"
	"github.com/nais/console/pkg/reconcilers"
	"github.com/nais/console/pkg/sqlc"
	"github.com/stretchr/testify/assert"
)

func TestQueryResolver_Users(t *testing.T) {
	ctx := context.Background()
	database := db.NewMockDatabase(t)
	auditLogger := auditlogger.NewMockAuditLogger(t)
	ch := make(chan reconcilers.Input, 100)
	resolver := graph.NewResolver(database, "example.com", ch, auditLogger).Query()

	t.Run("user without required authorization", func(t *testing.T) {
		users, err := resolver.Users(ctx)
		assert.Nil(t, users)
		assert.ErrorIs(t, err, authz.ErrNotAuthorized)
	})

	t.Run("user with authorization", func(t *testing.T) {
		user := &db.User{
			Email: "user@example.com",
			Name:  "User Name",
		}
		ctx := authz.ContextWithActor(ctx, user, []*db.Role{
			{
				UserRole:       &sqlc.UserRole{TargetID: uuid.NullUUID{}},
				Authorizations: []sqlc.AuthzName{sqlc.AuthzNameUsersList},
			},
		})

		database.On("GetUsers", ctx).Return([]*db.User{
			{Email: "user1@example.com"},
			{Email: "user2@example.com"},
		}, nil)

		users, err := resolver.Users(ctx)
		assert.NoError(t, err)
		assert.Len(t, users, 2)
	})
}
