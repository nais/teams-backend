package dependencytrack

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/nais/console/pkg/db"
	"github.com/nais/console/pkg/reconcilers"
	"github.com/nais/console/pkg/slug"
	"github.com/nais/console/pkg/sqlc"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestDependencytrackReconciler_Reconcile(t *testing.T) {

	//existingTeamName := "existingTeam"
	input := setupInput("someTeam", "user1@nais.io")

	for _, tt := range []struct {
		name   string
		preRun func(t *testing.T, mock *MockClient)
	}{
		{
			name: "team does not exist, new team created and new members added",
			preRun: func(t *testing.T, client *MockClient) {
				teamName := "someTeam"
				teamUuid := uuid.New().String()
				username := "user1@nais.io"

				client.On("GetTeams", mock.Anything).Return([]Team{}, nil).Once()
				client.On("CreateTeam", mock.Anything, teamName, []Permission{
					ViewPortfolioPermission,
				}).Return(&Team{
					Uuid:      teamUuid,
					Name:      "newteam",
					OidcUsers: nil,
				}, nil).Once()

				client.On("CreateUser", mock.Anything, username).Return(&User{
					Username: username,
					Email:    username,
				}).Return(nil).Once()

				client.On("AddToTeam", mock.Anything, username, teamUuid).Return(nil).Once()
			},
		},
		{
			name: "team exists, new members added",
			preRun: func(t *testing.T, client *MockClient) {
				teamName := "someTeam"
				teamUuid := uuid.New().String()
				username := "user1@nais.io"

				client.On("GetTeams", mock.Anything).Return([]Team{
					{
						Name: teamName,
						Uuid: teamUuid,
					},
				}, nil).Once()

				client.On("CreateUser", mock.Anything, username).Return(&User{
					Username: username,
					Email:    username,
				}).Return(nil).Once()

				client.On("AddToTeam", mock.Anything, username, teamUuid).Return(nil).Once()
			},
		},
	} {
		mockClient := NewMockClient(t)
		reconciler, err := New(context.Background(), mockClient)
		assert.NoError(t, err)

		if tt.preRun != nil {
			tt.preRun(t, mockClient)
		}

		err = reconciler.Reconcile(context.Background(), input)
		assert.NoError(t, err)
	}
}

func setupInput(teamSlug string, members ...string) reconcilers.Input {
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
		CorrelationID:   uuid.New(),
		Team:            inputTeam,
		TeamMembers:     inputMembers,
		NumSyncAttempts: 0,
	}
}

/*
func TestDependencytrackReconciler_Reconcile(t *testing.T) {
	//database := db.NewMockDatabase(t)
	//auditLogger := auditlogger.NewMockAuditLogger(t)
	//log := logger.NewMockLogger(t)

	mock := newMock()

	reconciler, err := New(context.Background(), mock)
	assert.NoError(t, err)

	permission := ViewPortfolioPermission
	defaultTeamPermissions := []*Permission{
		&permission,
	}

	existingTeamName := "existingTeam"

	for _, tt := range []struct {
		name    string
		team    string
		members []string
		preRun  func(t *testing.T)
	}{
		{
			name:    "team exists, new member added",
			team:    existingTeamName,
			members: []string{"nybruker@dev-nais.io"},
			preRun: func(t *testing.T) {
				_, err = mock.CreateTeam(context.Background(), existingTeamName, []Permission{
					permission,
				})
				assert.NoError(t, err)
				err = mock.CreateUser(context.Background(), "shouldBeDeleted@nais.io")
				assert.NoError(t, err)
			},
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			if tt.preRun != nil {
				tt.preRun(t)
			}

			inputTeam := db.Team{
				Team: &sqlc.Team{
					Slug:    slug.Slug(tt.team),
					Purpose: "teamPurpose",
				},
			}

			inputMembers := make([]*db.User, 0)
			for _, member := range tt.members {
				inputMembers = append(inputMembers, &db.User{
					User: &sqlc.User{
						Email: member,
					},
				})
			}

			input := reconcilers.Input{
				CorrelationID:   uuid.New(),
				Team:            inputTeam,
				TeamMembers:     inputMembers,
				NumSyncAttempts: 0,
			}

			err = reconciler.Reconcile(context.Background(), input)
			if err != nil {
				t.Fatal(err)
			}

			assert.True(t, teamExists(mock, tt.team), "inputTeam should exist")
			assert.ElementsMatch(t, defaultTeamPermissions, mock.permissions[tt.team])
			assert.Equal(t, tt.members, mock.membersInTeam(tt.team), "members should match")

			mock.reset()
		})
	}
}

func teamExists(mock *mock, team string) bool {
	for _, t := range mock.teams {
		if t.Name == team {
			return true
		}
	}

	return false
}
*/
