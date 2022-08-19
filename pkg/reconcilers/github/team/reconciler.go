package github_team_reconciler

import (
	"context"
	"errors"
	"fmt"
	"github.com/bradleyfalzon/ghinstallation/v2"
	"github.com/google/go-github/v43/github"
	"github.com/nais/console/pkg/auditlogger"
	"github.com/nais/console/pkg/config"
	helpers "github.com/nais/console/pkg/console"
	"github.com/nais/console/pkg/db"
	"github.com/nais/console/pkg/dbmodels"
	"github.com/nais/console/pkg/reconcilers"
	"github.com/nais/console/pkg/sqlc"
	"github.com/shurcooL/githubv4"
	log "github.com/sirupsen/logrus"
	"gorm.io/gorm"
	"io/ioutil"
	"net/http"
	"strings"
)

const (
	Name           = "github:team"
	OpCreate       = "github:team:create"
	OpAddMembers   = "github:team:add-members"
	OpAddMember    = "github:team:add-member"
	OpDeleteMember = "github:team:delete-member"
	OpMapSSOUser   = "github:team:map-sso-user"
)

var errGitHubUserNotFound = errors.New("GitHub user does not exist")

func New(queries sqlc.Querier, db *gorm.DB, system sqlc.System, auditLogger auditlogger.AuditLogger, org, domain string, teamsService TeamsService, graphClient GraphClient) *githubTeamReconciler {
	return &githubTeamReconciler{
		queries:      queries,
		db:           db,
		system:       system,
		auditLogger:  auditLogger,
		org:          org,
		domain:       domain,
		teamsService: teamsService,
		graphClient:  graphClient,
	}
}

func NewFromConfig(queries sqlc.Querier, db *gorm.DB, cfg *config.Config, system sqlc.System, auditLogger auditlogger.AuditLogger) (reconcilers.Reconciler, error) {
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

	return New(queries, db, system, auditLogger, cfg.GitHub.Organization, cfg.TenantDomain, restClient.Teams, graphClient), nil
}

func (r *githubTeamReconciler) Reconcile(ctx context.Context, input reconcilers.Input) error {
	state := &reconcilers.GitHubState{}
	err := dbmodels.LoadSystemState(r.db, r.system.ID, input.Team.ID, state)
	if err != nil {
		return fmt.Errorf("unable to load system state for team '%s' in system '%s': %w", input.Team.Slug, r.system.Name, err)
	}

	githubTeam, err := r.getOrCreateTeam(ctx, *state, input.Corr, input.Team)
	if err != nil {
		return fmt.Errorf("unable to get or create a GitHub team for team '%s' in system '%s': %w", input.Team.Slug, r.system.Name, err)
	}

	err = dbmodels.SetSystemState(r.db, r.system.ID, input.Team.ID, reconcilers.GitHubState{Slug: githubTeam.Slug})
	if err != nil {
		log.Errorf("system state not persisted: %s", err)
	}

	return r.connectUsers(ctx, githubTeam, input.Corr, input.Team)
}

func (r *githubTeamReconciler) System() sqlc.System {
	return r.system
}

func (r *githubTeamReconciler) getOrCreateTeam(ctx context.Context, state reconcilers.GitHubState, corr sqlc.Correlation, team dbmodels.Team) (*github.Team, error) {
	if state.Slug != nil {
		existingTeam, resp, err := r.teamsService.GetTeamBySlug(ctx, r.org, *state.Slug)
		if resp == nil && err != nil {
			return nil, fmt.Errorf("unable to fetch GitHub team '%s': %w", *state.Slug, err)
		}

		switch resp.StatusCode {
		case http.StatusNotFound:
			break
		case http.StatusOK:
			return existingTeam, nil
		default:
			body, _ := ioutil.ReadAll(resp.Body)
			return nil, fmt.Errorf("server error from GitHub: %s: %s", resp.Status, string(body))
		}
	}

	description := helpers.TeamPurpose(team.Purpose)
	githubTeam, resp, err := r.teamsService.CreateTeam(ctx, r.org, github.NewTeam{
		Name:        string(team.Slug),
		Description: &description,
	})
	err = httpError(http.StatusCreated, *resp, err)
	if err != nil {
		return nil, fmt.Errorf("unable to create GitHub team: %w", err)
	}

	r.auditLogger.Logf(OpCreate, corr, r.system, nil, &team, nil, "created GitHub team '%s'", *githubTeam.Slug)

	return githubTeam, nil
}

func (r *githubTeamReconciler) connectUsers(ctx context.Context, githubTeam *github.Team, corr sqlc.Correlation, team dbmodels.Team) error {
	membersAccordingToGitHub, err := r.getTeamMembers(ctx, *githubTeam.Slug)
	if err != nil {
		return fmt.Errorf("%s: list existing members in GitHub team '%s': %w", OpAddMembers, *githubTeam.Slug, err)
	}

	membersAccordingToConsole := helpers.DomainUsers(team.Users, r.domain)
	consoleUserWithGitHubUser, err := r.mapSSOUsers(ctx, membersAccordingToConsole)
	if err != nil {
		return err
	}

	var targetUser *dbmodels.User
	membersToRemove := remoteOnlyMembers(membersAccordingToGitHub, consoleUserWithGitHubUser)
	for _, gitHubUser := range membersToRemove {
		username := gitHubUser.GetLogin()
		resp, err := r.teamsService.RemoveTeamMembershipBySlug(ctx, r.org, *githubTeam.Slug, username)
		err = httpError(http.StatusNoContent, *resp, err)
		if err != nil {
			log.Warnf("%s: unable to remove member '%s' from GitHub team '%s': %s", OpDeleteMember, username, *githubTeam.Slug, err)
			continue
		}

		targetUser = nil
		email, err := r.getEmailFromGitHubUsername(ctx, username)
		if err != nil {
			log.Warnf("%s: unable to get email from GitHub username '%s' for audit log purposes: %s", OpDeleteMember, username, err)
		}

		if email != nil {
			targetUser = db.GetUserByEmail(r.db, *email)
		}

		r.auditLogger.Logf(OpDeleteMember, corr, r.system, nil, &team, targetUser, "deleted member '%s' from GitHub team '%s'", username, *githubTeam.Slug)
	}

	membersToAdd := localOnlyMembers(consoleUserWithGitHubUser, membersAccordingToGitHub)
	for username, consoleUser := range membersToAdd {
		_, resp, err := r.teamsService.AddTeamMembershipBySlug(ctx, r.org, *githubTeam.Slug, username, &github.TeamAddTeamMembershipOptions{})
		err = httpError(http.StatusOK, *resp, err)
		if err != nil {
			log.Warnf("%s: unable to add member '%s' to GitHub team '%s': %s", OpAddMember, username, *githubTeam.Slug, err)
			continue
		}

		r.auditLogger.Logf(OpAddMember, corr, r.system, nil, &team, consoleUser, "added member '%s' to GitHub team '%s'", username, *githubTeam.Slug)
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
func localOnlyMembers(consoleUsers map[string]*dbmodels.User, membersAccordingToGitHub []*github.User) map[string]*dbmodels.User {
	gitHubUsernameMap := make(map[string]*github.User, 0)
	for _, gitHubUser := range membersAccordingToGitHub {
		gitHubUsernameMap[gitHubUser.GetLogin()] = gitHubUser
	}

	localOnly := make(map[string]*dbmodels.User, 0)
	for gitHubUsername, consoleUser := range consoleUsers {
		if _, exists := gitHubUsernameMap[gitHubUsername]; !exists {
			localOnly[gitHubUsername] = consoleUser
		}
	}
	return localOnly
}

// remoteOnlyMembers Given a list of GitHub team members and a mapping of known GitHub usernames to Console users,
// return members not present in the mapping.
func remoteOnlyMembers(membersAccordingToGitHub []*github.User, consoleUsers map[string]*dbmodels.User) []*github.User {
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
func (r *githubTeamReconciler) mapSSOUsers(ctx context.Context, users []*dbmodels.User) (map[string]*dbmodels.User, error) {
	userMap := make(map[string]*dbmodels.User)
	for _, user := range users {
		githubUsername, err := r.getGitHubUsernameFromEmail(ctx, user.Email)
		if err == errGitHubUserNotFound {
			log.Warnf("%s: no GitHub user for email: '%s'", OpMapSSOUser, user.Email)
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
		body, _ := ioutil.ReadAll(resp.Body)
		return fmt.Errorf("unexpected response error %s: %s", resp.Status, string(body))
	}

	return nil
}
