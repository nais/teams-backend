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

func NewFromConfig(ctx context.Context, cfg *config.Config) (reconcilers.Reconciler, error) {
	//log = log.WithSystem(string(Name))
	// TODO: get from config
	c := NewClient("http://localhost:9001/api/v1", "admin", "yolo")
	return &dependencytrackReconciler{
		client: c,
	}, nil
}

func (r *dependencytrackReconciler) Name() sqlc.ReconcilerName {
	return Name
}

func (r *dependencytrackReconciler) Reconcile(ctx context.Context, input reconcilers.Input) error {

	return r.createTeamAndUsers(ctx, input)
}

func (r *dependencytrackReconciler) Delete(ctx context.Context, teamSlug slug.Slug, correlationID uuid.UUID) error {
	teams, err := r.client.GetTeams(ctx)
	if err != nil {
		return err
	}
	teamName := teamSlug.String()

	team := GetTeam(teams, teamName)
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

	team := GetTeam(teams, teamName)
	uuid := team.Uuid

	if uuid == "" {
		team, err := r.client.CreateTeam(ctx, teamName, []Permission{
			ViewPortfolioPermission,
		})
		if err != nil {
			return err
		}
		uuid = team.Uuid
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

		err = r.client.AddToTeam(ctx, user.Email, uuid)
		if err != nil {
			return err
		}

		// TODO: audit log
	}
	return nil
}

func (r *dependencytrackReconciler) deleteUsersNotInConsole(ctx context.Context, team *Team, consoleUsers []*db.User) error {

	usersToRemove := make([]User, 0)
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
