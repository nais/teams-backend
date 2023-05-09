package graph_test

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/nais/console/pkg/auditlogger"
	"github.com/nais/console/pkg/authz"
	"github.com/nais/console/pkg/db"
	"github.com/nais/console/pkg/deployproxy"
	"github.com/nais/console/pkg/graph"
	"github.com/nais/console/pkg/logger"
	"github.com/nais/console/pkg/roles"
	"github.com/nais/console/pkg/sqlc"
	"github.com/nais/console/pkg/usersync"
	"github.com/stretchr/testify/assert"
)

func TestMutationResolver_Roles(t *testing.T) {
	serviceAccount := &db.ServiceAccount{
		ServiceAccount: &sqlc.ServiceAccount{
			ID:   uuid.New(),
			Name: "User Name",
		},
	}
	ctx := authz.ContextWithActor(context.Background(), serviceAccount, []*db.Role{
		{
			RoleName: sqlc.RoleNameAdmin,
			Authorizations: []roles.Authorization{
				roles.AuthorizationTeamsCreate,
			},
		},
	})

	userSyncRuns := usersync.NewRunsHandler(5)
	auditLogger := auditlogger.NewMockAuditLogger(t)
	database := db.NewMockDatabase(t)
	deployProxy := deployproxy.NewMockProxy(t)
	log, err := logger.GetLogger("text", "info")
	assert.NoError(t, err)
	userSync := make(chan<- uuid.UUID)
	resolver := graph.
		NewResolver(nil, database, deployProxy, "example.com", userSync, auditLogger, []string{"env"}, log, userSyncRuns).
		ServiceAccount()

	t.Run("get roles for serviceAccount", func(t *testing.T) {
		role := &db.Role{
			Authorizations:         nil,
			RoleName:               "",
			TargetServiceAccountID: nil,
			TargetTeamSlug:         nil,
		}

		database.On("GetServiceAccountRoles", ctx, serviceAccount.ID).
			Return([]*db.Role{role}, nil)

		roles, err := resolver.Roles(ctx, serviceAccount)
		assert.NoError(t, err)
		assert.Equal(t, roles[0], role)
	})
}
