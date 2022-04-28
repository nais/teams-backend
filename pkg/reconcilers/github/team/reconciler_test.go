package github_team_reconciler_test

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/google/go-github/v43/github"
	"github.com/nais/console/pkg/auditlogger"
	"github.com/nais/console/pkg/dbmodels"
	"github.com/nais/console/pkg/reconcilers"
	github_team_reconciler "github.com/nais/console/pkg/reconcilers/github/team"
	"github.com/shurcooL/githubv4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestAzureReconciler_Reconcile(t *testing.T) {
	const teamName = "myteam"

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*2)
	defer cancel()

	ch := make(chan *dbmodels.AuditLog, 100)
	logger := auditlogger.New(ch)
	defer close(ch)

	const org = "my-organization"
	const description = "this describes my organization"

	const createLogin = "should-create"
	const createEmail = "should-create@example.com"
	const keepLogin = "should-keep"
	const keepEmail = "should-keep@example.com"
	const removeLogin = "should-remove"

	reconcilerInput := reconcilers.Input{
		System:          nil,
		Synchronization: nil,
		Team: &dbmodels.Team{
			Slug:    strp(teamName),
			Name:    strp(teamName),
			Purpose: strp(description),
			Users: []*dbmodels.User{
				{
					Email: strp(createEmail),
				},
				{
					Email: strp(keepEmail),
				},
			},
		},
	}

	// Give the reconciler enough data to create an entire team from scratch,
	// remove members that shouldn't be present, and add members that should.
	t.Run("create everything from scratch", func(t *testing.T) {
		teamsService := github_team_reconciler.NewMockTeamsService(t)
		graphClient := github_team_reconciler.NewMockGraphClient(t)
		reconciler := github_team_reconciler.New(logger, org, teamsService, graphClient)

		configureGetTeamBySlug(teamsService, org, teamName)
		configureCreateTeam(teamsService, org, teamName, description)

		configureRegisterLoginEmail(graphClient, org, keepEmail, keepLogin)
		configureRegisterLoginEmail(graphClient, org, createEmail, createLogin)

		configureListTeamMembersBySlug(teamsService, org, teamName, keepLogin, removeLogin)
		configureAddTeamMembershipBySlug(teamsService, org, teamName, createLogin)
		configureRemoveTeamMembershipBySlug(teamsService, org, teamName, removeLogin)

		err := reconciler.Reconcile(ctx, reconcilerInput)

		assert.NoError(t, err)
		teamsService.AssertExpectations(t)
		graphClient.AssertExpectations(t)
	})

	t.Run("GetTeamBySlug error", func(t *testing.T) {
		teamsService := github_team_reconciler.NewMockTeamsService(t)
		graphClient := github_team_reconciler.NewMockGraphClient(t)
		reconciler := github_team_reconciler.New(logger, org, teamsService, graphClient)

		expectedErr := fmt.Errorf("GetTeamBySlug failed")
		teamsService.On("GetTeamBySlug", mock.Anything, org, teamName).
			Return(nil, nil, expectedErr).Once()

		err := reconciler.Reconcile(ctx, reconcilerInput)

		assert.EqualError(t, err, expectedErr.Error())
		teamsService.AssertExpectations(t)
		graphClient.AssertExpectations(t)
	})

	t.Run("GetTeamBySlug client error", func(t *testing.T) {
		teamsService := github_team_reconciler.NewMockTeamsService(t)
		graphClient := github_team_reconciler.NewMockGraphClient(t)
		reconciler := github_team_reconciler.New(logger, org, teamsService, graphClient)

		expectedErr := fmt.Errorf("server raised error: 418: I'm a teapot: this is a body")
		teamsService.On("GetTeamBySlug", mock.Anything, org, teamName).
			Return(nil, &github.Response{
				Response: &http.Response{
					StatusCode: http.StatusTeapot,
					Status:     "418: I'm a teapot",
					Body:       ioutil.NopCloser(strings.NewReader("this is a body")),
				},
			}, nil).Once()

		err := reconciler.Reconcile(ctx, reconcilerInput)

		assert.EqualError(t, err, expectedErr.Error())
		teamsService.AssertExpectations(t)
		graphClient.AssertExpectations(t)
	})

	t.Run("GetTeamBySlug team exists", func(t *testing.T) {
		teamsService := github_team_reconciler.NewMockTeamsService(t)
		graphClient := github_team_reconciler.NewMockGraphClient(t)
		reconciler := github_team_reconciler.New(logger, org, teamsService, graphClient)

		teamsService.On("GetTeamBySlug", mock.Anything, org, teamName).
			Return(&github.Team{
				Slug: strp(teamName),
			}, &github.Response{
				Response: &http.Response{
					StatusCode: http.StatusOK,
					Status:     "200: OK",
				},
			}, nil).Once()

		configureRegisterLoginEmail(graphClient, org, keepEmail, keepLogin)
		configureRegisterLoginEmail(graphClient, org, createEmail, createLogin)

		configureListTeamMembersBySlug(teamsService, org, teamName, keepLogin, removeLogin)
		configureAddTeamMembershipBySlug(teamsService, org, teamName, createLogin)
		configureRemoveTeamMembershipBySlug(teamsService, org, teamName, removeLogin)

		err := reconciler.Reconcile(ctx, reconcilerInput)

		assert.NoError(t, err)
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
				query := args.Get(1).(*github_team_reconciler.LookupSSOUserQuery)
				query.Organization.SamlIdentityProvider.ExternalIdentities.Nodes = []github_team_reconciler.LookupSSOUserNode{
					{
						User: github_team_reconciler.LookupSSOUser{
							Login: githubv4.String(login),
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
					Login: strp(createLogin),
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
					Login: strp(keepLogin),
				},
				{
					Login: strp(removeLogin),
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
			Description: strp(description),
		}).
		Return(
			&github.Team{
				Slug: strp(teamName),
			},
			&github.Response{
				Response: &http.Response{
					StatusCode: http.StatusCreated,
				},
			},
			nil,
		).Once()
}

func configureGetTeamBySlug(teamsService *github_team_reconciler.MockTeamsService, org string, teamName string) *mock.Call {
	return teamsService.On("GetTeamBySlug", mock.Anything, org, teamName).
		Return(nil, &github.Response{
			Response: &http.Response{
				StatusCode: http.StatusNotFound,
			},
		}, nil).Once()
}

func strp(s string) *string {
	return &s
}
