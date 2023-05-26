package dependencytrack

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/nais/teams-backend/pkg/config"
	"github.com/nais/teams-backend/pkg/dependencytrack"

	"github.com/nais/dependencytrack/pkg/client"

	"github.com/google/uuid"
	"github.com/nais/teams-backend/pkg/auditlogger"
	"github.com/nais/teams-backend/pkg/db"
	"github.com/nais/teams-backend/pkg/logger"
	"github.com/nais/teams-backend/pkg/reconcilers"
	"github.com/nais/teams-backend/pkg/slug"
	"github.com/nais/teams-backend/pkg/sqlc"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestNewFromConfig(t *testing.T) {
	log, err := logger.GetLogger("text", "info")
	assert.NoError(t, err)
	audit := auditlogger.NewMockAuditLogger(t)
	audit.On("WithSystemName", sqlc.SystemNameNaisDependencytrack).Return(audit)
	database := db.NewMockDatabase(t)

	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		fmt.Printf("Request: %s %s\n", req.Method, req.URL.String())
		rw.WriteHeader(http.StatusOK)
		_, err = rw.Write([]byte("4.8.0"))
		if err != nil {
			return
		}
	}))

	cfg := &config.Config{
		DependencyTrack: config.DependencyTrack{
			Instances: []dependencytrack.DependencyTrackInstance{
				{
					Endpoint: "https://dependencytrack-backend.dev-gcp.nav.cloud.nais.io",
					Username: "na",
					Password: "na",
				},
				{
					Endpoint: server.URL,
					Username: "na",
					Password: "na",
				},
			},
		},
	}
	_, err = NewFromConfig(context.Background(), database, cfg, audit, log)
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
	audit.On("WithSystemName", sqlc.SystemNameNaisDependencytrack).Return(audit)
	database := db.NewMockDatabase(t)
	mockClient := NewMockClient(t)

	instances := map[string]client.Client{"mock": mockClient}

	ctx := context.Background()

	t.Run("team does not exist, new team created and new members added", func(t *testing.T) {
		database.On("LoadReconcilerStateForTeam", ctx, Name, input.Team.Slug, mock.Anything).Return(nil).Once()
		mockClient.On("CreateTeam", mock.Anything, teamName, []client.Permission{
			client.ViewPortfolioPermission,
		}).Return(&client.Team{
			Uuid:      teamUuid,
			Name:      teamName,
			OidcUsers: nil,
		}, nil).Once()

		audit.
			On("Logf", ctx, database, mock.MatchedBy(func(t []auditlogger.Target) bool {
				return len(t) == 1 && t[0].Identifier == string(input.Team.Slug)
			}), mock.MatchedBy(func(f auditlogger.Fields) bool {
				return f.Action == sqlc.AuditActionDependencytrackGroupCreate && f.CorrelationID == correlationID
			}), mock.Anything, teamName, teamUuid).
			Return(nil).
			Once()

		mockClient.On("CreateOidcUser", mock.Anything, username).Return(&client.User{
			Username: username,
			Email:    username,
		}).Return(nil).Once()

		mockClient.On("AddToTeam", mock.Anything, username, teamUuid).Return(nil).Once()
		database.On("SetReconcilerStateForTeam", ctx, Name, input.Team.Slug, mock.Anything).Return(nil).Once()

		reconciler, err := New(database, audit, instances, log)
		assert.NoError(t, err)

		err = reconciler.Reconcile(context.Background(), input)
		assert.NoError(t, err)
	})

	t.Run("team exists, new members added", func(t *testing.T) {
		database.On("LoadReconcilerStateForTeam", ctx, Name, input.Team.Slug, mock.Anything).Run(func(args mock.Arguments) {
			state := args.Get(3).(*reconcilers.DependencyTrackState)
			state.Instances = []*reconcilers.DependencyTrackInstanceState{
				{
					Endpoint: "mock",
					TeamID:   teamUuid,
					Members:  []string{},
				},
			}
		}).Return(nil).Once()
		mockClient.On("CreateOidcUser", mock.Anything, username).Return(&client.User{
			Username: username,
			Email:    username,
		}).Return(nil).Once()

		mockClient.On("AddToTeam", mock.Anything, username, teamUuid).Return(nil).Once()
		database.On("SetReconcilerStateForTeam", ctx, Name, input.Team.Slug, mock.Anything).Return(nil).Once()

		reconciler, err := New(database, audit, instances, log)
		assert.NoError(t, err)

		err = reconciler.Reconcile(context.Background(), input)
		assert.NoError(t, err)
	})

	t.Run("team exists all input members exists, no new members added", func(t *testing.T) {
		database.On("LoadReconcilerStateForTeam", ctx, Name, input.Team.Slug, mock.Anything).Run(func(args mock.Arguments) {
			state := args.Get(3).(*reconcilers.DependencyTrackState)
			state.Instances = []*reconcilers.DependencyTrackInstanceState{
				{
					Endpoint: "mock",
					TeamID:   teamUuid,
					Members:  []string{username},
				},
			}
		}).Return(nil).Once()
		database.On("SetReconcilerStateForTeam", ctx, Name, input.Team.Slug, mock.Anything).Return(nil).Once()

		reconciler, err := New(database, audit, instances, log)
		assert.NoError(t, err)

		err = reconciler.Reconcile(context.Background(), input)
		assert.NoError(t, err)
	})

	t.Run("usermembership removed from existing team", func(t *testing.T) {
		// usernameInInput := "user1@nais.io"
		usernameNotInInput := "userNotInConsole@nais.io"
		database.On("LoadReconcilerStateForTeam", ctx, Name, input.Team.Slug, mock.Anything).Run(func(args mock.Arguments) {
			state := args.Get(3).(*reconcilers.DependencyTrackState)
			state.Instances = []*reconcilers.DependencyTrackInstanceState{
				{
					Endpoint: "mock",
					TeamID:   teamUuid,
					Members:  []string{usernameNotInInput},
				},
			}
		}).Return(nil).Once()

		mockClient.On("CreateOidcUser", mock.Anything, username).Return(&client.User{
			Username: username,
			Email:    username,
		}).Return(nil).Once()
		mockClient.On("AddToTeam", mock.Anything, username, teamUuid).Return(nil).Once()
		mockClient.On("DeleteUserMembership", mock.Anything, teamUuid, usernameNotInInput).Return(nil).Once()

		database.On("SetReconcilerStateForTeam", ctx, Name, input.Team.Slug, mock.Anything).Return(nil).Once()

		reconciler, err := New(database, audit, instances, log)
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

	mockClient := NewMockClient(t)
	database := db.NewMockDatabase(t)
	auditLogger := auditlogger.NewMockAuditLogger(t)
	auditLogger.On("WithSystemName", sqlc.SystemNameNaisDependencytrack).Return(auditLogger)

	t.Run("team exists, delete team from console should remove team from dependencytrack", func(t *testing.T) {
		teamUuid := uuid.New().String()

		database.On("LoadReconcilerStateForTeam", ctx, Name, input.Team.Slug, mock.Anything).Run(func(args mock.Arguments) {
			state := args.Get(3).(*reconcilers.DependencyTrackState)
			state.Instances = []*reconcilers.DependencyTrackInstanceState{
				{
					Endpoint: "mock",
					TeamID:   teamUuid,
					Members:  []string{},
				},
			}
		}).Return(nil).Once()

		mockClient.On("DeleteTeam", mock.Anything, teamUuid).Return(nil).Once()
		database.On("RemoveReconcilerStateForTeam", ctx, Name, input.Team.Slug, mock.Anything).Return(nil).Once()

		reconciler, err := New(database, auditLogger, map[string]client.Client{"mock": mockClient}, log)
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
