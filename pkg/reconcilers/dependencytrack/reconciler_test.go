package dependencytrack

import (
	"context"
	"testing"

	"github.com/nais/console/pkg/dependencytrack"

	"github.com/google/uuid"
	"github.com/nais/console/pkg/auditlogger"
	"github.com/nais/console/pkg/db"
	"github.com/nais/console/pkg/logger"
	"github.com/nais/console/pkg/reconcilers"
	"github.com/nais/console/pkg/slug"
	"github.com/nais/console/pkg/sqlc"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestDependencytrackReconciler_Reconcile(t *testing.T) {

	correlationID := uuid.New()
	input := setupInput(correlationID, "someTeam", "user1@nais.io")

	teamName := input.Team.Slug.String()
	teamUuid := uuid.New().String()
	username := input.TeamMembers[0].Email

	log, err := logger.GetLogger("text", "info")
	assert.NoError(t, err)

	auditLogger := auditlogger.NewMockAuditLogger(t)
	database := db.NewMockDatabase(t)
	mockClient := dependencytrack.NewMockClient(t)

	instances := map[string]dependencytrack.Client{"mock": mockClient}

	ctx := context.Background()

	t.Run("team does not exist, new team created and new members added", func(t *testing.T) {
		database.On("LoadReconcilerStateForTeam", ctx, Name, input.Team.Slug, mock.Anything).Return(nil).Once()
		mockClient.On("CreateTeam", mock.Anything, teamName, []dependencytrack.Permission{
			dependencytrack.ViewPortfolioPermission,
		}).Return(&dependencytrack.Team{
			Uuid:      teamUuid,
			Name:      teamName,
			OidcUsers: nil,
		}, nil).Once()

		mockClient.On("CreateUser", mock.Anything, username).Return(&dependencytrack.User{
			Username: username,
			Email:    username,
		}).Return(nil).Once()

		mockClient.On("AddToTeam", mock.Anything, username, teamUuid).Return(nil).Once()
		database.On("SetReconcilerStateForTeam", ctx, Name, input.Team.Slug, mock.Anything).Return(nil).Once()

		reconciler, err := New(database, auditLogger, instances, log)
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
		mockClient.On("CreateUser", mock.Anything, username).Return(&dependencytrack.User{
			Username: username,
			Email:    username,
		}).Return(nil).Once()

		mockClient.On("AddToTeam", mock.Anything, username, teamUuid).Return(nil).Once()
		database.On("SetReconcilerStateForTeam", ctx, Name, input.Team.Slug, mock.Anything).Return(nil).Once()

		reconciler, err := New(database, auditLogger, instances, log)
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

		reconciler, err := New(database, auditLogger, instances, log)
		assert.NoError(t, err)

		err = reconciler.Reconcile(context.Background(), input)
		assert.NoError(t, err)
	})

	t.Run("usermembership removed from existing team", func(t *testing.T) {

		//usernameInInput := "user1@nais.io"
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

		mockClient.On("CreateUser", mock.Anything, username).Return(&dependencytrack.User{
			Username: username,
			Email:    username,
		}).Return(nil).Once()
		mockClient.On("AddToTeam", mock.Anything, username, teamUuid).Return(nil).Once()
		mockClient.On("DeleteUserMembership", mock.Anything, teamUuid, usernameNotInInput).Return(nil).Once()

		database.On("SetReconcilerStateForTeam", ctx, Name, input.Team.Slug, mock.Anything).Return(nil).Once()

		reconciler, err := New(database, auditLogger, instances, log)
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

	mockClient := dependencytrack.NewMockClient(t)
	database := db.NewMockDatabase(t)
	auditLogger := auditlogger.NewMockAuditLogger(t)

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

		reconciler, err := New(database, auditLogger, map[string]dependencytrack.Client{"mock": mockClient}, log)
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
		CorrelationID:   correlationId,
		Team:            inputTeam,
		TeamMembers:     inputMembers,
		NumSyncAttempts: 0,
	}
}
