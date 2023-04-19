package graph_test

import (
	"context"
	"testing"

	"github.com/nais/console/pkg/roles"

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

func TestMutationResolver_Role(t *testing.T) {
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

	auditLogger := auditlogger.NewMockAuditLogger(t)
	database := db.NewMockDatabase(t)
	deployProxy := deployproxy.NewMockProxy(t)
	log, err := logger.GetLogger("text", "info")
	assert.NoError(t, err)
	userSync := make(chan<- uuid.UUID)
	resolver := graph.NewResolver(nil, database, deployProxy, "example.com", userSync, auditLogger, []string{"env"}, log).Role()

	t.Run("get role name", func(t *testing.T) {
		role := &db.Role{
			Authorizations:         nil,
			RoleName:               sqlc.RoleNameAdmin,
			TargetServiceAccountID: nil,
			TargetTeamSlug:         nil,
		}

		roleName, err := resolver.Name(ctx, role)
		assert.NoError(t, err)
		assert.Equal(t, sqlc.RoleNameAdmin, roleName)
	})
}
