package github_team_reconciler

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/nais/console/pkg/slug"

	"github.com/bradleyfalzon/ghinstallation/v2"
	"github.com/google/go-github/v43/github"
	"github.com/google/uuid"
	"github.com/nais/console/pkg/auditlogger"
	"github.com/nais/console/pkg/config"
	helpers "github.com/nais/console/pkg/console"
	"github.com/nais/console/pkg/db"
	"github.com/nais/console/pkg/reconcilers"
	"github.com/nais/console/pkg/sqlc"
	"github.com/shurcooL/githubv4"
	log "github.com/sirupsen/logrus"
)

const Name = sqlc.SystemNameGithubTeam

var errGitHubUserNotFound = errors.New("GitHub user does not exist")

func New(database db.Database, auditLogger auditlogger.AuditLogger, org, domain string, teamsService TeamsService, graphClient GraphClient) *githubTeamReconciler {
	return &githubTeamReconciler{
		database:     database,
		auditLogger:  auditLogger,
		org:          org,
		domain:       domain,
		teamsService: teamsService,
		graphClient:  graphClient,
	}
}

func NewFromConfig(ctx context.Context, database db.Database, cfg *config.Config, auditLogger auditlogger.AuditLogger) (reconcilers.Reconciler, error) {
	if !cfg.GitHub.Enabled {
		return nil, reconcilers.ErrReconcilerNotEnabled
	}

	transport, err := ghinstallation.NewKeyFromFile(
		http.DefaultTransport,
		cfg.GitHub.AppID,
		cfg.GitHub.AppInstallationID,
		cfg.GitHub.PrivateKeyPath,
	)
	if err != nil {
		return nil, err
	}

	// Note that both HTTP clients and transports are safe for concurrent use according to the docs,
	// so we can safely reuse them across objects and concurrent synchronizations.
	httpClient := &http.Client{
		Transport: transport,
	}
	restClient := github.NewClient(httpClient)
	graphClient := githubv4.NewClient(httpClient)

	return New(database, auditLogger, cfg.GitHub.Organization, cfg.TenantDomain, restClient.Teams, graphClient), nil
}

func (r *githubTeamReconciler) Name() sqlc.SystemName {
	return Name
}

func (r *githubTeamReconciler) Reconcile(ctx context.Context, input reconcilers.Input) error {
	state := &reconcilers.GitHubState{}
	err := r.database.LoadSystemState(ctx, r.Name(), input.Team.ID, state)
	if err != nil {
		return fmt.Errorf("unable to load system state for team %q in system %q: %w", input.Team.Slug, r.Name(), err)
	}

	githubTeam, err := r.getOrCreateTeam(ctx, *state, input.CorrelationID, input.Team)
	if err != nil {
		return fmt.Errorf("unable to get or create a GitHub team for team %q in system %q: %w", input.Team.Slug, r.Name(), err)
	}

	slug := slug.Slug(*githubTeam.Slug)
	err = r.database.SetSystemState(ctx, r.Name(), input.Team.ID, reconcilers.GitHubState{Slug: &slug})
	if err != nil {
		log.Errorf("system state not persisted: %s", err)
	}

	return r.connectUsers(ctx, githubTeam, input)
}

func (r *githubTeamReconciler) getOrCreateTeam(ctx context.Context, state reconcilers.GitHubState, correlationID uuid.UUID, team db.Team) (*github.Team, error) {
	if state.Slug != nil {
		existingTeam, resp, err := r.teamsService.GetTeamBySlug(ctx, r.org, string(*state.Slug))
		if resp == nil && err != nil {
			return nil, fmt.Errorf("unable to fetch GitHub team %q: %w", *state.Slug, err)
		}

		switch resp.StatusCode {
		case http.StatusNotFound:
			break
		case http.StatusOK:
			return existingTeam, nil
		default:
			body, _ := io.ReadAll(resp.Body)
			return nil, fmt.Errorf("server error from GitHub: %s: %s", resp.Status, string(body))
		}
	}

	description := helpers.TeamPurpose(&team.Purpose.String)
	githubTeam, resp, err := r.teamsService.CreateTeam(ctx, r.org, github.NewTeam{
		Name:        string(team.Slug),
		Description: &description,
	})
	err = httpError(http.StatusCreated, *resp, err)
	if err != nil {
		return nil, fmt.Errorf("unable to create GitHub team: %w", err)
	}

	fields := auditlogger.Fields{
		Action:         sqlc.AuditActionGithubTeamCreate,
		CorrelationID:  correlationID,
		TargetTeamSlug: &team.Slug,
	}
	r.auditLogger.Logf(ctx, fields, "created GitHub team %q", *githubTeam.Slug)

	return githubTeam, nil
}

func (r *githubTeamReconciler) connectUsers(ctx context.Context, githubTeam *github.Team, input reconcilers.Input) error {
	membersAccordingToGitHub, err := r.getTeamMembers(ctx, *githubTeam.Slug)
	if err != nil {
		return fmt.Errorf("list existing members in GitHub team %q: %w", *githubTeam.Slug, err)
	}

	membersAccordingToConsole := helpers.DomainUsers(input.TeamMembers, r.domain)
	consoleUserWithGitHubUser, err := r.mapSSOUsers(ctx, membersAccordingToConsole)
	if err != nil {
		return err
	}

	membersToRemove := remoteOnlyMembers(membersAccordingToGitHub, consoleUserWithGitHubUser)
	for _, gitHubUser := range membersToRemove {
		username := gitHubUser.GetLogin()
		resp, err := r.teamsService.RemoveTeamMembershipBySlug(ctx, r.org, *githubTeam.Slug, username)
		err = httpError(http.StatusNoContent, *resp, err)
		if err != nil {
			log.Warnf("unable to remove member %q from GitHub team %q: %s", username, *githubTeam.Slug, err)
			continue
		}

		email, err := r.getEmailFromGitHubUsername(ctx, username)
		if err != nil {
			log.Warnf("unable to get email from GitHub username %q for audit log purposes: %s", username, err)
		}

		if email != nil {
			_, err = r.database.GetUserByEmail(ctx, *email)
			if err != nil {
				log.Warnf("unable to get Console user with email %q: %s", *email, err)
				email = nil
			}
		}

		fields := auditlogger.Fields{
			Action:         sqlc.AuditActionGithubTeamDeleteMember,
			CorrelationID:  input.CorrelationID,
			TargetTeamSlug: &input.Team.Slug,
		}
		r.auditLogger.Logf(ctx, fields, "deleted member %q from GitHub team %q", username, *githubTeam.Slug)
	}

	membersToAdd := localOnlyMembers(consoleUserWithGitHubUser, membersAccordingToGitHub)
	for username, consoleUser := range membersToAdd {
		_, resp, err := r.teamsService.AddTeamMembershipBySlug(ctx, r.org, *githubTeam.Slug, username, &github.TeamAddTeamMembershipOptions{})
		err = httpError(http.StatusOK, *resp, err)
		if err != nil {
			log.Warnf("unable to add member %q to GitHub team %q: %s", username, *githubTeam.Slug, err)
			continue
		}

		fields := auditlogger.Fields{
			Action:         sqlc.AuditActionGithubTeamAddMember,
			CorrelationID:  input.CorrelationID,
			TargetTeamSlug: &input.Team.Slug,
			TargetUser:     &consoleUser.Email,
		}
		r.auditLogger.Logf(ctx, fields, "added member %q to GitHub team %q", username, *githubTeam.Slug)
	}

	return nil
}

// getTeamMembers Get all team members in a GitHub team using a paginated query
func (r *githubTeamReconciler) getTeamMembers(ctx context.Context, slug string) ([]*github.User, error) {
	const maxPerPage = 100
	opt := &github.TeamListTeamMembersOptions{
		ListOptions: github.ListOptions{
			PerPage: maxPerPage,
		},
	}

	allMembers := make([]*github.User, 0)
	for {
		members, resp, err := r.teamsService.ListTeamMembersBySlug(ctx, r.org, slug, opt)
		err = httpError(http.StatusOK, *resp, err)
		if err != nil {
			return nil, err
		}
		allMembers = append(allMembers, members...)
		if resp.NextPage == 0 {
			break
		}
		opt.Page = resp.NextPage
	}

	return allMembers, nil
}

// localOnlyMembers Given a mapping of GitHub usernames to Console users, and a list of GitHub team members according to
// GitHub, return members only present in the mapping.
func localOnlyMembers(consoleUsers map[string]*db.User, membersAccordingToGitHub []*github.User) map[string]*db.User {
	gitHubUsernameMap := make(map[string]*github.User, 0)
	for _, gitHubUser := range membersAccordingToGitHub {
		gitHubUsernameMap[gitHubUser.GetLogin()] = gitHubUser
	}

	localOnly := make(map[string]*db.User, 0)
	for gitHubUsername, consoleUser := range consoleUsers {
		if _, exists := gitHubUsernameMap[gitHubUsername]; !exists {
			localOnly[gitHubUsername] = consoleUser
		}
	}
	return localOnly
}

// remoteOnlyMembers Given a list of GitHub team members and a mapping of known GitHub usernames to Console users,
// return members not present in the mapping.
func remoteOnlyMembers(membersAccordingToGitHub []*github.User, consoleUsers map[string]*db.User) []*github.User {
	unknownMembers := make([]*github.User, 0)
	for _, member := range membersAccordingToGitHub {
		if _, exists := consoleUsers[member.GetLogin()]; !exists {
			unknownMembers = append(unknownMembers, member)
		}
	}
	return unknownMembers
}

// mapSSOUsers Return a mapping of GitHub usernames to Console user objects. Console users with no matching GitHub user
// will be ignored.
func (r *githubTeamReconciler) mapSSOUsers(ctx context.Context, users []*db.User) (map[string]*db.User, error) {
	userMap := make(map[string]*db.User)
	for _, user := range users {
		githubUsername, err := r.getGitHubUsernameFromEmail(ctx, user.Email)
		if err == errGitHubUserNotFound {
			log.Warnf("no GitHub user for email: %q", user.Email)
			continue
		}
		if err != nil {
			return nil, err
		}
		userMap[*githubUsername] = user
	}

	return userMap, nil
}

// getGitHubUsernameFromEmail Look up a GitHub username from an SSO e-mail address connected to that user account.
func (r *githubTeamReconciler) getGitHubUsernameFromEmail(ctx context.Context, email string) (*string, error) {
	var query LookupGitHubSamlUserByEmail

	variables := map[string]interface{}{
		"org":      githubv4.String(r.org),
		"username": githubv4.String(email),
	}

	err := r.graphClient.Query(ctx, &query, variables)
	if err != nil {
		return nil, err
	}

	nodes := query.Organization.SamlIdentityProvider.ExternalIdentities.Nodes
	if len(nodes) == 0 {
		return nil, errGitHubUserNotFound
	}

	username := string(nodes[0].User.Login)
	return &username, nil
}

// getEmailFromGitHubUsername Look up a GitHub username from an SSO e-mail address connected to that user account.
func (r *githubTeamReconciler) getEmailFromGitHubUsername(ctx context.Context, username string) (*string, error) {
	var query LookupGitHubSamlUserByGitHubUsername

	variables := map[string]interface{}{
		"org":   githubv4.String(r.org),
		"login": githubv4.String(username),
	}

	err := r.graphClient.Query(ctx, &query, variables)
	if err != nil {
		return nil, err
	}

	nodes := query.Organization.SamlIdentityProvider.ExternalIdentities.Nodes
	if len(nodes) == 0 {
		return nil, errGitHubUserNotFound
	}

	email := strings.ToLower(string(nodes[0].SamlIdentity.Username))
	return &email, nil
}

// httpError Return an error if the response status code is not as expected, or if the passed err is already set to an
// error
func httpError(expected int, resp github.Response, err error) error {
	if err != nil {
		return err
	}

	if resp.StatusCode != expected {
		if resp.Body == nil {
			return errors.New("unknown error")
		}
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("unexpected response error %s: %s", resp.Status, string(body))
	}

	return nil
}
