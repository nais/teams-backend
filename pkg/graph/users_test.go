package graph_test

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/nais/teams-backend/pkg/auditlogger"
	"github.com/nais/teams-backend/pkg/authz"
	"github.com/nais/teams-backend/pkg/db"
	"github.com/nais/teams-backend/pkg/deployproxy"
	"github.com/nais/teams-backend/pkg/graph"
	"github.com/nais/teams-backend/pkg/logger"
	"github.com/nais/teams-backend/pkg/roles"
	"github.com/nais/teams-backend/pkg/sqlc"
	"github.com/nais/teams-backend/pkg/usersync"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestQueryResolver_Users(t *testing.T) {
	ctx := context.Background()
	database := db.NewMockDatabase(t)
	deployProxy := deployproxy.NewMockProxy(t)
	auditLogger := auditlogger.NewMockAuditLogger(t)
	gcpEnvironments := []string{"env"}
	log, err := logger.GetLogger("text", "info")
	assert.NoError(t, err)
	userSync := make(chan<- uuid.UUID)
	userSyncRuns := usersync.NewRunsHandler(5)
	resolver := graph.
		NewResolver(nil, database, deployProxy, "example.com", userSync, auditLogger, gcpEnvironments, log, userSyncRuns).
		Query()

	t.Run("unauthenticated user", func(t *testing.T) {
		users, err := resolver.Users(ctx, nil, nil)
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
				Authorizations: []roles.Authorization{roles.AuthorizationUsersList},
			},
		})

		database.On("GetUsers", ctx, mock.Anything, mock.Anything).Return([]*db.User{
			{User: &sqlc.User{Email: "user1@example.com"}},
			{User: &sqlc.User{Email: "user2@example.com"}},
		}, nil)

		users, err := resolver.Users(ctx, nil, nil)
		assert.NoError(t, err)
		assert.Len(t, users, 2)
	})
}
