package dependencytrack_reconciler_test

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/nais/teams-backend/pkg/types"

	"github.com/google/uuid"
	"github.com/nais/dependencytrack/pkg/client"
	"github.com/nais/teams-backend/pkg/auditlogger"
	"github.com/nais/teams-backend/pkg/config"
	"github.com/nais/teams-backend/pkg/db"
	"github.com/nais/teams-backend/pkg/logger"
	"github.com/nais/teams-backend/pkg/reconcilers"
	dependencytrackReconciler "github.com/nais/teams-backend/pkg/reconcilers/dependencytrack"
	"github.com/nais/teams-backend/pkg/slug"
	"github.com/nais/teams-backend/pkg/sqlc"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestNewFromConfig(t *testing.T) {
	log, err := logger.GetLogger("text", "info")
	assert.NoError(t, err)
	database := db.NewMockDatabase(t)

	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		fmt.Printf("Request: %s %s\n", req.Method, req.URL.String())
		rw.WriteHeader(http.StatusOK)
		_, err = rw.Write([]byte("4.8.0"))
		assert.NoError(t, err)
	}))

	cfg := &config.Config{
		DependencyTrack: config.DependencyTrack{
			Endpoint: server.URL,
			Username: "na",
			Password: "na",
		},
	}
	_, err = dependencytrackReconciler.NewFromConfig(context.Background(), database, cfg, log)
	assert.NoError(t, err)
}

func TestDependencytrackReconciler_Reconcile(t *testing.T) {
	correlationID := uuid.New()
	input := setupInput(correlationID, "someTeam", "user1@nais.io")

	teamName := input.Team.Slug.String()
	teamUuid := uuid.New().String()
	username := input.TeamMembers[0].Email

	log, err := logger.GetLogger("text", "info")
	assert.NoError(t, err)

	audit := auditlogger.NewMockAuditLogger(t)
	database := db.NewMockDatabase(t)
	mockClient := dependencytrackReconciler.NewMockClient(t)

	dpTrack := dependencytrackReconciler.DpTrack{
		Endpoint: "mock",
		Client:   mockClient,
	}

	ctx := context.Background()

	t.Run("team does not exist, new team created and new members added", func(t *testing.T) {
		database.On("LoadReconcilerStateForTeam", ctx, dependencytrackReconciler.Name, input.Team.Slug, mock.Anything).Return(nil).Once()
		mockClient.On("CreateTeam", mock.Anything, teamName, []client.Permission{
			client.ViewPortfolioPermission,
			client.ViewVulnerabilityPermission,
			client.ViewPolicyViolationPermission,
		}).Return(&client.Team{
			Uuid:      teamUuid,
			Name:      teamName,
			OidcUsers: nil,
		}, nil).Once()

		audit.EXPECT().
			Logf(ctx, mock.MatchedBy(func(t []auditlogger.Target) bool {
				return len(t) == 1 && t[0].Identifier == string(input.Team.Slug)
			}), mock.MatchedBy(func(f auditlogger.Fields) bool {
				return f.Action == types.AuditActionDependencytrackTeamCreate && f.CorrelationID == correlationID
			}), mock.Anything, teamName, teamUuid).
			Return().
			Once()

		audit.EXPECT().
			Logf(ctx, mock.MatchedBy(func(t []auditlogger.Target) bool {
				return len(t) == 2 && t[0].Identifier == string(input.Team.Slug) && t[1].Identifier == "user1@nais.io"
			}), mock.MatchedBy(func(f auditlogger.Fields) bool {
				return f.Action == types.AuditActionDependencytrackTeamAddMember && f.CorrelationID == correlationID
			}), mock.Anything, "user1@nais.io", input.Team.Slug).
			Return().
			Once()

		mockClient.On("CreateOidcUser", mock.Anything, username).Return(&client.User{
			Username: username,
			Email:    username,
		}).Return(nil).Once()

		mockClient.On("AddToTeam", mock.Anything, username, teamUuid).Return(nil).Once()
		database.On("SetReconcilerStateForTeam", ctx, dependencytrackReconciler.Name, input.Team.Slug, mock.Anything).Return(nil).Once()

		reconciler, err := dependencytrackReconciler.New(database, audit, dpTrack, log)
		assert.NoError(t, err)

		err = reconciler.Reconcile(context.Background(), input)
		assert.NoError(t, err)
	})

	t.Run("team exists, new members added", func(t *testing.T) {
		audit.EXPECT().
			Logf(ctx, mock.MatchedBy(func(t []auditlogger.Target) bool {
				return len(t) == 2 && t[0].Identifier == string(input.Team.Slug) && t[1].Identifier == "user1@nais.io"
			}), mock.MatchedBy(func(f auditlogger.Fields) bool {
				return f.Action == types.AuditActionDependencytrackTeamAddMember && f.CorrelationID == correlationID
			}), mock.Anything, "user1@nais.io", input.Team.Slug).
			Return().
			Once()

		database.On("LoadReconcilerStateForTeam", ctx, dependencytrackReconciler.Name, input.Team.Slug, mock.Anything).Run(func(args mock.Arguments) {
			state := args.Get(3).(*reconcilers.DependencyTrackState)
			state.TeamID = teamUuid
			state.Members = []string{}
		}).Return(nil).Once()
		mockClient.On("CreateOidcUser", mock.Anything, username).Return(&client.User{
			Username: username,
			Email:    username,
		}).Return(nil).Once()

		mockClient.On("AddToTeam", mock.Anything, username, teamUuid).Return(nil).Once()
		database.On("SetReconcilerStateForTeam", ctx, dependencytrackReconciler.Name, input.Team.Slug, mock.Anything).Return(nil).Once()

		reconciler, err := dependencytrackReconciler.New(database, audit, dpTrack, log)
		assert.NoError(t, err)

		err = reconciler.Reconcile(context.Background(), input)
		assert.NoError(t, err)
	})

	t.Run("team exists all input members exists, no new members added", func(t *testing.T) {
		database.On("LoadReconcilerStateForTeam", ctx, dependencytrackReconciler.Name, input.Team.Slug, mock.Anything).Run(func(args mock.Arguments) {
			state := args.Get(3).(*reconcilers.DependencyTrackState)
			state.TeamID = teamUuid
			state.Members = []string{username}
		}).Return(nil).Once()
		database.On("SetReconcilerStateForTeam", ctx, dependencytrackReconciler.Name, input.Team.Slug, mock.Anything).Return(nil).Once()

		reconciler, err := dependencytrackReconciler.New(database, audit, dpTrack, log)
		assert.NoError(t, err)

		err = reconciler.Reconcile(context.Background(), input)
		assert.NoError(t, err)
	})

	t.Run("usermembership removed from existing team", func(t *testing.T) {
		usernameNotInInput := "userNotInTeamsBackend@nais.io"

		audit.EXPECT().
			Logf(ctx, mock.MatchedBy(func(t []auditlogger.Target) bool {
				return len(t) == 2 && t[0].Identifier == string(input.Team.Slug) && t[1].Identifier == "user1@nais.io"
			}), mock.MatchedBy(func(f auditlogger.Fields) bool {
				return f.Action == types.AuditActionDependencytrackTeamAddMember && f.CorrelationID == correlationID
			}), mock.Anything, "user1@nais.io", input.Team.Slug).
			Return().
			Once()

		audit.EXPECT().
			Logf(ctx, mock.MatchedBy(func(t []auditlogger.Target) bool {
				return len(t) == 2 && t[0].Identifier == string(input.Team.Slug) && t[1].Identifier == usernameNotInInput
			}), mock.MatchedBy(func(f auditlogger.Fields) bool {
				return f.Action == types.AuditActionDependencytrackTeamDeleteMember && f.CorrelationID == correlationID
			}), mock.Anything, usernameNotInInput, input.Team.Slug).
			Return().
			Once()

		database.On("LoadReconcilerStateForTeam", ctx, dependencytrackReconciler.Name, input.Team.Slug, mock.Anything).Run(func(args mock.Arguments) {
			state := args.Get(3).(*reconcilers.DependencyTrackState)
			state.TeamID = teamUuid
			state.Members = []string{usernameNotInInput}
		}).Return(nil).Once()

		mockClient.On("CreateOidcUser", mock.Anything, username).Return(&client.User{
			Username: username,
			Email:    username,
		}).Return(nil).Once()
		mockClient.On("AddToTeam", mock.Anything, username, teamUuid).Return(nil).Once()
		mockClient.On("DeleteUserMembership", mock.Anything, teamUuid, usernameNotInInput).Return(nil).Once()

		database.On("SetReconcilerStateForTeam", ctx, dependencytrackReconciler.Name, input.Team.Slug, mock.Anything).Return(nil).Once()

		reconciler, err := dependencytrackReconciler.New(database, audit, dpTrack, log)
		assert.NoError(t, err)

		err = reconciler.Reconcile(context.Background(), input)
		assert.NoError(t, err)
	})
}

func TestDependencytrackReconciler_Delete(t *testing.T) {
	log, err := logger.GetLogger("text", "info")
	assert.NoError(t, err)
	ctx := context.Background()
	correlationID := uuid.New()
	input := setupInput(correlationID, "someTeam", "user1@nais.io")

	mockClient := dependencytrackReconciler.NewMockClient(t)
	database := db.NewMockDatabase(t)
	auditLogger := auditlogger.NewMockAuditLogger(t)

	dpTrack := dependencytrackReconciler.DpTrack{
		Endpoint: "mock",
		Client:   mockClient,
	}

	t.Run("team exists, delete team from teams-backend should remove team from dependencytrack", func(t *testing.T) {
		teamUuid := uuid.New().String()

		database.On("LoadReconcilerStateForTeam", ctx, dependencytrackReconciler.Name, input.Team.Slug, mock.Anything).Run(func(args mock.Arguments) {
			state := args.Get(3).(*reconcilers.DependencyTrackState)
			state.TeamID = teamUuid
			state.Members = []string{}
		}).Return(nil).Once()

		mockClient.On("DeleteTeam", mock.Anything, teamUuid).Return(nil).Once()
		database.On("RemoveReconcilerStateForTeam", ctx, dependencytrackReconciler.Name, input.Team.Slug, mock.Anything).Return(nil).Once()

		reconciler, err := dependencytrackReconciler.New(database, auditLogger, dpTrack, log)
		assert.NoError(t, err)

		err = reconciler.Delete(context.Background(), input.Team.Slug, uuid.New())
		assert.NoError(t, err)
	})
}

func setupInput(correlationId uuid.UUID, teamSlug string, members ...string) reconcilers.Input {
	inputTeam := db.Team{
		Team: &sqlc.Team{
			Slug:    slug.Slug(teamSlug),
			Purpose: "teamPurpose",
		},
	}

	inputMembers := make([]*db.User, 0)
	for _, member := range members {
		inputMembers = append(inputMembers, &db.User{
			User: &sqlc.User{
				Email: member,
			},
		})
	}

	return reconcilers.Input{
		CorrelationID: correlationId,
		Team:          inputTeam,
		TeamMembers:   inputMembers,
	}
}
