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
			Authorizations: []sqlc.AuthzName{
				sqlc.AuthzNameTeamsCreate,
			},
		},
	})

	reconcilerQueue := teamsync.NewMockQueue(t)
	auditLogger := auditlogger.NewMockAuditLogger(t)
	database := db.NewMockDatabase(t)
	deployProxy := deployproxy.NewMockProxy(t)
	log, err := logger.GetLogger("text", "info")
	assert.NoError(t, err)
	userSync := make(chan<- uuid.UUID)
	resolver := graph.NewResolver(nil, database, deployProxy, "example.com", reconcilerQueue, userSync, auditLogger, []string{"env"}, log).ServiceAccount()

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
