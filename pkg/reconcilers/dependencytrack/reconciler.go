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
	DpTrack     DpTrack
}

type DpTrack struct {
	Endpoint string
	Client   dependencytrack.Client
}

const (
	Name = sqlc.ReconcilerNameNaisDependencytrack
)

func New(database db.Database, auditLogger auditlogger.AuditLogger, dp DpTrack, log logger.Logger) (reconcilers.Reconciler, error) {
	return &reconciler{
		database:    database,
		auditLogger: auditLogger,
		log:         log.WithComponent(types.ComponentNameNaisDependencytrack),
		DpTrack:     dp,
	}, nil
}

func NewFromConfig(ctx context.Context, database db.Database, cfg *config.Config, log logger.Logger) (reconcilers.Reconciler, error) {
	if cfg.DependencyTrack.Endpoint == "" || cfg.DependencyTrack.Username == "" || cfg.DependencyTrack.Password == "" {
		return nil, fmt.Errorf("no dependencytrack instances configured")
	}

	dp := DpTrack{
		Endpoint: cfg.DependencyTrack.Endpoint,
		Client: dependencytrack.New(
			cfg.DependencyTrack.Endpoint,
			cfg.DependencyTrack.Username,
			cfg.DependencyTrack.Password,
			dependencytrack.WithLogger(log.WithFields(logrus.Fields{
				"instance": cfg.DependencyTrack.Endpoint,
			})),
			dependencytrack.WithResponseCallback(incExternalHttpCalls),
		),
	}
	pingCtx, cancel := context.WithTimeout(ctx, time.Second*1)
	defer cancel()

	if _, err := dp.Client.Version(pingCtx); err != nil {
		log.Warnf("dependencytrack instance %q is not available, skipping", cfg.DependencyTrack.Endpoint)
		return nil, nil
	}
	log.Infof("dependencytrack instance %q added to reconciler", dp.Client)
	return New(database, auditlogger.New(database, types.ComponentNameNaisDependencytrack, log), dp, log)
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

	stateMembers := make([]string, 0)
	for _, member := range input.TeamMembers {
		stateMembers = append(stateMembers, member.Email)
	}

	r.log.Debugf("reconciling team %q in dependencytrack instance %q", input.Team.Slug, r.DpTrack.Client)
	instance := state
	teamId, err := r.syncTeamAndUsers(ctx, input, r.DpTrack.Client, instance)
	if err != nil {
		return err
	}
	updatedInstance := &reconcilers.DependencyTrackState{
		TeamID:  teamId,
		Members: stateMembers,
	}

	err = r.database.SetReconcilerStateForTeam(ctx, r.Name(), input.Team.Slug, updatedInstance)
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

	err = r.DpTrack.Client.DeleteTeam(ctx, state.TeamID)
	if err != nil {
		return err
	}
	return r.database.RemoveReconcilerStateForTeam(ctx, r.Name(), teamSlug)
}

func (r *reconciler) syncTeamAndUsers(ctx context.Context, input reconcilers.Input, client dependencytrack.Client, instanceState *reconcilers.DependencyTrackState) (string, error) {
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
				r.auditLogger.Logf(ctx, targets, fields, "Added member %q to Dependencytrack team %q", user.Email, input.Team.Slug)
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
				r.auditLogger.Logf(ctx, targets, fields, "Deleted member %q from Dependencytrack team %q", user, input.Team.Slug)
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
	r.auditLogger.Logf(ctx, targets, fields, "Created dependencytrack team %q with ID %q", team.Name, team.Uuid)

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
		r.auditLogger.Logf(ctx, targets, fields, "Added member %q to Dependencytrack team %q", user.Email, input.Team.Slug)
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
