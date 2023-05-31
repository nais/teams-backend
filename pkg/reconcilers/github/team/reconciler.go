package github_team_reconciler

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"sort"
	"strings"

	"github.com/google/go-github/v50/github"
	"github.com/google/uuid"
	"github.com/nais/teams-backend/pkg/auditlogger"
	"github.com/nais/teams-backend/pkg/config"
	"github.com/nais/teams-backend/pkg/db"
	"github.com/nais/teams-backend/pkg/helpers"
	"github.com/nais/teams-backend/pkg/logger"
	"github.com/nais/teams-backend/pkg/metrics"
	"github.com/nais/teams-backend/pkg/reconcilers"
	"github.com/nais/teams-backend/pkg/slug"
	"github.com/nais/teams-backend/pkg/sqlc"
	"github.com/shurcooL/githubv4"
	"google.golang.org/api/impersonate"
)

const (
	Name              = sqlc.ReconcilerNameGithubTeam
	metricsSystemName = "github"
)

var errGitHubUserNotFound = errors.New("GitHub user does not exist")

func New(database db.Database, auditLogger auditlogger.AuditLogger, org, domain string, teamsService TeamsService, graphClient GraphClient, log logger.Logger) *githubTeamReconciler {
	return &githubTeamReconciler{
		database:     database,
		auditLogger:  auditLogger.WithSystemName(sqlc.SystemNameGithubTeam),
		org:          org,
		domain:       domain,
		teamsService: teamsService,
		graphClient:  graphClient,
		log:          log.WithSystem(string(Name)),
	}
}

func NewFromConfig(ctx context.Context, database db.Database, cfg *config.Config, auditLogger auditlogger.AuditLogger, log logger.Logger) (reconcilers.Reconciler, error) {
	if cfg.GitHub.AuthEndpoint == "" {
		return nil, fmt.Errorf("missing required configuration: TEAMS_BACKEND_GITHUB_AUTH_ENDPOINT")
	}

	if cfg.GoogleManagementProjectID == "" {
		return nil, fmt.Errorf("missing required configuration: TEAMS_BACKEND_GOOGLE_MANAGEMENT_PROJECT_ID")
	}

	ts, err := impersonate.IDTokenSource(ctx, impersonate.IDTokenConfig{
		Audience:        cfg.GitHub.AuthEndpoint,
		TargetPrincipal: fmt.Sprintf("console@%s.iam.gserviceaccount.com", cfg.GoogleManagementProjectID),
	})
	if err != nil {
		return nil, err
	}

	httpClient := NewGitHubAuthClient(ctx, cfg.GitHub.AuthEndpoint, ts)
	return New(database, auditLogger, cfg.GitHub.Organization, cfg.TenantDomain, github.NewClient(httpClient).Teams, githubv4.NewClient(httpClient), log), nil
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

	teamSlug := slug.Slug(*githubTeam.Slug)
	err = r.database.SetReconcilerStateForTeam(ctx, r.Name(), input.Team.Slug, reconcilers.GitHubState{
		Slug:         &teamSlug,
		Repositories: repos,
	})
	if err != nil {
		r.log.WithError(err).Error("persist system state")
	}

	return r.connectUsers(ctx, githubTeam, input)
}

func (r *githubTeamReconciler) Delete(ctx context.Context, teamSlug slug.Slug, correlationID uuid.UUID) error {
	state := &reconcilers.GitHubState{}
	err := r.database.LoadReconcilerStateForTeam(ctx, r.Name(), teamSlug, state)
	if err != nil {
		return fmt.Errorf("load reconciler state for team %q in reconciler %q: %w", teamSlug, r.Name(), err)
	}

	if state.Slug == nil {
		r.log.Warnf("missing slug in reconciler state for team %q in reconciler %q, assume already deleted", teamSlug, r.Name())
		return r.database.RemoveReconcilerStateForTeam(ctx, r.Name(), teamSlug)
	}

	gitHubTeamSlug := *state.Slug

	resp, err := r.teamsService.DeleteTeamBySlug(ctx, r.org, string(gitHubTeamSlug))
	metrics.IncExternalHTTPCalls(metricsSystemName, unwrapResponse(resp), err)
	if err != nil {
		return fmt.Errorf("delete GitHub team %q for team %q: %w", gitHubTeamSlug, teamSlug, err)
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("unexpected server response from GitHub: %q: %q", resp.Status, string(body))
	}

	targets := []auditlogger.Target{
		auditlogger.TeamTarget(teamSlug),
	}
	fields := auditlogger.Fields{
		Action:        sqlc.AuditActionGithubTeamDelete,
		CorrelationID: correlationID,
	}
	r.auditLogger.Logf(ctx, r.database, targets, fields, "Delete GitHub team with slug %q", gitHubTeamSlug)

	return r.database.RemoveReconcilerStateForTeam(ctx, r.Name(), teamSlug)
}

func (r *githubTeamReconciler) syncTeamInfo(ctx context.Context, team db.Team, githubTeam github.Team) error {
	var slug string

	if gitHubTeamIsUpToDate(team, githubTeam) {
		return nil
	}

	slug = *githubTeam.Slug
	newTeam := github.NewTeam{
		Name:        slug,
		Description: &team.Purpose,
		Privacy:     helpers.Strp("closed"),
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
		// FIXME: https://github.com/nais/teams-backend/issues/77
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
		Privacy:     helpers.Strp("closed"),
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

	teamsBackendUserWithGitHubUser, err := r.mapSSOUsers(ctx, input.TeamMembers)
	if err != nil {
		return err
	}

	membersToRemove := remoteOnlyMembers(membersAccordingToGitHub, teamsBackendUserWithGitHubUser)
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
				r.log.WithError(err).Warnf("get teams-backend user with email %q", *email)
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

	membersToAdd := localOnlyMembers(teamsBackendUserWithGitHubUser, membersAccordingToGitHub)
	for username, teamsBackendUser := range membersToAdd {
		_, resp, err := r.teamsService.AddTeamMembershipBySlug(ctx, r.org, *githubTeam.Slug, username, &github.TeamAddTeamMembershipOptions{})
		metrics.IncExternalHTTPCalls(metricsSystemName, unwrapResponse(resp), err)
		err = httpError(http.StatusOK, resp, err)
		if err != nil {
			r.log.WithError(err).Warnf("add member %q to GitHub team %q", username, *githubTeam.Slug)
			continue
		}

		targets := []auditlogger.Target{
			auditlogger.TeamTarget(input.Team.Slug),
			auditlogger.UserTarget(teamsBackendUser.Email),
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

// localOnlyMembers Given a mapping of GitHub usernames to teams-backend users, and a list of GitHub team members according to
// GitHub, return members only present in the mapping.
func localOnlyMembers(teamsBackendUsers map[string]*db.User, membersAccordingToGitHub []*github.User) map[string]*db.User {
	gitHubUsernameMap := make(map[string]*github.User, 0)
	for _, gitHubUser := range membersAccordingToGitHub {
		gitHubUsernameMap[gitHubUser.GetLogin()] = gitHubUser
	}

	localOnly := make(map[string]*db.User, 0)
	for gitHubUsername, teamsBackendUser := range teamsBackendUsers {
		if _, exists := gitHubUsernameMap[gitHubUsername]; !exists {
			localOnly[gitHubUsername] = teamsBackendUser
		}
	}
	return localOnly
}

// remoteOnlyMembers Given a list of GitHub team members and a mapping of known GitHub usernames to teams-backend users,
// return members not present in the mapping.
func remoteOnlyMembers(membersAccordingToGitHub []*github.User, teamsBackendUsers map[string]*db.User) []*github.User {
	unknownMembers := make([]*github.User, 0)
	for _, member := range membersAccordingToGitHub {
		if _, exists := teamsBackendUsers[member.GetLogin()]; !exists {
			unknownMembers = append(unknownMembers, member)
		}
	}
	return unknownMembers
}

// mapSSOUsers Return a mapping of GitHub usernames to teams-backend user objects. teams-backend users with no matching
// GitHub user will be ignored.
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

// gitHubTeamIsUpToDate check if a GitHub team is up to date compared to the teams-backend team
func gitHubTeamIsUpToDate(naisTeam db.Team, gitHubTeam github.Team) bool {
	if naisTeam.Purpose != helpers.StringWithFallback(gitHubTeam.Description, "") {
		return false
	}

	if gitHubTeam.GetPrivacy() != "closed" {
		return false
	}

	return true
}
