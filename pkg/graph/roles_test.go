package graph_test

import (
	"context"
	"github.com/google/uuid"
	"testing"

	"github.com/nais/console/pkg/auditlogger"
	"github.com/nais/console/pkg/authz"
	"github.com/nais/console/pkg/db"
	"github.com/nais/console/pkg/graph"
	"github.com/nais/console/pkg/reconcilers"
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
			Authorizations: []sqlc.AuthzName{
				sqlc.AuthzNameTeamsCreate,
			},
		},
	})

	reconcilers := make(chan reconcilers.Input, 100)
	auditLogger := auditlogger.NewMockAuditLogger(t)
	database := db.NewMockDatabase(t)
	resolver := graph.NewResolver(database, "example.com", reconcilers, auditLogger, []string{"env"}).Role()

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
