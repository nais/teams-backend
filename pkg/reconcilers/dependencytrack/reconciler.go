package dependencytrack_reconciler

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/google/uuid"
	dependencytrack "github.com/nais/dependencytrack/pkg/client"
	"github.com/nais/teams-backend/pkg/auditlogger"
	"github.com/nais/teams-backend/pkg/config"
	"github.com/nais/teams-backend/pkg/db"
	dtrack "github.com/nais/teams-backend/pkg/dependencytrack"
	"github.com/nais/teams-backend/pkg/logger"
	"github.com/nais/teams-backend/pkg/metrics"
	"github.com/nais/teams-backend/pkg/reconcilers"
	"github.com/nais/teams-backend/pkg/slug"
	"github.com/nais/teams-backend/pkg/sqlc"
	"github.com/nais/teams-backend/pkg/types"
	"github.com/sirupsen/logrus"
)

type reconciler struct {
	database    db.Database
	auditLogger auditlogger.AuditLogger
	log         logger.Logger
	clients     map[string]dependencytrack.Client
}

const (
	Name = sqlc.ReconcilerNameNaisDependencytrack
)

func New(database db.Database, auditLogger auditlogger.AuditLogger, clients map[string]dependencytrack.Client, log logger.Logger) (reconcilers.Reconciler, error) {
	return &reconciler{
		database:    database,
		auditLogger: auditLogger.WithComponentName(types.ComponentNameNaisDependencytrack),
		log:         log.WithComponent(types.ComponentNameNaisDependencytrack),
		clients:     clients,
	}, nil
}

func NewFromConfig(ctx context.Context, database db.Database, cfg *config.Config, audit auditlogger.AuditLogger, log logger.Logger) (reconcilers.Reconciler, error) {
	clients := make(map[string]dependencytrack.Client, 0)
	if len(cfg.DependencyTrack.Instances) == 0 {
		return nil, fmt.Errorf("no dependencytrack instances configured")
	}

	for _, instance := range cfg.DependencyTrack.Instances {
		c := createClient(ctx, instance, log)
		if c != nil {
			clients[instance.Endpoint] = *c
			log.Infof("dependencytrack instance %q added to reconciler", instance.Endpoint)
		}
	}
	return New(database, audit, clients, log)
}

func (r *reconciler) Name() sqlc.ReconcilerName {
	return Name
}

func (r *reconciler) Reconcile(ctx context.Context, input reconcilers.Input) error {
	state := &reconcilers.DependencyTrackState{}
	err := r.database.LoadReconcilerStateForTeam(ctx, r.Name(), input.Team.Slug, state)
	if err != nil {
		return fmt.Errorf("unable to load reconciler state for team %q in reconciler %q: %w", input.Team.Slug, r.Name(), err)
	}

	updatedInstances := make([]*reconcilers.DependencyTrackInstanceState, 0)
	stateMembers := make([]string, 0)
	for _, member := range input.TeamMembers {
		stateMembers = append(stateMembers, member.Email)
	}

	for endpoint, client := range r.clients {
		r.log.Debugf("reconciling team %q in dependencytrack instance %q", input.Team.Slug, endpoint)
		instance := instanceByEndpoint(state.Instances, endpoint)
		teamId, err := r.syncTeamAndUsers(ctx, input, client, instance)
		if err != nil {
			return err
		}
		updatedInstances = append(updatedInstances, &reconcilers.DependencyTrackInstanceState{
			Endpoint: endpoint,
			TeamID:   teamId,
			Members:  stateMembers,
		})
	}

	updateState := &reconcilers.DependencyTrackState{
		Instances: updatedInstances,
	}

	err = r.database.SetReconcilerStateForTeam(ctx, r.Name(), input.Team.Slug, updateState)
	if err != nil {
		r.log.WithError(err).Error("persist reconciler state")
	}

	return nil
}

func (r *reconciler) Delete(ctx context.Context, teamSlug slug.Slug, _ uuid.UUID) error {
	state := &reconcilers.DependencyTrackState{}
	err := r.database.LoadReconcilerStateForTeam(ctx, r.Name(), teamSlug, state)
	if err != nil {
		return fmt.Errorf("load reconciler state for team %q in reconciler %q: %w", teamSlug, r.Name(), err)
	}

	for endpoint, client := range r.clients {
		instanceState := instanceByEndpoint(state.Instances, endpoint)
		if instanceState != nil {
			err = client.DeleteTeam(ctx, instanceState.TeamID)
			if err != nil {
				return err
			}
		}
	}
	return r.database.RemoveReconcilerStateForTeam(ctx, r.Name(), teamSlug)
}

func createClient(ctx context.Context, instance dtrack.DependencyTrackInstance, log logger.Logger) *dependencytrack.Client {
	client := dependencytrack.New(
		instance.Endpoint,
		instance.Username,
		instance.Password,
		dependencytrack.WithLogger(log.WithFields(logrus.Fields{
			"instance": instance.Endpoint,
		})),
		dependencytrack.WithResponseCallback(incExternalHttpCalls),
	)
	pingCtx, cancel := context.WithTimeout(ctx, time.Second*1)
	defer cancel()
	if _, err := client.Version(pingCtx); err != nil {
		log.Warnf("dependencytrack instance %q is not available, skipping", instance.Endpoint)
		return nil
	}
	return &client
}

func (r *reconciler) syncTeamAndUsers(ctx context.Context, input reconcilers.Input, client dependencytrack.Client, instanceState *reconcilers.DependencyTrackInstanceState) (string, error) {
	if instanceState != nil && instanceState.TeamID != "" {
		r.log.Debugf("team %q already exists in dependencytrack instance state.", input.Team.Slug)
		for _, user := range input.TeamMembers {
			if !contains(instanceState.Members, user.Email) {
				err := client.CreateOidcUser(ctx, user.Email)
				r.log.Debugf("creating user %q in dependencytrack.", user.Email)
				if err != nil {
					return "", err
				}
				err = client.AddToTeam(ctx, user.Email, instanceState.TeamID)
				if err != nil {
					return "", err
				}
				r.log.Debugf("adding user %q to team %q dependencytrack.", user.Email, input.Team.Slug)

				targets := []auditlogger.Target{
					auditlogger.TeamTarget(input.Team.Slug),
					auditlogger.UserTarget(user.Email),
				}
				fields := auditlogger.Fields{
					Action:        types.AuditActionDependencytrackTeamAddMember,
					CorrelationID: input.CorrelationID,
				}
				r.auditLogger.Logf(ctx, r.database, targets, fields, "Added member %q to Dependencytrack team %q", user.Email, input.Team.Slug)
			}
		}

		for _, user := range instanceState.Members {
			if !inputMembersContains(input.TeamMembers, user) {
				err := client.DeleteUserMembership(ctx, instanceState.TeamID, user)
				if err != nil {
					return "", err
				}

				targets := []auditlogger.Target{
					auditlogger.TeamTarget(input.Team.Slug),
					auditlogger.UserTarget(user),
				}
				fields := auditlogger.Fields{
					Action:        types.AuditActionDependencytrackTeamDeleteMember,
					CorrelationID: input.CorrelationID,
				}
				r.auditLogger.Logf(ctx, r.database, targets, fields, "Deleted member %q from Dependencytrack team %q", user, input.Team.Slug)
			}
		}
		return instanceState.TeamID, nil
	}
	r.log.Debugf("team %q does not exist in dependencytrack instance state, creating.", input.Team.Slug)

	team, err := createTeam(ctx, input, client)
	if err != nil {
		return "", err
	}

	targets := []auditlogger.Target{
		auditlogger.TeamTarget(input.Team.Slug),
	}
	fields := auditlogger.Fields{
		Action:        types.AuditActionDependencytrackTeamCreate,
		CorrelationID: input.CorrelationID,
	}
	r.auditLogger.Logf(ctx, r.database, targets, fields, "Created dependencytrack team %q with ID %q", team.Name, team.Uuid)

	r.log.Debugf("created team %q in dependencytrack.", input.Team.Slug)
	for _, user := range input.TeamMembers {
		err = client.CreateOidcUser(ctx, user.Email)
		if err != nil {
			return "", err
		}

		r.log.Debugf("creating user %q in dependencytrack.", user.Email)

		err = client.AddToTeam(ctx, user.Email, team.Uuid)
		if err != nil {
			return "", err
		}
		r.log.Debugf("adding user %q to team %q dependencytrack.", user.Email, input.Team.Slug)

		targets := []auditlogger.Target{
			auditlogger.TeamTarget(input.Team.Slug),
			auditlogger.UserTarget(user.Email),
		}
		fields := auditlogger.Fields{
			Action:        types.AuditActionDependencytrackTeamAddMember,
			CorrelationID: input.CorrelationID,
		}
		r.auditLogger.Logf(ctx, r.database, targets, fields, "Added member %q to Dependencytrack team %q", user.Email, input.Team.Slug)
	}

	return team.Uuid, nil
}

func createTeam(ctx context.Context, input reconcilers.Input, client dependencytrack.Client) (*dependencytrack.Team, error) {
	permissions := []dependencytrack.Permission{
		dependencytrack.ViewPortfolioPermission,
		dependencytrack.ViewVulnerabilityPermission,
		dependencytrack.ViewPolicyViolationPermission,
	}

	if teamIsNaisTeam(input.Team.Slug) {
		extraPermissions := []dependencytrack.Permission{
			dependencytrack.AccessManagementPermission,
			dependencytrack.PolicyManagementPermission,
			dependencytrack.PolicyViolationAnalysisPermission,
			dependencytrack.SystemConfigurationPermission,
		}
		permissions = append(permissions, extraPermissions...)
	}

	team, err := client.CreateTeam(ctx, string(input.Team.Slug), permissions)
	if err != nil {
		return nil, err
	}

	return team, err
}

func incExternalHttpCalls(resp *http.Response, err error) {
	metrics.IncExternalHTTPCalls(string(Name), resp, err)
}

func instanceByEndpoint(instances []*reconcilers.DependencyTrackInstanceState, endpoint string) *reconcilers.DependencyTrackInstanceState {
	for _, i := range instances {
		if i.Endpoint == endpoint {
			return i
		}
	}
	return nil
}

func inputMembersContains(inputMembers []*db.User, user string) bool {
	for _, u := range inputMembers {
		if u.Email == user {
			return true
		}
	}
	return false
}

func contains(ss []string, s string) bool {
	for _, v := range ss {
		if v == s {
			return true
		}
	}
	return false
}

func teamIsNaisTeam(teamSlug slug.Slug) bool {
	return string(teamSlug) == "nais" || string(teamSlug) == "aura"
}
