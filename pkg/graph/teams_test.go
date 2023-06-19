package graph_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/nais/teams-backend/pkg/types"

	"github.com/google/uuid"
	"github.com/nais/teams-backend/pkg/auditlogger"
	"github.com/nais/teams-backend/pkg/authz"
	"github.com/nais/teams-backend/pkg/db"
	"github.com/nais/teams-backend/pkg/deployproxy"
	"github.com/nais/teams-backend/pkg/graph"
	"github.com/nais/teams-backend/pkg/graph/apierror"
	"github.com/nais/teams-backend/pkg/graph/model"
	"github.com/nais/teams-backend/pkg/logger"
	"github.com/nais/teams-backend/pkg/roles"
	"github.com/nais/teams-backend/pkg/slug"
	"github.com/nais/teams-backend/pkg/sqlc"
	"github.com/nais/teams-backend/pkg/teamsync"
	"github.com/nais/teams-backend/pkg/usersync"
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
			Authorizations: []roles.Authorization{
				roles.AuthorizationTeamsCreate,
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
			Authorizations: []roles.Authorization{
				roles.AuthorizationTeamsCreate,
			},
		},
	})

	userSyncRuns := usersync.NewRunsHandler(5)
	database := db.NewMockDatabase(t)
	teamSyncHandler := teamsync.NewMockHandler(t)
	auditLogger := auditlogger.NewMockAuditLogger(t)

	deployProxy := deployproxy.NewMockProxy(t)
	gcpEnvironments := []string{"env"}
	log, err := logger.GetLogger("text", "info")
	assert.NoError(t, err)
	userSync := make(chan<- uuid.UUID)
	teamSlug := slug.Slug("some-slug")
	slackChannel := "#my-slack-channel"
	resolver := graph.
		NewResolver(teamSyncHandler, database, deployProxy, "example.com", userSync, auditLogger, gcpEnvironments, log, userSyncRuns).
		Mutation()

	t.Run("create team with empty purpose", func(t *testing.T) {
		_, err := resolver.CreateTeam(ctx, model.CreateTeamInput{
			Slug:         &teamSlug,
			Purpose:      "  ",
			SlackChannel: slackChannel,
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
			On("CreateTeam", txCtx, teamSlug, "some purpose", slackChannel).
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

		auditLogger.
			On("Logf", ctx, database, mock.MatchedBy(func(targets []auditlogger.Target) bool {
				return targets[0].Identifier == string(createdTeam.Slug)
			}), mock.MatchedBy(func(fields auditlogger.Fields) bool {
				return fields.Actor.User == user
			}), "Team created").
			Return(nil).
			Once()

		teamSyncHandler.
			On("Schedule", mock.MatchedBy(func(input teamsync.Input) bool {
				return input.TeamSlug == createdTeam.Slug
			})).
			Return(nil).
			Once()

		returnedTeam, err := resolver.CreateTeam(ctx, model.CreateTeamInput{
			Slug:         &teamSlug,
			Purpose:      " some purpose ",
			SlackChannel: slackChannel,
		})
		assert.NoError(t, err)
		assert.Equal(t, createdTeam.Slug, returnedTeam.Slug)
	})

	t.Run("calling with SA, adds sa as team owner", func(t *testing.T) {
		createdTeam := &db.Team{
			Team: &sqlc.Team{Slug: teamSlug},
		}
		txCtx := context.Background()
		dbtx := db.NewMockDatabase(t)

		dbtx.
			On("CreateTeam", txCtx, teamSlug, "some purpose", slackChannel).
			Return(createdTeam, nil).
			Once()

		dbtx.
			On("AssignTeamRoleToServiceAccount", txCtx, serviceAccount.GetID(), sqlc.RoleNameTeamowner, teamSlug).
			Return(nil).
			Once()

		database.
			On("Transaction", saCtx, mock.Anything).
			Run(func(args mock.Arguments) {
				fn := args.Get(1).(db.DatabaseTransactionFunc)
				fn(txCtx, dbtx)
			}).
			Return(nil).
			Once()

		auditLogger.
			On("Logf", saCtx, database, mock.MatchedBy(func(targets []auditlogger.Target) bool {
				return targets[0].Identifier == string(createdTeam.Slug)
			}), mock.MatchedBy(func(fields auditlogger.Fields) bool {
				return fields.Actor.User == serviceAccount
			}), "Team created").
			Return(nil).
			Once()

		teamSyncHandler.
			On("Schedule", mock.MatchedBy(func(input teamsync.Input) bool {
				return input.TeamSlug == createdTeam.Slug
			})).
			Return(nil).
			Once()

		returnedTeam, err := resolver.CreateTeam(saCtx, model.CreateTeamInput{
			Slug:         &teamSlug,
			Purpose:      " some purpose ",
			SlackChannel: slackChannel,
		})

		assert.NoError(t, err)
		assert.Equal(t, createdTeam.Slug, returnedTeam.Slug)
	})
}

func TestMutationResolver_RequestTeamDeletion(t *testing.T) {
	const tenantDomain = "example.com"
	teamSyncHandler := teamsync.NewMockHandler(t)
	database := db.NewMockDatabase(t)
	deployProxy := deployproxy.NewMockProxy(t)
	auditLogger := auditlogger.NewMockAuditLogger(t)
	log := logger.NewMockLogger(t)
	log.
		On("WithComponent", types.ComponentNameGraphqlApi).
		Return(log)
	userSync := make(chan<- uuid.UUID)
	gcpEnvironments := []string{"env"}
	ctx := context.Background()
	teamSlug := slug.Slug("my-team")
	userSyncRuns := usersync.NewRunsHandler(5)

	t.Run("service accounts can not create delete keys", func(t *testing.T) {
		resolver := graph.
			NewResolver(teamSyncHandler, database, deployProxy, tenantDomain, userSync, auditLogger, gcpEnvironments, log, userSyncRuns).
			Mutation()

		serviceAccount := db.ServiceAccount{
			ServiceAccount: &sqlc.ServiceAccount{
				ID:   uuid.New(),
				Name: "service-account",
			},
		}

		ctx := authz.ContextWithActor(ctx, serviceAccount, []*db.Role{})
		key, err := resolver.RequestTeamDeletion(ctx, &teamSlug)
		assert.Nil(t, key)
		assert.ErrorContains(t, err, "Service accounts are not allowed")
	})

	t.Run("missing authz", func(t *testing.T) {
		resolver := graph.
			NewResolver(teamSyncHandler, database, deployProxy, tenantDomain, userSync, auditLogger, gcpEnvironments, log, userSyncRuns).
			Mutation()

		user := db.User{
			User: &sqlc.User{
				ID:    uuid.New(),
				Email: "user@example.com",
				Name:  "User Name",
			},
		}
		ctx := authz.ContextWithActor(ctx, user, []*db.Role{})

		key, err := resolver.RequestTeamDeletion(ctx, &teamSlug)
		assert.Nil(t, key)
		assert.ErrorContains(t, err, "required authorization")
	})

	t.Run("missing team", func(t *testing.T) {
		user := db.User{
			User: &sqlc.User{
				ID:    uuid.New(),
				Email: "user@example.com",
				Name:  "User Name",
			},
		}
		ctx := authz.ContextWithActor(ctx, user, []*db.Role{
			{
				RoleName: sqlc.RoleNameTeamowner,
				Authorizations: []roles.Authorization{
					roles.AuthorizationTeamsUpdate,
				},
			},
		})

		database := db.NewMockDatabase(t)
		database.
			On("GetTeamBySlug", ctx, teamSlug).
			Return(nil, fmt.Errorf("some error")).
			Once()

		resolver := graph.
			NewResolver(teamSyncHandler, database, deployProxy, tenantDomain, userSync, auditLogger, gcpEnvironments, log, userSyncRuns).
			Mutation()

		key, err := resolver.RequestTeamDeletion(ctx, &teamSlug)
		assert.Nil(t, key)
		assert.ErrorIs(t, err, apierror.ErrTeamNotExist)
	})

	t.Run("create key", func(t *testing.T) {
		userID := uuid.New()
		user := db.User{
			User: &sqlc.User{
				ID:    userID,
				Email: "user@example.com",
				Name:  "User Name",
			},
		}
		team := &db.Team{
			Team: &sqlc.Team{
				Slug:         teamSlug,
				SlackChannel: "#channel",
			},
		}
		ctx := authz.ContextWithActor(ctx, user, []*db.Role{
			{
				RoleName: sqlc.RoleNameTeamowner,
				Authorizations: []roles.Authorization{
					roles.AuthorizationTeamsUpdate,
				},
			},
		})

		key := &db.TeamDeleteKey{
			TeamDeleteKey: &sqlc.TeamDeleteKey{
				Key:         uuid.UUID{},
				TeamSlug:    teamSlug,
				CreatedAt:   time.Time{},
				CreatedBy:   uuid.UUID{},
				ConfirmedAt: nil,
			},
		}

		database := db.NewMockDatabase(t)
		database.
			On("GetTeamBySlug", ctx, teamSlug).
			Return(team, nil).
			Once()
		database.
			On("CreateTeamDeleteKey", ctx, teamSlug, userID).
			Return(key, nil).
			Once()

		auditLogger := auditlogger.NewMockAuditLogger(t)
		auditLogger.
			On(
				"Logf",
				ctx,
				database,
				mock.MatchedBy(func(targets []auditlogger.Target) bool {
					return targets[0].Identifier == string(teamSlug) && targets[0].Type == types.AuditLogsTargetTypeTeam
				}),
				mock.MatchedBy(func(fields auditlogger.Fields) bool {
					return fields.Action == types.AuditActionGraphqlApiTeamsRequestDelete && fields.Actor.User == user
				}),
				mock.AnythingOfType("string"),
			).
			Return(nil).
			Once()

		resolver := graph.
			NewResolver(teamSyncHandler, database, deployProxy, tenantDomain, userSync, auditLogger, gcpEnvironments, log, userSyncRuns).
			Mutation()

		returnedKey, err := resolver.RequestTeamDeletion(ctx, &teamSlug)
		assert.Equal(t, key, returnedKey)
		assert.NoError(t, err)
	})
}
