package github_team_reconciler

import (
	"context"
	"fmt"
	"github.com/bradleyfalzon/ghinstallation/v2"
	"github.com/google/go-github/v43/github"
	"github.com/nais/console/pkg/dbmodels"
	"github.com/nais/console/pkg/reconcilers"
	"github.com/shurcooL/githubv4"
	"net/http"
	"time"
)

type gitHubReconciler struct {
	logs                chan<- *dbmodels.AuditLog
	ghAppId             int64
	ghAppInstallationId int64
	org                 string
	privateKeyPath      string
}

func (s *gitHubReconciler) getTransport() (*ghinstallation.Transport, error) {
	itr, err := ghinstallation.NewKeyFromFile(
		http.DefaultTransport,
		s.ghAppId,
		s.ghAppInstallationId,
		s.privateKeyPath,
	)

	if err != nil {
		return nil, err
	}

	return itr, nil
}

func (s *gitHubReconciler) getRestClient() (*github.Client, error) {
	transport, err := s.getTransport()

	if err != nil {
		return nil, err
	}

	return github.NewClient(&http.Client{
		Transport: transport,
	}), nil
}

func (s *gitHubReconciler) getGraphQlClient() (*githubv4.Client, error) {
	transport, err := s.getTransport()

	if err != nil {
		return nil, err
	}

	return githubv4.NewClient(&http.Client{
		Transport: transport,
	}), nil
}

func New(logs chan<- *dbmodels.AuditLog, ghAppId, ghInstallationId int64, org, privateKeyPath string) *gitHubReconciler {
	return &gitHubReconciler{
		logs:                logs,
		ghAppId:             ghAppId,
		ghAppInstallationId: ghInstallationId,
		org:                 org,
		privateKeyPath:      privateKeyPath,
	}
}

func (s *gitHubReconciler) Name() string {
	return "github:team"
}

func (s *gitHubReconciler) Op(operation string) string {
	return s.Name() + ":" + operation
}

func (s *gitHubReconciler) Reconcile(ctx context.Context, in reconcilers.Input) error {
	client, err := s.getRestClient()

	if err != nil {
		return in.AuditLog(nil, false, s.Op("create"), "retrieve API client: %s", err)
	}

	_, err = s.getOrCreateTeam(ctx, client.Teams, in)
	if err != nil {
		return in.AuditLog(nil, false, s.Op("create"), "ensure team exists: %s", err)
	}

	// TODO: sync members

	return nil
}

func (s *gitHubReconciler) getOrCreateTeam(ctx context.Context, teamsService *github.TeamsService, in reconcilers.Input) (*github.Team, error) {
	if in.Team == nil || in.Team.Slug == nil {
		return nil, fmt.Errorf("refusing to create team with empty slug")
	}

	existingTeam, _, err := teamsService.GetTeamBySlug(ctx, s.org, *in.Team.Slug)

	if err == nil {
		return existingTeam, nil
	}

	description := stringWithFallback(in.Team.Purpose, fmt.Sprintf("auto-generated by nais console on %s", time.Now().Format(time.RFC1123Z)))

	newTeam := github.NewTeam{
		Name:        stringWithFallback(in.Team.Name, fmt.Sprintf("NAIS team '%s'", *in.Team.Slug)),
		Description: &description,
	}

	team, _, err := teamsService.CreateTeam(ctx, s.org, newTeam)
	if err != nil {
		return nil, fmt.Errorf("create new team: %w", err)
	}

	s.logs <- in.AuditLog(nil, true, s.Op("create"), "successfully created team")

	return team, nil
}

func stringWithFallback(strp *string, fallback string) string {
	if strp == nil {
		return fallback
	}
	return *strp
}
