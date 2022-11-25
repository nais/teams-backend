package graph_test

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/nais/console/pkg/auditlogger"
	"github.com/nais/console/pkg/authz"
	"github.com/nais/console/pkg/db"
	"github.com/nais/console/pkg/graph"
	"github.com/nais/console/pkg/graph/model"
	"github.com/nais/console/pkg/logger"
	"github.com/nais/console/pkg/reconcilers"
	"github.com/nais/console/pkg/slug"
	"github.com/nais/console/pkg/sqlc"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestMutationResolver_CreateTeam(t *testing.T) {
	user := db.User{
		User: &sqlc.User{
			ID:    uuid.New(),
			Email: "user@example.com",
			Name:  "User Name",
		},
	}

	ctx := authz.ContextWithActor(context.Background(), user, []*db.Role{
		{
			RoleName: sqlc.RoleNameAdmin,
			Authorizations: []sqlc.AuthzName{
				sqlc.AuthzNameTeamsCreate,
			},
		},
	})

	serviceAccount := db.ServiceAccount{
		ServiceAccount: &sqlc.ServiceAccount{
			ID:   uuid.New(),
			Name: "User Name",
		},
	}

	saCtx := authz.ContextWithActor(context.Background(), serviceAccount, []*db.Role{
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
	gcpEnvironments := []string{"env"}
	log, err := logger.GetLogger("text", "info")
	assert.NoError(t, err)
	resolver := graph.NewResolver(database, "example.com", reconcilers, auditLogger, gcpEnvironments, log).Mutation()
	teamSlug := slug.Slug("some-slug")

	t.Run("create team with empty purpose", func(t *testing.T) {
		_, err := resolver.CreateTeam(ctx, model.CreateTeamInput{
			Slug:    &teamSlug,
			Purpose: "  ",
		})
		assert.ErrorContains(t, err, "You must specify the purpose for your team")
	})

	t.Run("create team", func(t *testing.T) {
		createdTeam := &db.Team{
			Team: &sqlc.Team{Slug: teamSlug},
		}
		txCtx := context.Background()
		dbtx := db.NewMockDatabase(t)

		dbtx.
			On("CreateTeam", txCtx, teamSlug, "some purpose").
			Return(createdTeam, nil).
			Once()
		dbtx.
			On("SetTeamMemberRole", txCtx, user.ID, createdTeam.Slug, sqlc.RoleNameTeamowner).
			Return(nil).
			Once()

		database.
			On("Transaction", ctx, mock.Anything).
			Run(func(args mock.Arguments) {
				fn := args.Get(1).(db.DatabaseTransactionFunc)
				fn(txCtx, dbtx)
			}).
			Return(nil).
			Once()
		database.
			On("GetTeamMembers", ctx, createdTeam.Slug).
			Return([]*db.User{&user}, nil).
			Once()

		auditLogger.
			On("Logf", ctx, mock.MatchedBy(func(targets []auditlogger.Target) bool {
				return targets[0].Identifier == string(createdTeam.Slug)
			}), mock.MatchedBy(func(fields auditlogger.Fields) bool {
				return fields.Actor.User == user
			}), "Team created").
			Return(nil).
			Once()

		returnedTeam, err := resolver.CreateTeam(ctx, model.CreateTeamInput{
			Slug:    &teamSlug,
			Purpose: " some purpose ",
		})
		assert.NoError(t, err)
		assert.Equal(t, createdTeam.Slug, returnedTeam.Slug)

		input := <-reconcilers
		assert.Equal(t, createdTeam.Slug, input.Team.Slug)
	})

	t.Run("calling with SA, does not change roles", func(t *testing.T) {
		createdTeam := &db.Team{
			Team: &sqlc.Team{Slug: teamSlug},
		}
		txCtx := context.Background()
		dbtx := db.NewMockDatabase(t)

		dbtx.
			On("CreateTeam", txCtx, teamSlug, "some purpose").
			Return(createdTeam, nil).
			Once()

		database.
			On("Transaction", saCtx, mock.Anything).
			Run(func(args mock.Arguments) {
				fn := args.Get(1).(db.DatabaseTransactionFunc)
				fn(txCtx, dbtx)
			}).
			Return(nil).
			Once()
		database.
			On("GetTeamMembers", saCtx, createdTeam.Slug).
			Return([]*db.User{&user}, nil).
			Once()

		auditLogger.
			On("Logf", saCtx, mock.MatchedBy(func(targets []auditlogger.Target) bool {
				return targets[0].Identifier == string(createdTeam.Slug)
			}), mock.MatchedBy(func(fields auditlogger.Fields) bool {
				return fields.Actor.User == serviceAccount
			}), "Team created").
			Return(nil).
			Once()

		returnedTeam, err := resolver.CreateTeam(saCtx, model.CreateTeamInput{
			Slug:    &teamSlug,
			Purpose: " some purpose ",
		})

		assert.NoError(t, err)
		assert.Equal(t, createdTeam.Slug, returnedTeam.Slug)

		input := <-reconcilers
		assert.Equal(t, createdTeam.Slug, input.Team.Slug)
	})
}
