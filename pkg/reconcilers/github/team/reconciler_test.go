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

	addUser := &dbmodels.User{
		Email: strp("add@example.com"),
	}
	keepUser := &dbmodels.User{}

	t.Run("create everything from scratch", func(t *testing.T) {
		teamsService := github_team_reconciler.NewMockTeamsService(t)
		graphClient := github_team_reconciler.NewMockGraphClient(t)
		reconciler := github_team_reconciler.New(logger, org, teamsService, graphClient)

		teamsService.On("GetTeamBySlug", mock.Anything, org, teamName).
			Return(nil, &github.Response{Response: &http.Response{StatusCode: http.StatusNotFound}}, nil).Once()

		err := reconciler.Reconcile(ctx, reconcilers.Input{
			System:          nil,
			Synchronization: nil,
			Team: &dbmodels.Team{
				Slug:    strp(teamName),
				Name:    strp(teamName),
				Purpose: strp(teamName),
				Users: []*dbmodels.User{
					addUser, keepUser,
				},
			},
		})

		assert.NoError(t, err)
		teamsService.AssertExpectations(t)
	})
}

func strp(s string) *string {
	return &s
}
