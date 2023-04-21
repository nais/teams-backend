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
)

type dependencytrackReconciler struct {
	database    db.Database
	auditLogger auditlogger.AuditLogger
	log         logger.Logger
	clients     map[string]dependencytrack.Client
}

// TODO: add to DB
const (
	Name                             = sqlc.ReconcilerName("nais:dependencytrack")
	AuditActionDependencytrackCreate = sqlc.AuditAction("dependencytrack:group:create")
)

func New(database db.Database, auditLogger auditlogger.AuditLogger, clients map[string]dependencytrack.Client, log logger.Logger) (reconcilers.Reconciler, error) {
	return &dependencytrackReconciler{
		database:    database,
		auditLogger: auditLogger,
		log:         log,
		clients:     clients,
	}, nil
}

func NewFromConfig(ctx context.Context, database db.Database, cfg *config.Config, auditLogger auditlogger.AuditLogger, log logger.Logger) (reconcilers.Reconciler, error) {
	log = log.WithSystem(string(Name))
	clients := make(map[string]dependencytrack.Client, 0)
	for _, instance := range cfg.DependencyTrack.Instances {
		client := dependencytrack.NewClient(instance.Endpoint, instance.Username, instance.Password, nil)
		clients[instance.Endpoint] = client
	}
	return New(database, auditLogger, clients, log)
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

	instances := make(map[string]*reconcilers.DependencyTrackTeam)
	members := make([]string, 0)
	for _, member := range input.TeamMembers {
		members = append(members, member.Email)
	}

	for endpoint, client := range r.clients {
		teamId, err := r.syncTeamAndUsers(ctx, input, client, state.Instances[endpoint])
		if err != nil {
			return err
		}
		instances[endpoint] = &reconcilers.DependencyTrackTeam{
			TeamID:  teamId,
			Members: members,
		}
	}

	updateState := &reconcilers.DependencyTrackState{
		Instances: instances,
	}

	err = r.database.SetReconcilerStateForTeam(ctx, r.Name(), input.Team.Slug, updateState)
	if err != nil {
		r.log.WithError(err).Error("persist system state")

	}

	return nil
}

func (r *dependencytrackReconciler) Delete(ctx context.Context, teamSlug slug.Slug, correlationID uuid.UUID) error {
	state := &reconcilers.DependencyTrackState{}
	err := r.database.LoadReconcilerStateForTeam(ctx, r.Name(), teamSlug, state)
	if err != nil {
		return fmt.Errorf("load reconciler state for team %q in reconciler %q: %w", teamSlug, r.Name(), err)
	}

	for endpoint, client := range r.clients {
		team := state.Instances[endpoint]
		if team != nil {
			err = client.DeleteTeam(ctx, state.Instances[endpoint].TeamID)
			if err != nil {
				return err
			}
		}
	}
	return r.database.RemoveReconcilerStateForTeam(ctx, r.Name(), teamSlug)
}

func (r *dependencytrackReconciler) syncTeamAndUsers(ctx context.Context, input reconcilers.Input, client dependencytrack.Client, teamState *reconcilers.DependencyTrackTeam) (string, error) {

	if teamState != nil && teamState.TeamID != "" {
		for _, user := range input.TeamMembers {
			if !contains(teamState.Members, user.Email) {
				err := client.AddToTeam(ctx, user.Email, teamState.TeamID)
				if err != nil {
					return "", err
				}
			}
		}

		for _, user := range teamState.Members {
			if !inputMembersContains(input.TeamMembers, user) {
				err := client.DeleteUserMembership(ctx, teamState.TeamID, user)
				if err != nil {
					return "", err
				}
			}
		}
		return teamState.TeamID, nil
	}

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
		Action:        AuditActionDependencytrackCreate,
		CorrelationID: input.CorrelationID,
	}
	err = r.auditLogger.Logf(ctx, r.database, targets, fields, "Created dependencytrack team %q with ID %q", team.Name, team.Uuid)
	if err != nil {
		return "", err
	}

	for _, user := range input.TeamMembers {
		err = client.CreateUser(ctx, user.Email)
		if err != nil {
			return "", err
		}

		err = client.AddToTeam(ctx, user.Email, team.Uuid)
		if err != nil {
			return "", err
		}
	}

	return team.Uuid, nil
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

func teamByName(teams []dependencytrack.Team, name string) *dependencytrack.Team {
	for _, t := range teams {
		if t.Name == name {
			return &t
		}
	}
	return nil
}
