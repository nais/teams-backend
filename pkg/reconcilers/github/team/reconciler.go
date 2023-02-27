package github_team_reconciler

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"sort"
	"strconv"
	"strings"

	"github.com/bradleyfalzon/ghinstallation/v2"
	"github.com/google/go-github/v50/github"
	"github.com/google/uuid"
	"github.com/nais/console/pkg/auditlogger"
	"github.com/nais/console/pkg/config"
	helpers "github.com/nais/console/pkg/console"
	"github.com/nais/console/pkg/db"
	"github.com/nais/console/pkg/logger"
	"github.com/nais/console/pkg/metrics"
	"github.com/nais/console/pkg/reconcilers"
	"github.com/nais/console/pkg/slug"
	"github.com/nais/console/pkg/sqlc"
	"github.com/shurcooL/githubv4"
)

const (
	Name              = sqlc.ReconcilerNameGithubTeam
	metricsSystemName = "github"
)

var errGitHubUserNotFound = errors.New("GitHub user does not exist")

func New(database db.Database, auditLogger auditlogger.AuditLogger, org, domain string, teamsService TeamsService, graphClient GraphClient, log logger.Logger) *githubTeamReconciler {
	return &githubTeamReconciler{
		database:     database,
		auditLogger:  auditLogger,
		org:          org,
		domain:       domain,
		teamsService: teamsService,
		graphClient:  graphClient,
		log:          log,
	}
}

func NewFromConfig(ctx context.Context, database db.Database, cfg *config.Config, auditLogger auditlogger.AuditLogger, log logger.Logger) (reconcilers.Reconciler, error) {
	log = log.WithSystem(string(Name))

	config, err := getReconfilerConfig(ctx, cfg, database)
	if err != nil {
		return nil, err
	}

	transport, err := ghinstallation.NewKeyFromFile(
		http.DefaultTransport,
		config.appID,
		config.installationID,
		config.privateKeyPath,
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

	return New(database, auditLogger, config.org, cfg.TenantDomain, restClient.Teams, graphClient, log), nil
}

func (r *githubTeamReconciler) Name() sqlc.ReconcilerName {
	return Name
}

func (r *githubTeamReconciler) Reconcile(ctx context.Context, input reconcilers.Input) error {
	state := &reconcilers.GitHubState{}
	err := r.database.LoadReconcilerStateForTeam(ctx, r.Name(), input.Team.Slug, state)
	if err != nil {
		return fmt.Errorf("unable to load system state for team %q in system %q: %w", input.Team.Slug, r.Name(), err)
	}

	githubTeam, err := r.getOrCreateTeam(ctx, *state, input.CorrelationID, input.Team)
	if err != nil {
		return fmt.Errorf("unable to get or create a GitHub team for team %q in system %q: %w", input.Team.Slug, r.Name(), err)
	}

	err = r.removeTeamIDPSync(ctx, *githubTeam.Slug)
	if err != nil {
		return err
	}

	err = r.syncTeamInfo(ctx, input.Team, *githubTeam)
	if err != nil {
		return err
	}

	repos, err := r.getTeamRepositories(ctx, *githubTeam.Slug)
	if err != nil {
		return err
	}

	slug := slug.Slug(*githubTeam.Slug)
	updatedState := reconcilers.GitHubState{
		Slug:         &slug,
		Repositories: repos,
	}

	err = r.database.SetReconcilerStateForTeam(ctx, r.Name(), input.Team.Slug, updatedState)
	if err != nil {
		r.log.WithError(err).Error("persist system state")
	}

	return r.connectUsers(ctx, githubTeam, input)
}

func (r *githubTeamReconciler) Delete(ctx context.Context, teamSlug slug.Slug, correlationID uuid.UUID) error {
	return nil
}

func (r *githubTeamReconciler) syncTeamInfo(ctx context.Context, team db.Team, githubTeam github.Team) error {
	var slug string

	if team.Purpose == helpers.StringWithFallback(githubTeam.Description, "") {
		return nil
	}

	slug = *githubTeam.Slug
	newTeam := github.NewTeam{
		Name:        slug,
		Description: &team.Purpose,
	}

	_, resp, err := r.teamsService.EditTeamBySlug(ctx, r.org, slug, newTeam, false)
	metrics.IncExternalHTTPCalls(metricsSystemName, unwrapResponse(resp), err)

	if resp == nil && err != nil {
		return fmt.Errorf("sync team info for GitHub team %q: %w", slug, err)
	}

	if resp.StatusCode >= 300 {
		return fmt.Errorf("sync team info for GitHub team %q: %s", slug, resp.Status)
	}

	return nil
}

func (r *githubTeamReconciler) removeTeamIDPSync(ctx context.Context, slug string) error {
	grpList := github.IDPGroupList{
		Groups: make([]*github.IDPGroup, 0),
	}
	idpList, resp, err := r.teamsService.CreateOrUpdateIDPGroupConnectionsBySlug(ctx, r.org, slug, grpList)
	metrics.IncExternalHTTPCalls(metricsSystemName, unwrapResponse(resp), err)
	if err != nil && strings.Contains(err.Error(), "team is not externally managed") {
		// Special case: org has not been configured for team IDP sync, which we don't want to treat as an error
		// FIXME: https://github.com/nais/console/issues/77
		return nil
	}

	if resp == nil && err != nil {
		return fmt.Errorf("unable to delete IDP sync from GitHub team %q: %w", slug, err)
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unable to delete IDP sync for GitHub team %q: %s", slug, resp.Status)
	}

	if len(idpList.Groups) > 0 {
		return fmt.Errorf("tried to delete IDP sync from GitHub team %q, but %d connections still remain", slug, len(idpList.Groups))
	}

	return nil
}

func (r *githubTeamReconciler) getOrCreateTeam(ctx context.Context, state reconcilers.GitHubState, correlationID uuid.UUID, team db.Team) (*github.Team, error) {
	slug := team.Slug.String()

	if state.Slug != nil {
		slug = state.Slug.String()

		existingTeam, resp, err := r.teamsService.GetTeamBySlug(ctx, r.org, string(*state.Slug))
		metrics.IncExternalHTTPCalls(metricsSystemName, unwrapResponse(resp), err)
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

	githubTeam, resp, err := r.teamsService.CreateTeam(ctx, r.org, github.NewTeam{
		Name:        slug,
		Description: &team.Purpose,
	})
	metrics.IncExternalHTTPCalls(metricsSystemName, unwrapResponse(resp), err)
	err = httpError(http.StatusCreated, resp, err)
	if err != nil {
		return nil, fmt.Errorf("unable to create GitHub team: %w", err)
	}

	targets := []auditlogger.Target{
		auditlogger.TeamTarget(team.Slug),
	}
	fields := auditlogger.Fields{
		Action:        sqlc.AuditActionGithubTeamCreate,
		CorrelationID: correlationID,
	}
	r.auditLogger.Logf(ctx, r.database, targets, fields, "Created GitHub team %q", *githubTeam.Slug)

	return githubTeam, nil
}

func (r *githubTeamReconciler) connectUsers(ctx context.Context, githubTeam *github.Team, input reconcilers.Input) error {
	membersAccordingToGitHub, err := r.getTeamMembers(ctx, *githubTeam.Slug)
	if err != nil {
		return fmt.Errorf("list existing members in GitHub team %q: %w", *githubTeam.Slug, err)
	}

	consoleUserWithGitHubUser, err := r.mapSSOUsers(ctx, input.TeamMembers)
	if err != nil {
		return err
	}

	membersToRemove := remoteOnlyMembers(membersAccordingToGitHub, consoleUserWithGitHubUser)
	for _, gitHubUser := range membersToRemove {
		username := gitHubUser.GetLogin()
		resp, err := r.teamsService.RemoveTeamMembershipBySlug(ctx, r.org, *githubTeam.Slug, username)
		metrics.IncExternalHTTPCalls(metricsSystemName, unwrapResponse(resp), err)
		err = httpError(http.StatusNoContent, resp, err)
		if err != nil {
			r.log.WithError(err).Warnf("remove member %q from GitHub team %q", username, *githubTeam.Slug)
			continue
		}

		email, err := r.getEmailFromGitHubUsername(ctx, username)
		if err != nil {
			r.log.WithError(err).Warnf("get email from GitHub username %q for audit log purposes", username)
		}

		if email != nil {
			_, err = r.database.GetUserByEmail(ctx, *email)
			if err != nil {
				r.log.WithError(err).Warnf("get Console user with email %q", *email)
				email = nil
			}
		}

		targets := []auditlogger.Target{
			auditlogger.TeamTarget(input.Team.Slug),
		}
		fields := auditlogger.Fields{
			Action:        sqlc.AuditActionGithubTeamDeleteMember,
			CorrelationID: input.CorrelationID,
		}
		r.auditLogger.Logf(ctx, r.database, targets, fields, "Deleted member %q from GitHub team %q", username, *githubTeam.Slug)
	}

	membersToAdd := localOnlyMembers(consoleUserWithGitHubUser, membersAccordingToGitHub)
	for username, consoleUser := range membersToAdd {
		_, resp, err := r.teamsService.AddTeamMembershipBySlug(ctx, r.org, *githubTeam.Slug, username, &github.TeamAddTeamMembershipOptions{})
		metrics.IncExternalHTTPCalls(metricsSystemName, unwrapResponse(resp), err)
		err = httpError(http.StatusOK, resp, err)
		if err != nil {
			r.log.WithError(err).Warnf("add member %q to GitHub team %q", username, *githubTeam.Slug)
			continue
		}

		targets := []auditlogger.Target{
			auditlogger.TeamTarget(input.Team.Slug),
			auditlogger.UserTarget(consoleUser.Email),
		}
		fields := auditlogger.Fields{
			Action:        sqlc.AuditActionGithubTeamAddMember,
			CorrelationID: input.CorrelationID,
		}
		r.auditLogger.Logf(ctx, r.database, targets, fields, "Added member %q to GitHub team %q", username, *githubTeam.Slug)
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
		metrics.IncExternalHTTPCalls(metricsSystemName, unwrapResponse(resp), err)
		err = httpError(http.StatusOK, resp, err)
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
			r.log.WithError(err).Warnf("no GitHub user for email: %q", user.Email)
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
	metrics.IncExternalCallsByError(metricsSystemName, err)
	if err != nil {
		return nil, err
	}

	nodes := query.Organization.SamlIdentityProvider.ExternalIdentities.Nodes
	if len(nodes) == 0 {
		return nil, errGitHubUserNotFound
	}

	username := string(nodes[0].User.Login)
	if len(username) == 0 {
		return nil, errGitHubUserNotFound
	}

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
	metrics.IncExternalCallsByError(metricsSystemName, err)
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

func (r *githubTeamReconciler) getTeamRepositories(ctx context.Context, teamSlug string) ([]*reconcilers.GitHubRepository, error) {
	const maxPerPage = 100
	opts := &github.ListOptions{
		PerPage: maxPerPage,
	}

	allRepos := make([]*reconcilers.GitHubRepository, 0)
	for {
		repos, resp, err := r.teamsService.ListTeamReposBySlug(ctx, r.org, teamSlug, opts)
		metrics.IncExternalHTTPCalls(metricsSystemName, unwrapResponse(resp), err)
		err = httpError(http.StatusOK, resp, err)
		if err != nil {
			return nil, err
		}
		for _, repo := range repos {
			permissions := make([]*reconcilers.GitHubRepositoryPermission, 0)
			for name, granted := range repo.GetPermissions() {
				permissions = append(permissions, &reconcilers.GitHubRepositoryPermission{
					Name:    name,
					Granted: granted,
				})
			}

			sort.SliceStable(permissions, func(a, b int) bool {
				return permissions[a].Name < permissions[b].Name
			})

			allRepos = append(allRepos, &reconcilers.GitHubRepository{
				Name:        repo.GetFullName(),
				Permissions: permissions,
			})
		}
		if resp.NextPage == 0 {
			break
		}
		opts.Page = resp.NextPage
	}

	sort.SliceStable(allRepos, func(a, b int) bool {
		return allRepos[a].Name < allRepos[b].Name
	})

	return allRepos, nil
}

func getReconfilerConfig(ctx context.Context, cfg *config.Config, database db.Database) (*reconcilerConfig, error) {
	config, err := database.DangerousGetReconcilerConfigValues(ctx, Name)
	if err != nil {
		return nil, err
	}

	installationID, err := strconv.ParseInt(config.GetValue(sqlc.ReconcilerConfigKeyGithubAppInstallationID), 10, 64)
	if err != nil {
		return nil, fmt.Errorf("unable to convert app installation ID %q to an integer", config.GetValue(sqlc.ReconcilerConfigKeyGithubAppInstallationID))
	}

	return &reconcilerConfig{
		appID:          cfg.GitHub.ApplicationID,
		privateKeyPath: cfg.GitHub.PrivateKeyPath,
		org:            config.GetValue(sqlc.ReconcilerConfigKeyGithubOrg),
		installationID: installationID,
	}, nil
}

// httpError Return an error if the response status code is not as expected, or if the passed err is already set to an
// error
func httpError(expected int, resp *github.Response, err error) error {
	if err != nil {
		return err
	}

	if resp == nil {
		return fmt.Errorf("no response")
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

func unwrapResponse(resp *github.Response) *http.Response {
	if resp == nil {
		return nil
	}
	return resp.Response
}
