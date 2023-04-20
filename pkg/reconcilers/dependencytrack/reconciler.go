package dependencytrack

import (
	"context"

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
	client      dependencytrack.Client
}

// TODO: add to DB
const (
	Name                             = sqlc.ReconcilerName("nais:dependencytrack")
	AuditActionDependencytrackCreate = sqlc.AuditAction("dependencytrack:group:create")
)

func New(database db.Database, auditLogger auditlogger.AuditLogger, client dependencytrack.Client, log logger.Logger) (reconcilers.Reconciler, error) {
	return &dependencytrackReconciler{
		database:    database,
		auditLogger: auditLogger,
		log:         log,
		client:      client,
	}, nil
}

func NewFromConfig(ctx context.Context, database db.Database, cfg *config.Config, auditLogger auditlogger.AuditLogger, log logger.Logger) (reconcilers.Reconciler, error) {
	log = log.WithSystem(string(Name))
	c := dependencytrack.NewClient(
		"TODO",
		"TODO",
		"TODO",
		nil,
	)
	return New(database, auditLogger, c, log)
}

func (r *dependencytrackReconciler) Name() sqlc.ReconcilerName {
	return Name
}

func (r *dependencytrackReconciler) Reconcile(ctx context.Context, input reconcilers.Input) error {
	err := r.createTeamAndUsers(ctx, input)
	if err != nil {
		return err
	}

	return nil
}

func (r *dependencytrackReconciler) Delete(ctx context.Context, teamSlug slug.Slug, correlationID uuid.UUID) error {
	teams, err := r.client.GetTeams(ctx)
	if err != nil {
		return err
	}
	teamName := teamSlug.String()

	team := teamByName(teams, teamName)
	err = r.client.DeleteTeam(ctx, team.Uuid)
	if err != nil {
		return err
	}
	return nil
}

func (r *dependencytrackReconciler) createTeamAndUsers(ctx context.Context, input reconcilers.Input) error {

	teams, err := r.client.GetTeams(ctx)
	if err != nil {
		return err
	}
	teamName := input.Team.Slug.String()

	team := teamByName(teams, teamName)
	if team == nil {
		team, err = r.client.CreateTeam(ctx, teamName, []dependencytrack.Permission{
			dependencytrack.ViewPortfolioPermission,
		})
		if err != nil {
			return err
		}

		targets := []auditlogger.Target{
			auditlogger.TeamTarget(input.Team.Slug),
		}
		fields := auditlogger.Fields{
			Action:        AuditActionDependencytrackCreate,
			CorrelationID: input.CorrelationID,
		}
		err := r.auditLogger.Logf(ctx, r.database, targets, fields, "Created dependencytrack team %q with ID %q", team.Name, team.Uuid)
		if err != nil {
			return err
		}
	}

	err = r.deleteUsersNotInConsole(ctx, team, input.TeamMembers)
	if err != nil {
		return err
	}

	for _, user := range input.TeamMembers {
		err = r.client.CreateUser(ctx, user.Email)
		if err != nil {
			return err
		}

		err = r.client.AddToTeam(ctx, user.Email, team.Uuid)
		if err != nil {
			return err
		}

	}
	return nil
}

func (r *dependencytrackReconciler) deleteUsersNotInConsole(ctx context.Context, team *dependencytrack.Team, consoleUsers []*db.User) error {

	usersToRemove := make([]dependencytrack.User, 0)
	for _, u := range team.OidcUsers {
		found := false
		for _, cu := range consoleUsers {
			if u.Username == cu.Email {
				found = true
			}
		}
		if !found {
			usersToRemove = append(usersToRemove, u)
		}
	}

	for _, u := range usersToRemove {
		err := r.client.DeleteUserMembership(ctx, team.Uuid, u.Username)
		if err != nil {
			return err
		}
	}
	return nil
}

func teamByName(teams []dependencytrack.Team, name string) *dependencytrack.Team {
	for _, t := range teams {
		if t.Name == name {
			return &t
		}
	}
	return nil
}
