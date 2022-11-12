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

	reconcilers := make(chan reconcilers.Input, 100)
	auditLogger := auditlogger.NewMockAuditLogger(t)
	database := db.NewMockDatabase(t)
	resolver := graph.NewResolver(database, "example.com", reconcilers, auditLogger, []string{"env"}).ServiceAccount()

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
