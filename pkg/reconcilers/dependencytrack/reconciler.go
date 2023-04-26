package dependencytrack

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/nais/console/pkg/auditlogger"
	"github.com/nais/console/pkg/config"
	"github.com/nais/console/pkg/db"
	"github.com/nais/console/pkg/dependencytrack"
	"github.com/nais/console/pkg/logger"
	"github.com/nais/console/pkg/reconcilers"
	"github.com/nais/console/pkg/slug"
	"github.com/nais/console/pkg/sqlc"
	log "github.com/sirupsen/logrus"
)

type dependencytrackReconciler struct {
	database    db.Database
	auditLogger auditlogger.AuditLogger
	log         logger.Logger
	clients     map[string]dependencytrack.Client
}

// TODO: add to DB
const (
	Name = sqlc.ReconcilerNameNaisDependencytrack
)

func New(database db.Database, auditLogger auditlogger.AuditLogger, clients map[string]dependencytrack.Client, log logger.Logger) (reconcilers.Reconciler, error) {
	return &dependencytrackReconciler{
		database:    database,
		auditLogger: auditLogger.WithSystemName(sqlc.SystemNameNaisDependencytrack),
		log:         log,
		clients:     clients,
	}, nil
}

func NewFromConfig(_ context.Context, database db.Database, cfg *config.Config, audit auditlogger.AuditLogger, log logger.Logger) (reconcilers.Reconciler, error) {
	log = log.WithSystem(string(Name))
	clients := make(map[string]dependencytrack.Client, 0)
	if len(cfg.DependencyTrack.Instances) == 0 {
		return nil, fmt.Errorf("no dependencytrack instances configured")
	}

	for _, instance := range cfg.DependencyTrack.Instances {
		client := dependencytrack.NewClient(instance.Endpoint, instance.Username, instance.Password, nil, log)
		clients[instance.Endpoint] = client
	}
	return New(database, audit, clients, log)
}

func (r *dependencytrackReconciler) Name() sqlc.ReconcilerName {
	return Name
}

func (r *dependencytrackReconciler) Reconcile(ctx context.Context, input reconcilers.Input) error {
	state := &reconcilers.DependencyTrackState{}
	err := r.database.LoadReconcilerStateForTeam(ctx, r.Name(), input.Team.Slug, state)
	if err != nil {
		return fmt.Errorf("unable to load system state for team %q in system %q: %w", input.Team.Slug, r.Name(), err)
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
		r.log.WithError(err).Error("persist system state")

	}

	return nil
}

func (r *dependencytrackReconciler) Delete(ctx context.Context, teamSlug slug.Slug, _ uuid.UUID) error {
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

func (r *dependencytrackReconciler) syncTeamAndUsers(ctx context.Context, input reconcilers.Input, client dependencytrack.Client, instanceState *reconcilers.DependencyTrackInstanceState) (string, error) {

	if instanceState != nil && instanceState.TeamID != "" {
		log.Debugf("team %q already exists in dependencytrack instance state.", input.Team.Slug)
		for _, user := range input.TeamMembers {
			if !contains(instanceState.Members, user.Email) {
				err := client.CreateUser(ctx, user.Email)
				log.Debugf("creating user %q in dependencytrack.", user.Email)
				if err != nil {
					return "", err
				}
				err = client.AddToTeam(ctx, user.Email, instanceState.TeamID)
				if err != nil {
					return "", err
				}
				log.Debugf("adding user %q to team %q dependencytrack.", user.Email, input.Team.Slug)
			}
		}

		for _, user := range instanceState.Members {
			if !inputMembersContains(input.TeamMembers, user) {
				err := client.DeleteUserMembership(ctx, instanceState.TeamID, user)
				if err != nil {
					return "", err
				}
			}
		}
		return instanceState.TeamID, nil
	}
	log.Debugf("team %q does not exist in dependencytrack instance state, creating.", input.Team.Slug)

	team, err := client.CreateTeam(ctx, input.Team.Slug.String(), []dependencytrack.Permission{
		dependencytrack.ViewPortfolioPermission,
	})
	if err != nil {
		return "", err
	}

	targets := []auditlogger.Target{
		auditlogger.TeamTarget(input.Team.Slug),
	}
	fields := auditlogger.Fields{
		Action:        sqlc.AuditActionDependencytrackGroupCreate,
		CorrelationID: input.CorrelationID,
	}
	err = r.auditLogger.Logf(ctx, r.database, targets, fields, "Created dependencytrack team %q with ID %q", team.Name, team.Uuid)
	if err != nil {
		return "", err
	}

	log.Debugf("created team %q in dependencytrack.", input.Team.Slug)
	for _, user := range input.TeamMembers {
		err = client.CreateUser(ctx, user.Email)
		if err != nil {
			return "", err
		}

		log.Debugf("creating user %q in dependencytrack.", user.Email)

		err = client.AddToTeam(ctx, user.Email, team.Uuid)
		if err != nil {
			return "", err
		}
		log.Debugf("adding user %q to team %q dependencytrack.", user.Email, input.Team.Slug)
	}

	return team.Uuid, nil
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
