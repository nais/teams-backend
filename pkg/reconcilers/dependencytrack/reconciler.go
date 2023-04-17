package dependencytrack

import (
	"context"

	"github.com/google/uuid"
	"github.com/nais/console/pkg/auditlogger"
	"github.com/nais/console/pkg/config"
	"github.com/nais/console/pkg/db"
	"github.com/nais/console/pkg/logger"
	"github.com/nais/console/pkg/reconcilers"
	"github.com/nais/console/pkg/slug"
	"github.com/nais/console/pkg/sqlc"
)

type dependencytrackReconciler struct {
	database    db.Database
	auditLogger auditlogger.AuditLogger
	log         logger.Logger
	client      *Client
}

const Name = sqlc.ReconcilerName("nais:dependencytrack")

func NewFromConfig(ctx context.Context, database db.Database, cfg *config.Config, auditLogger auditlogger.AuditLogger, log logger.Logger) (reconcilers.Reconciler, error) {
	log = log.WithSystem(string(Name))
	// TODO: get from config
	c := NewClient("http://localhost:9001/api/v1", "admin", "yolo")
	return &dependencytrackReconciler{
		database:    database,
		auditLogger: auditLogger,
		log:         log,
		client:      c,
	}, nil
}

func (r *dependencytrackReconciler) Name() sqlc.ReconcilerName {
	return Name
}

func (r *dependencytrackReconciler) Reconcile(ctx context.Context, input reconcilers.Input) error {

	// TODO: implement update and delete
	teams, err := r.client.GetTeams(ctx)
	if err != nil {

	}
	if ok := TeamExist(teams, input.Team.Slug.String()); ok {
		//update,sync
	}
	return r.createTeamAndUsers(ctx, input)
}

func (r *dependencytrackReconciler) Delete(ctx context.Context, teamSlug slug.Slug, correlationID uuid.UUID) error {
	team := teamSlug.String()
	err := r.client.DeleteTeam(ctx, team)
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
	team := input.Team.Slug.String()

	uuid := GetTeamUuid(teams, team)
	if uuid == "" {
		team, err := r.client.CreateTeam(ctx, team, []Permission{
			ViewPortfolioPermission,
		})
		if err != nil {
			return err
		}
		uuid = team.Uuid
	}

	for _, user := range input.TeamMembers {
		err = r.client.CreateUser(ctx, user.Email)
		if err != nil {
			return err
		}

		err = r.client.AddToTeam(ctx, user.Email, uuid)
		if err != nil {
			return err
		}

		// TODO: audit log
	}
	return nil
}

func TeamExist(teams []Team, team string) bool {

	if GetTeamUuid(teams, team) == "" {
		return false
	}
	return true
}
