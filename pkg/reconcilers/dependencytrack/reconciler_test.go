package dependencytrack

import (
	"context"
	"github.com/nais/console/pkg/dependencytrack"
	"testing"

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

	ctx := context.Background()

	for _, tt := range []struct {
		name   string
		preRun func(t *testing.T, mock *dependencytrack.MockClient)
	}{
		{
			name: "team does not exist, new team created and new members added",
			preRun: func(t *testing.T, client *dependencytrack.MockClient) {

				client.On("GetTeams", mock.Anything).Return([]dependencytrack.Team{}, nil).Once()
				client.On("CreateTeam", mock.Anything, teamName, []dependencytrack.Permission{
					dependencytrack.ViewPortfolioPermission,
				}).Return(&dependencytrack.Team{
					Uuid:      teamUuid,
					Name:      teamName,
					OidcUsers: nil,
				}, nil).Once()

				auditLogger.
					On("Logf", ctx, database, mock.MatchedBy(func(t []auditlogger.Target) bool {
						return len(t) == 1 && t[0].Identifier == string(input.Team.Slug)
					}), mock.MatchedBy(func(f auditlogger.Fields) bool {
						return f.Action == AuditActionDependencytrackCreate && f.CorrelationID == correlationID
					}), mock.Anything, teamName, teamUuid).
					Return(nil).
					Once()

				client.On("CreateUser", mock.Anything, username).Return(&dependencytrack.User{
					Username: username,
					Email:    username,
				}).Return(nil).Once()

				client.On("AddToTeam", mock.Anything, username, teamUuid).Return(nil).Once()
			},
		},
		{
			name: "team exists, new members added",
			preRun: func(t *testing.T, client *dependencytrack.MockClient) {

				client.On("GetTeams", mock.Anything).Return([]dependencytrack.Team{
					{
						Name: teamName,
						Uuid: teamUuid,
					},
				}, nil).Once()

				client.On("CreateUser", mock.Anything, username).Return(&dependencytrack.User{
					Username: username,
					Email:    username,
				}).Return(nil).Once()

				client.On("AddToTeam", mock.Anything, username, teamUuid).Return(nil).Once()
			},
		},
		{
			name: "usermembership removed from existing team",
			preRun: func(t *testing.T, client *dependencytrack.MockClient) {
				usernameInConsole := "user1@nais.io"
				usernameNotInConsole := "userNotInConsole@nais.io"

				client.On("GetTeams", mock.Anything).Return([]dependencytrack.Team{
					{
						Name: teamName,
						Uuid: teamUuid,
						OidcUsers: []dependencytrack.User{
							{
								Username: usernameInConsole,
								Email:    usernameInConsole,
							},
							{
								Username: usernameNotInConsole,
								Email:    usernameNotInConsole,
							},
						},
					},
				}, nil).Once()

				client.On("DeleteUserMembership", mock.Anything, teamUuid, usernameNotInConsole).Return(nil).Once()

				client.On("CreateUser", mock.Anything, username).Return(&dependencytrack.User{
					Username: username,
					Email:    username,
				}).Return(nil).Once()

				client.On("AddToTeam", mock.Anything, username, teamUuid).Return(nil).Once()
			},
		},
	} {
		mockClient := dependencytrack.NewMockClient(t)
		reconciler, err := New(database, auditLogger, []dependencytrack.Client{mockClient}, log)
		assert.NoError(t, err)

		if tt.preRun != nil {
			tt.preRun(t, mockClient)
		}

		err = reconciler.Reconcile(context.Background(), input)
		assert.NoError(t, err)
	}
}

func TestDependencytrackReconciler_Delete(t *testing.T) {

	log, err := logger.GetLogger("text", "info")
	assert.NoError(t, err)

	correlationID := uuid.New()
	input := setupInput(correlationID, "someTeam", "user1@nais.io")

	teamName := input.Team.Slug.String()

	for _, tt := range []struct {
		name   string
		preRun func(t *testing.T, mock *dependencytrack.MockClient)
	}{
		{
			name: "delete team from console should remove team from dependencytrack",
			preRun: func(t *testing.T, client *dependencytrack.MockClient) {
				teamNotInConsoleUuid := uuid.New().String()

				client.On("GetTeams", mock.Anything).Return([]dependencytrack.Team{
					{
						Name: teamName,
						Uuid: teamNotInConsoleUuid,
					},
				}, nil).Once()

				client.On("DeleteTeam", mock.Anything, teamNotInConsoleUuid).Return(nil).Once()
			},
		},
	} {
		mockClient := dependencytrack.NewMockClient(t)
		database := db.NewMockDatabase(t)
		auditLogger := auditlogger.NewMockAuditLogger(t)
		reconciler, err := New(database, auditLogger, []dependencytrack.Client{mockClient}, log)
		assert.NoError(t, err)

		if tt.preRun != nil {
			tt.preRun(t, mockClient)
		}

		err = reconciler.Delete(context.Background(), input.Team.Slug, uuid.New())
		assert.NoError(t, err)
	}
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
