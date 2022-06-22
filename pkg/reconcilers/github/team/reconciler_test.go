package github_team_reconciler_test

import (
	"context"
	"github.com/google/go-github/v43/github"
	"github.com/google/uuid"
	"github.com/nais/console/pkg/auditlogger"
	helpers "github.com/nais/console/pkg/console"
	"github.com/nais/console/pkg/dbmodels"
	"github.com/nais/console/pkg/reconcilers"
	github_team_reconciler "github.com/nais/console/pkg/reconcilers/github/team"
	"github.com/nais/console/pkg/test"
	"github.com/shurcooL/githubv4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"io/ioutil"
	"net/http"
	"strings"
	"testing"
)

func modelWithId() dbmodels.Model {
	id, _ := uuid.NewUUID()
	return dbmodels.Model{ID: &id}
}

func TestGitHubReconciler_getOrCreateTeam(t *testing.T) {
	const (
		domain      = "example.com"
		org         = "org"
		teamName    = "Team Name"
		teamSlug    = "slug"
		teamPurpose = "purpose"
	)

	ctx := context.Background()
	corr := dbmodels.Correlation{Model: modelWithId()}
	system := dbmodels.System{Model: modelWithId(), Name: github_team_reconciler.Name}
	team := dbmodels.Team{
		Model:   modelWithId(),
		Slug:    teamSlug,
		Name:    teamName,
		Purpose: helpers.Strp(teamPurpose),
	}
	input := reconcilers.Input{
		Corr: corr,
		Team: team,
	}

	auditLogger := &auditlogger.MockAuditLogger{}
	auditLogger.On("Logf", github_team_reconciler.OpCreate, corr, system, mock.Anything, &team, mock.Anything, mock.Anything, mock.Anything).Return(nil)

	t.Run("no existing state, github team available", func(t *testing.T) {
		db := test.GetTestDB()
		db.AutoMigrate(&dbmodels.SystemState{})
		teamsService := github_team_reconciler.NewMockTeamsService(t)
		teamsService.
			On(
				"CreateTeam",
				ctx,
				org,
				github.NewTeam{Name: teamSlug, Description: helpers.Strp(teamPurpose)},
			).
			Return(
				&github.Team{Slug: helpers.Strp(teamSlug)},
				&github.Response{Response: &http.Response{StatusCode: http.StatusCreated}},
				nil,
			).
			Once()
		teamsService.
			On(
				"ListTeamMembersBySlug",
				mock.Anything,
				org,
				teamSlug,
				mock.Anything,
			).
			Return(
				[]*github.User{},
				&github.Response{Response: &http.Response{StatusCode: http.StatusOK}},
				nil,
			).
			Once()

		reconciler := github_team_reconciler.New(db, system, auditLogger, org, domain, teamsService, github_team_reconciler.NewMockGraphClient(t))
		err := reconciler.Reconcile(ctx, input)
		assert.NoError(t, err)
	})

	t.Run("no existing state, github team not available", func(t *testing.T) {
		db := test.GetTestDB()
		db.AutoMigrate(&dbmodels.SystemState{})
		teamsService := github_team_reconciler.NewMockTeamsService(t)
		teamsService.
			On(
				"CreateTeam",
				ctx,
				org,
				github.NewTeam{Name: teamSlug, Description: helpers.Strp(teamPurpose)},
			).
			Return(
				nil,
				&github.Response{Response: &http.Response{StatusCode: http.StatusUnprocessableEntity}},
				nil,
			).
			Once()

		reconciler := github_team_reconciler.New(db, system, auditLogger, org, domain, teamsService, github_team_reconciler.NewMockGraphClient(t))
		err := reconciler.Reconcile(ctx, input)
		assert.Error(t, err)
	})

	t.Run("existing state, github team exists", func(t *testing.T) {
		db := test.GetTestDB()
		db.AutoMigrate(&dbmodels.SystemState{})
		systemState := &dbmodels.SystemState{
			SystemID: *system.ID,
			TeamID:   *team.ID,
		}
		systemState.State.Set(github_team_reconciler.GitHubState{Slug: helpers.Strp("existing-slug")})
		db.Create(systemState)
		teamsService := github_team_reconciler.NewMockTeamsService(t)
		teamsService.
			On(
				"GetTeamBySlug",
				ctx,
				org,
				"existing-slug",
			).
			Return(
				&github.Team{Slug: helpers.Strp("existing-slug")},
				&github.Response{Response: &http.Response{StatusCode: http.StatusOK}},
				nil,
			).
			Once()
		teamsService.
			On(
				"ListTeamMembersBySlug",
				mock.Anything,
				org,
				"existing-slug",
				mock.Anything,
			).
			Return(
				[]*github.User{},
				&github.Response{Response: &http.Response{StatusCode: http.StatusOK}},
				nil,
			).
			Once()

		reconciler := github_team_reconciler.New(db, system, auditLogger, org, domain, teamsService, github_team_reconciler.NewMockGraphClient(t))
		err := reconciler.Reconcile(ctx, input)
		assert.NoError(t, err)
	})

	t.Run("existing state, github team no longer exists", func(t *testing.T) {
		db := test.GetTestDB()
		db.AutoMigrate(&dbmodels.SystemState{})
		initialState := &github_team_reconciler.GitHubState{Slug: helpers.Strp("existing-slug")}
		systemState := &dbmodels.SystemState{
			SystemID: *system.ID,
			TeamID:   *team.ID,
		}
		systemState.State.Set(initialState)
		db.Create(systemState)
		teamsService := github_team_reconciler.NewMockTeamsService(t)
		teamsService.
			On(
				"GetTeamBySlug",
				ctx,
				org,
				"existing-slug",
			).
			Return(
				nil,
				&github.Response{Response: &http.Response{StatusCode: http.StatusNotFound}},
				nil,
			).
			Once()
		teamsService.
			On(
				"CreateTeam",
				ctx,
				org,
				github.NewTeam{Name: teamSlug, Description: helpers.Strp(teamPurpose)},
			).
			Return(
				&github.Team{Slug: helpers.Strp(teamSlug)},
				&github.Response{Response: &http.Response{StatusCode: http.StatusCreated}},
				nil,
			).
			Once()
		teamsService.
			On(
				"ListTeamMembersBySlug",
				mock.Anything,
				org,
				teamSlug,
				mock.Anything,
			).
			Return(
				[]*github.User{},
				&github.Response{Response: &http.Response{StatusCode: http.StatusOK}},
				nil,
			).
			Once()

		reconciler := github_team_reconciler.New(db, system, auditLogger, org, domain, teamsService, github_team_reconciler.NewMockGraphClient(t))
		err := reconciler.Reconcile(ctx, input)
		assert.NoError(t, err)

		// fetch updated system state
		updatedState := &github_team_reconciler.GitHubState{}
		db.Where("id = ?", systemState.ID).First(systemState)
		systemState.State.AssignTo(updatedState)
		assert.Equal(t, "slug", *updatedState.Slug)
	})
}

func TestGitHubReconciler_Reconcile(t *testing.T) {
	const (
		domain      = "example.com"
		org         = "my-organization"
		teamName    = "myteam"
		teamSlug    = "myteam"
		teamPurpose = "some purpose"

		createLogin = "should-create"
		createEmail = "should-create@example.com"
		keepLogin   = "should-keep"
		keepEmail   = "should-keep@example.com"
		removeLogin = "should-remove"
		removeEmail = "should-remove@example.com"
	)

	ctx := context.Background()

	auditLogger := &auditlogger.MockAuditLogger{}
	auditLogger.On("Logf", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)

	system := dbmodels.System{Model: modelWithId(), Name: github_team_reconciler.Name}
	corr := dbmodels.Correlation{Model: modelWithId()}

	team := dbmodels.Team{
		Model:   modelWithId(),
		Slug:    teamSlug,
		Name:    teamName,
		Purpose: helpers.Strp(teamPurpose),
		Users: []*dbmodels.User{
			{
				Email: createEmail,
			},
			{
				Email: keepEmail,
			},
		},
	}

	input := reconcilers.Input{
		Corr: corr,
		Team: team,
	}

	// Give the reconciler enough data to create an entire team from scratch,
	// remove members that shouldn't be present, and add members that should.
	t.Run("create everything from scratch", func(t *testing.T) {
		db := test.GetTestDB()
		db.AutoMigrate(&dbmodels.SystemState{})
		teamsService := github_team_reconciler.NewMockTeamsService(t)
		graphClient := github_team_reconciler.NewMockGraphClient(t)
		reconciler := github_team_reconciler.New(db, system, auditLogger, org, domain, teamsService, graphClient)

		configureCreateTeam(teamsService, org, teamName, teamPurpose)

		configureLookupEmail(graphClient, org, removeLogin, removeEmail)

		configureRegisterLoginEmail(graphClient, org, keepEmail, keepLogin)
		configureRegisterLoginEmail(graphClient, org, createEmail, createLogin)

		configureListTeamMembersBySlug(teamsService, org, teamName, keepLogin, removeLogin)
		configureAddTeamMembershipBySlug(teamsService, org, teamName, createLogin)
		configureRemoveTeamMembershipBySlug(teamsService, org, teamName, removeLogin)

		err := reconciler.Reconcile(ctx, input)

		assert.NoError(t, err)
		teamsService.AssertExpectations(t)
		graphClient.AssertExpectations(t)
	})

	t.Run("GetTeamBySlug error", func(t *testing.T) {
		db := test.GetTestDB()
		db.AutoMigrate(&dbmodels.SystemState{})
		systemState := &dbmodels.SystemState{
			SystemID: *system.ID,
			TeamID:   *team.ID,
		}
		systemState.State.Set(github_team_reconciler.GitHubState{Slug: helpers.Strp("slug-from-state")})
		db.Create(systemState)

		teamsService := github_team_reconciler.NewMockTeamsService(t)
		graphClient := github_team_reconciler.NewMockGraphClient(t)
		reconciler := github_team_reconciler.New(db, system, auditLogger, org, domain, teamsService, graphClient)

		teamsService.On("GetTeamBySlug", mock.Anything, org, "slug-from-state").
			Return(nil, &github.Response{
				Response: &http.Response{
					StatusCode: http.StatusTeapot,
					Status:     "418: I'm a teapot",
					Body:       ioutil.NopCloser(strings.NewReader("this is a body")),
				},
			}, nil).Once()

		err := reconciler.Reconcile(ctx, input)

		assert.ErrorContainsf(t, err, "server error from GitHub: 418: I'm a teapot: this is a body", err.Error())
		teamsService.AssertExpectations(t)
		graphClient.AssertExpectations(t)
	})
}

func configureRegisterLoginEmail(graphClient *github_team_reconciler.MockGraphClient, org string, email string, login string) *mock.Call {
	return graphClient.On(
		"Query",
		mock.Anything,
		mock.Anything,
		map[string]interface{}{
			"org":      githubv4.String(org),
			"username": githubv4.String(email),
		},
	).
		Run(
			func(args mock.Arguments) {
				query := args.Get(1).(*github_team_reconciler.LookupGitHubSamlUserByEmail)
				query.Organization.SamlIdentityProvider.ExternalIdentities.Nodes = []github_team_reconciler.ExternalIdentity{
					{
						User: github_team_reconciler.GitHubUser{
							Login: githubv4.String(login),
						},
					},
				}
			},
		).
		Once().
		Return(nil)
}

func configureLookupEmail(graphClient *github_team_reconciler.MockGraphClient, org string, login, email string) *mock.Call {
	return graphClient.On(
		"Query",
		mock.Anything,
		mock.Anything,
		map[string]interface{}{
			"org":   githubv4.String(org),
			"login": githubv4.String(login),
		},
	).
		Run(
			func(args mock.Arguments) {
				query := args.Get(1).(*github_team_reconciler.LookupGitHubSamlUserByGitHubUsername)
				query.Organization.SamlIdentityProvider.ExternalIdentities.Nodes = []github_team_reconciler.ExternalIdentity{
					{
						SamlIdentity: github_team_reconciler.ExternalIdentitySamlAttributes{
							Username: githubv4.String(email),
						},
					},
				}
			},
		).
		Once().
		Return(nil)
}

func configureRemoveTeamMembershipBySlug(teamsService *github_team_reconciler.MockTeamsService, org string, teamName string, removeLogin string) *mock.Call {
	return teamsService.On("RemoveTeamMembershipBySlug", mock.Anything, org, teamName, removeLogin, mock.Anything).
		Return(
			&github.Response{
				Response: &http.Response{
					StatusCode: http.StatusNoContent,
				},
			},
			nil,
		).Once()
}

func configureAddTeamMembershipBySlug(teamsService *github_team_reconciler.MockTeamsService, org string, teamName string, createLogin string) *mock.Call {
	return teamsService.On("AddTeamMembershipBySlug", mock.Anything, org, teamName, createLogin, mock.Anything).
		Return(
			&github.Membership{
				User: &github.User{
					Login: helpers.Strp(createLogin),
				},
			},
			&github.Response{
				Response: &http.Response{
					StatusCode: http.StatusOK,
				},
			},
			nil,
		).Once()
}

func configureListTeamMembersBySlug(teamsService *github_team_reconciler.MockTeamsService, org string, teamName string, keepLogin string, removeLogin string) *mock.Call {
	return teamsService.On("ListTeamMembersBySlug", mock.Anything, org, teamName, mock.Anything).
		Return(
			[]*github.User{
				{
					Login: helpers.Strp(keepLogin),
				},
				{
					Login: helpers.Strp(removeLogin),
				},
			},
			&github.Response{
				Response: &http.Response{
					StatusCode: http.StatusOK,
				},
			},
			nil,
		).Once()
}

func configureCreateTeam(teamsService *github_team_reconciler.MockTeamsService, org string, teamName string, description string) *mock.Call {
	return teamsService.On("CreateTeam", mock.Anything, org,
		github.NewTeam{
			Name:        teamName,
			Description: helpers.Strp(description),
		}).
		Return(
			&github.Team{
				Slug: helpers.Strp(teamName),
			},
			&github.Response{
				Response: &http.Response{
					StatusCode: http.StatusCreated,
				},
			},
			nil,
		).Once()
}
