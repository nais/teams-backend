package github_team_reconciler_test

import (
	"context"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/nais/console/pkg/slug"

	"github.com/google/go-github/v43/github"
	"github.com/google/uuid"
	"github.com/nais/console/pkg/auditlogger"
	helpers "github.com/nais/console/pkg/console"
	"github.com/nais/console/pkg/db"
	"github.com/nais/console/pkg/reconcilers"
	github_team_reconciler "github.com/nais/console/pkg/reconcilers/github/team"
	"github.com/nais/console/pkg/sqlc"
	"github.com/shurcooL/githubv4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGitHubReconciler_getOrCreateTeam(t *testing.T) {
	domain := "example.com"
	org := "org"
	teamSlug := "slug"
	teamPurpose := "purpose"

	ctx := context.Background()
	correlationID := uuid.New()
	team := db.Team{
		Team: &sqlc.Team{
			ID:      uuid.New(),
			Slug:    slug.Slug(teamSlug),
			Purpose: teamPurpose,
		},
	}
	input := reconcilers.Input{
		CorrelationID: correlationID,
		Team:          team,
	}
	systemName := github_team_reconciler.Name

	t.Run("no existing state, github team available", func(t *testing.T) {
		database := db.NewMockDatabase(t)
		teamsService := github_team_reconciler.NewMockTeamsService(t)
		auditLogger := auditlogger.NewMockAuditLogger(t)
		gitHubClient := github_team_reconciler.NewMockGraphClient(t)

		database.
			On("LoadReconcilerStateForTeam", ctx, systemName, team.ID, mock.Anything).
			Return(nil).
			Once()
		database.
			On("SetReconcilerStateForTeam", ctx, systemName, team.ID, mock.Anything).
			Return(nil).
			Once()

		teamsService.
			On(
				"CreateTeam",
				ctx,
				org,
				github.NewTeam{Name: teamSlug, Description: &teamPurpose},
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

		configureSyncTeamInfo(teamsService, org, teamSlug, teamPurpose)
		configureDeleteTeamIDP(teamsService, org, teamSlug)

		slug := slug.Slug(teamSlug)
		auditLogger.
			On("Logf", ctx, mock.MatchedBy(func(t []auditlogger.Target) bool {
				return t[0].Identifier == string(slug)
			}), mock.MatchedBy(func(f auditlogger.Fields) bool {
				return f.Action == sqlc.AuditActionGithubTeamCreate && f.CorrelationID == correlationID
			}), mock.Anything, mock.Anything).
			Return(nil).
			Once()

		reconciler := github_team_reconciler.New(database, auditLogger, org, domain, teamsService, gitHubClient)
		err := reconciler.Reconcile(ctx, input)
		assert.NoError(t, err)
	})

	t.Run("no existing state, github team not available", func(t *testing.T) {
		database := db.NewMockDatabase(t)
		teamsService := github_team_reconciler.NewMockTeamsService(t)
		auditLogger := auditlogger.NewMockAuditLogger(t)
		gitHubClient := github_team_reconciler.NewMockGraphClient(t)

		database.
			On("LoadReconcilerStateForTeam", ctx, systemName, team.ID, mock.Anything).
			Return(nil).
			Once()

		teamsService.
			On(
				"CreateTeam",
				ctx,
				org,
				github.NewTeam{Name: teamSlug, Description: &teamPurpose},
			).
			Return(
				nil,
				&github.Response{Response: &http.Response{StatusCode: http.StatusUnprocessableEntity}},
				nil,
			).
			Once()

		reconciler := github_team_reconciler.New(database, auditLogger, org, domain, teamsService, gitHubClient)
		err := reconciler.Reconcile(ctx, input)
		assert.Error(t, err)
	})

	t.Run("existing state, github team exists", func(t *testing.T) {
		database := db.NewMockDatabase(t)
		teamsService := github_team_reconciler.NewMockTeamsService(t)
		auditLogger := auditlogger.NewMockAuditLogger(t)
		gitHubClient := github_team_reconciler.NewMockGraphClient(t)

		database.
			On("LoadReconcilerStateForTeam", ctx, systemName, team.ID, mock.Anything).
			Run(func(args mock.Arguments) {
				slug := slug.Slug(teamSlug)
				state := args.Get(3).(*reconcilers.GitHubState)
				state.Slug = &slug
			}).
			Return(nil).
			Once()
		database.
			On("SetReconcilerStateForTeam", ctx, systemName, team.ID, mock.MatchedBy(func(state reconcilers.GitHubState) bool {
				return string(*state.Slug) == teamSlug
			})).
			Return(nil).
			Once()

		teamsService.
			On(
				"GetTeamBySlug",
				ctx,
				org,
				teamSlug,
			).
			Return(
				&github.Team{Slug: helpers.Strp(teamSlug)},
				&github.Response{Response: &http.Response{StatusCode: http.StatusOK}},
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

		configureSyncTeamInfo(teamsService, org, teamSlug, teamPurpose)
		configureDeleteTeamIDP(teamsService, org, teamSlug)

		reconciler := github_team_reconciler.New(database, auditLogger, org, domain, teamsService, gitHubClient)
		err := reconciler.Reconcile(ctx, input)
		assert.NoError(t, err)
	})

	t.Run("existing state, github team no longer exists", func(t *testing.T) {
		database := db.NewMockDatabase(t)
		teamsService := github_team_reconciler.NewMockTeamsService(t)
		auditLogger := auditlogger.NewMockAuditLogger(t)
		gitHubClient := github_team_reconciler.NewMockGraphClient(t)
		const existingSlug = "existing-slug"

		database.
			On("LoadReconcilerStateForTeam", ctx, systemName, team.ID, mock.Anything).
			Run(func(args mock.Arguments) {
				slug := slug.Slug(existingSlug)
				state := args.Get(3).(*reconcilers.GitHubState)
				state.Slug = &slug
			}).
			Return(nil).
			Once()
		database.
			On("SetReconcilerStateForTeam", ctx, systemName, team.ID, mock.MatchedBy(func(state reconcilers.GitHubState) bool {
				return *state.Slug == existingSlug
			})).
			Return(nil).
			Once()

		teamsService.
			On(
				"GetTeamBySlug",
				ctx,
				org,
				existingSlug,
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
				github.NewTeam{Name: existingSlug, Description: &teamPurpose},
			).
			Return(
				&github.Team{Slug: helpers.Strp(existingSlug)},
				&github.Response{Response: &http.Response{StatusCode: http.StatusCreated}},
				nil,
			).
			Once()
		teamsService.
			On(
				"ListTeamMembersBySlug",
				mock.Anything,
				org,
				existingSlug,
				mock.Anything,
			).
			Return(
				[]*github.User{},
				&github.Response{Response: &http.Response{StatusCode: http.StatusOK}},
				nil,
			).
			Once()

		configureSyncTeamInfo(teamsService, org, existingSlug, teamPurpose)
		configureDeleteTeamIDP(teamsService, org, existingSlug)

		slug := slug.Slug(teamSlug)
		auditLogger.
			On("Logf", ctx, mock.MatchedBy(func(t []auditlogger.Target) bool {
				return t[0].Identifier == string(slug)
			}), mock.MatchedBy(func(f auditlogger.Fields) bool {
				return f.Action == sqlc.AuditActionGithubTeamCreate && f.CorrelationID == correlationID
			}), mock.Anything, mock.Anything).
			Return(nil).
			Once()

		reconciler := github_team_reconciler.New(database, auditLogger, org, domain, teamsService, gitHubClient)
		err := reconciler.Reconcile(ctx, input)
		assert.NoError(t, err)
	})
}

func TestGitHubReconciler_Reconcile(t *testing.T) {
	domain := "example.com"
	org := "my-organization"
	teamName := "myteam"
	teamSlug := slug.Slug("myteam")
	teamPurpose := "some purpose"

	createLogin := "should-create"
	createEmail := "should-create@example.com"
	keepLogin := "should-keep"
	keepEmail := "should-keep@example.com"
	removeLogin := "should-remove"
	removeEmail := "should-remove@example.com"

	ctx := context.Background()

	correlationID := uuid.New()

	team := db.Team{
		Team: &sqlc.Team{
			ID:      uuid.New(),
			Slug:    teamSlug,
			Purpose: teamPurpose,
		},
	}

	input := reconcilers.Input{
		CorrelationID: correlationID,
		Team:          team,
		TeamMembers: []*db.User{
			{Email: createEmail},
			{Email: keepEmail},
		},
	}

	systemName := github_team_reconciler.Name

	auditLogger := auditlogger.NewMockAuditLogger(t)
	auditLogger.
		On("Logf", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).
		Return(nil)

	// Give the reconciler enough data to create an entire team from scratch,
	// remove members that shouldn't be present, and add members that should.
	t.Run("create everything from scratch", func(t *testing.T) {
		database := db.NewMockDatabase(t)
		teamsService := github_team_reconciler.NewMockTeamsService(t)
		graphClient := github_team_reconciler.NewMockGraphClient(t)

		database.
			On("LoadReconcilerStateForTeam", ctx, systemName, team.ID, mock.Anything).
			Return(nil).
			Once()
		database.
			On("SetReconcilerStateForTeam", ctx, systemName, team.ID, mock.MatchedBy(func(state reconcilers.GitHubState) bool {
				return *state.Slug == teamSlug
			})).
			Return(nil).
			Once()
		database.
			On("GetUserByEmail", ctx, removeEmail).
			Return(&db.User{Email: removeEmail, Name: removeLogin}, nil).
			Once()

		configureCreateTeam(teamsService, org, teamName, teamPurpose)
		configureSyncTeamInfo(teamsService, org, teamName, teamPurpose)

		configureLookupEmail(graphClient, org, removeLogin, removeEmail)

		configureRegisterLoginEmail(graphClient, org, keepEmail, keepLogin)
		configureRegisterLoginEmail(graphClient, org, createEmail, createLogin)

		configureListTeamMembersBySlug(teamsService, org, teamName, keepLogin, removeLogin)
		configureAddTeamMembershipBySlug(teamsService, org, teamName, createLogin)
		configureRemoveTeamMembershipBySlug(teamsService, org, teamName, removeLogin)

		configureDeleteTeamIDP(teamsService, org, teamName)

		reconciler := github_team_reconciler.New(database, auditLogger, org, domain, teamsService, graphClient)
		err := reconciler.Reconcile(ctx, input)
		assert.NoError(t, err)
	})

	t.Run("GetTeamBySlug error", func(t *testing.T) {
		teamsService := github_team_reconciler.NewMockTeamsService(t)
		graphClient := github_team_reconciler.NewMockGraphClient(t)
		database := db.NewMockDatabase(t)

		database.
			On("LoadReconcilerStateForTeam", ctx, systemName, team.ID, mock.Anything).
			Run(func(args mock.Arguments) {
				slug := slug.Slug("slug-from-state")
				state := args.Get(3).(*reconcilers.GitHubState)
				state.Slug = &slug
			}).
			Return(nil).
			Once()

		teamsService.On("GetTeamBySlug", mock.Anything, org, "slug-from-state").
			Return(nil, &github.Response{
				Response: &http.Response{
					StatusCode: http.StatusTeapot,
					Status:     "418: I'm a teapot",
					Body:       io.NopCloser(strings.NewReader("this is a body")),
				},
			}, nil).Once()

		reconciler := github_team_reconciler.New(database, auditLogger, org, domain, teamsService, graphClient)
		err := reconciler.Reconcile(ctx, input)

		assert.ErrorContainsf(t, err, "server error from GitHub: 418: I'm a teapot: this is a body", err.Error())
	})
}

func configureRegisterLoginEmail(graphClient *github_team_reconciler.MockGraphClient, org, email, login string) *mock.Call {
	return graphClient.
		On("Query", mock.Anything, mock.Anything, map[string]interface{}{"org": githubv4.String(org), "username": githubv4.String(email)}).
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

func configureLookupEmail(graphClient *github_team_reconciler.MockGraphClient, org, login, email string) *mock.Call {
	return graphClient.
		On("Query", mock.Anything, mock.Anything, map[string]interface{}{"org": githubv4.String(org), "login": githubv4.String(login)}).
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

func configureRemoveTeamMembershipBySlug(teamsService *github_team_reconciler.MockTeamsService, org, teamName, removeLogin string) *mock.Call {
	return teamsService.
		On("RemoveTeamMembershipBySlug", mock.Anything, org, teamName, removeLogin, mock.Anything).
		Return(
			&github.Response{
				Response: &http.Response{
					StatusCode: http.StatusNoContent,
				},
			},
			nil,
		).
		Once()
}

func configureAddTeamMembershipBySlug(teamsService *github_team_reconciler.MockTeamsService, org, teamName, createLogin string) *mock.Call {
	return teamsService.
		On("AddTeamMembershipBySlug", mock.Anything, org, teamName, createLogin, mock.Anything).
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
		).
		Once()
}

func configureListTeamMembersBySlug(teamsService *github_team_reconciler.MockTeamsService, org, teamName, keepLogin, removeLogin string) *mock.Call {
	return teamsService.
		On("ListTeamMembersBySlug", mock.Anything, org, teamName, mock.Anything).
		Return(
			[]*github.User{
				{Login: helpers.Strp(keepLogin)},
				{Login: helpers.Strp(removeLogin)},
			},
			&github.Response{
				Response: &http.Response{
					StatusCode: http.StatusOK,
				},
			},
			nil,
		).
		Once()
}

func configureCreateTeam(teamsService *github_team_reconciler.MockTeamsService, org, teamName, description string) *mock.Call {
	return teamsService.
		On("CreateTeam", mock.Anything, org, github.NewTeam{Name: teamName, Description: helpers.Strp(description)}).
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
		).
		Once()
}

func configureDeleteTeamIDP(teamsService *github_team_reconciler.MockTeamsService, org, slug string) *mock.Call {
	return teamsService.
		On(
			"CreateOrUpdateIDPGroupConnectionsBySlug",
			mock.Anything,
			org,
			slug,
			github.IDPGroupList{Groups: []*github.IDPGroup{}},
		).
		Return(&github.IDPGroupList{},
			&github.Response{Response: &http.Response{StatusCode: http.StatusOK}},
			nil,
		).Once()
}

func configureSyncTeamInfo(teamsService *github_team_reconciler.MockTeamsService, org, slug, purpose string) *mock.Call {
	return teamsService.
		On(
			"EditTeamBySlug",
			mock.Anything,
			org,
			slug,
			github.NewTeam{
				Name:        slug,
				Description: &purpose,
			},
			false,
		).
		Return(&github.Team{},
			&github.Response{Response: &http.Response{StatusCode: http.StatusOK}},
			nil,
		).Once()
}
