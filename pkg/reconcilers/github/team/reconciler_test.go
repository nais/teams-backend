package github_team_reconciler_test

import (
	"context"
	"net/http"
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

	// Give the reconciler enough data to create an entire team from scratch,
	// remove members that shouldn't be present, and add members that should.
	t.Run("create everything from scratch", func(t *testing.T) {
		teamsService := github_team_reconciler.NewMockTeamsService(t)
		graphClient := github_team_reconciler.NewMockGraphClient(t)
		reconciler := github_team_reconciler.New(logger, org, teamsService, graphClient)

		teamsService.On("GetTeamBySlug", mock.Anything, org, teamName).
			Return(nil, &github.Response{
				Response: &http.Response{
					StatusCode: http.StatusNotFound,
				},
			}, nil).Once()

		teamsService.On("CreateTeam", mock.Anything, org,
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

		registerLoginEmailMock := func(org, email, login string) {
			graphClient.On(
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

		registerLoginEmailMock(org, keepEmail, keepLogin)
		registerLoginEmailMock(org, createEmail, createLogin)

		teamsService.On("ListTeamMembersBySlug", mock.Anything, org, teamName, mock.Anything).
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

		teamsService.On("AddTeamMembershipBySlug", mock.Anything, org, teamName, createLogin, mock.Anything).
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

		teamsService.On("RemoveTeamMembershipBySlug", mock.Anything, org, teamName, removeLogin, mock.Anything).
			Return(
				&github.Response{
					Response: &http.Response{
						StatusCode: http.StatusNoContent,
					},
				},
				nil,
			).Once()

		err := reconciler.Reconcile(ctx, reconcilers.Input{
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
		})

		assert.NoError(t, err)
		teamsService.AssertExpectations(t)
		graphClient.AssertExpectations(t)
	})
}

func strp(s string) *string {
	return &s
}
