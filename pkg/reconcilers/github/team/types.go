package github_team_reconciler

import (
	"context"

	"github.com/google/go-github/v50/github"
	"github.com/nais/teams-backend/pkg/auditlogger"
	"github.com/nais/teams-backend/pkg/db"
	"github.com/nais/teams-backend/pkg/logger"
	"github.com/shurcooL/githubv4"
)

type GraphClient interface {
	Query(ctx context.Context, q interface{}, variables map[string]interface{}) error
}

type TeamsService interface {
	AddTeamMembershipBySlug(ctx context.Context, org, slug, user string, opts *github.TeamAddTeamMembershipOptions) (*github.Membership, *github.Response, error)
	CreateTeam(ctx context.Context, org string, team github.NewTeam) (*github.Team, *github.Response, error)
	GetTeamBySlug(ctx context.Context, org, slug string) (*github.Team, *github.Response, error)
	EditTeamBySlug(ctx context.Context, org, slug string, team github.NewTeam, removeParent bool) (*github.Team, *github.Response, error)
	ListTeamMembersBySlug(ctx context.Context, org, slug string, opts *github.TeamListTeamMembersOptions) ([]*github.User, *github.Response, error)
	RemoveTeamMembershipBySlug(ctx context.Context, org, slug, user string) (*github.Response, error)
	CreateOrUpdateIDPGroupConnectionsBySlug(ctx context.Context, org, team string, opts github.IDPGroupList) (*github.IDPGroupList, *github.Response, error)
	ListTeamReposBySlug(ctx context.Context, org, slug string, opts *github.ListOptions) ([]*github.Repository, *github.Response, error)
	DeleteTeamBySlug(ctx context.Context, org, slug string) (*github.Response, error)
}

// githubTeamReconciler creates teams on GitHub and connects users to them.
type githubTeamReconciler struct {
	database     db.Database
	auditLogger  auditlogger.AuditLogger
	teamsService TeamsService
	graphClient  GraphClient
	org          string
	domain       string
	log          logger.Logger
}

type GitHubUser struct {
	Login githubv4.String
}

type ExternalIdentitySamlAttributes struct {
	Username githubv4.String
}

type ExternalIdentity struct {
	User         GitHubUser
	SamlIdentity ExternalIdentitySamlAttributes
}

type LookupGitHubSamlUserByEmail struct {
	Organization struct {
		SamlIdentityProvider struct {
			ExternalIdentities struct {
				Nodes []ExternalIdentity
			} `graphql:"externalIdentities(first: 1, userName: $username)"`
		}
	} `graphql:"organization(login: $org)"`
}

type LookupGitHubSamlUserByGitHubUsername struct {
	Organization struct {
		SamlIdentityProvider struct {
			ExternalIdentities struct {
				Nodes []ExternalIdentity
			} `graphql:"externalIdentities(first: 1, login: $login)"`
		}
	} `graphql:"organization(login: $org)"`
}
