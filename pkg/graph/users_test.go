package graph_test

import (
	"context"
	"testing"

	"github.com/nais/console/pkg/teamsync"

	"github.com/google/uuid"
	"github.com/nais/console/pkg/auditlogger"
	"github.com/nais/console/pkg/authz"
	"github.com/nais/console/pkg/db"
	"github.com/nais/console/pkg/deployproxy"
	"github.com/nais/console/pkg/graph"
	"github.com/nais/console/pkg/logger"
	"github.com/nais/console/pkg/sqlc"
	"github.com/stretchr/testify/assert"
)

func TestQueryResolver_Users(t *testing.T) {
	ctx := context.Background()
	database := db.NewMockDatabase(t)
	deployProxy := deployproxy.NewMockProxy(t)
	auditLogger := auditlogger.NewMockAuditLogger(t)
	reconcilerQueue := teamsync.NewMockQueue(t)
	gcpEnvironments := []string{"env"}
	log, err := logger.GetLogger("text", "info")
	assert.NoError(t, err)
	userSync := make(chan<- uuid.UUID)
	resolver := graph.NewResolver(nil, database, deployProxy, "example.com", reconcilerQueue, userSync, auditLogger, gcpEnvironments, log).Query()

	t.Run("unauthenticated user", func(t *testing.T) {
		users, err := resolver.Users(ctx)
		assert.Nil(t, users)
		assert.ErrorIs(t, err, authz.ErrNotAuthenticated)
	})

	t.Run("user with authorization", func(t *testing.T) {
		user := &db.User{
			User: &sqlc.User{
				Email: "user@example.com",
				Name:  "User Name",
			},
		}
		ctx := authz.ContextWithActor(ctx, user, []*db.Role{
			{
				Authorizations: []sqlc.AuthzName{sqlc.AuthzNameUsersList},
			},
		})

		database.On("GetUsers", ctx).Return([]*db.User{
			{User: &sqlc.User{Email: "user1@example.com"}},
			{User: &sqlc.User{Email: "user2@example.com"}},
		}, nil)

		users, err := resolver.Users(ctx)
		assert.NoError(t, err)
		assert.Len(t, users, 2)
	})
}
