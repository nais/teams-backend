package graph

import (
	"context"
	"fmt"

	"github.com/99designs/gqlgen/graphql"
	"github.com/google/uuid"
	"github.com/nais/console/pkg/auditlogger"
	"github.com/nais/console/pkg/db"
	"github.com/nais/console/pkg/deployproxy"
	"github.com/nais/console/pkg/graph/apierror"
	"github.com/nais/console/pkg/graph/model"
	"github.com/nais/console/pkg/logger"
	"github.com/nais/console/pkg/reconcilers"
	"github.com/nais/console/pkg/slug"
	"github.com/nais/console/pkg/sqlc"
)

// This file will not be regenerated automatically.
//
// It serves as dependency injection for your app, add any dependencies you require here.

type Resolver struct {
	database        db.Database
	deployProxy     deployproxy.Proxy
	tenantDomain    string
	teamReconciler  chan<- reconcilers.Input
	userSync        chan<- uuid.UUID
	systemName      sqlc.SystemName
	auditLogger     auditlogger.AuditLogger
	gcpEnvironments []string
	log             logger.Logger
}

func NewResolver(database db.Database, deployProxy deployproxy.Proxy, tenantDomain string, teamReconciler chan<- reconcilers.Input, userSync chan<- uuid.UUID, auditLogger auditlogger.AuditLogger, gcpEnvironments []string, log logger.Logger) *Resolver {
	return &Resolver{
		database:        database,
		deployProxy:     deployProxy,
		tenantDomain:    tenantDomain,
		systemName:      sqlc.SystemNameGraphqlApi,
		teamReconciler:  teamReconciler,
		auditLogger:     auditLogger,
		gcpEnvironments: gcpEnvironments,
		log:             log.WithSystem(string(sqlc.SystemNameGraphqlApi)),
		userSync:        userSync,
	}
}

// GetQueriedFields Get a map of queried fields for the given context with the field names as keys
func GetQueriedFields(ctx context.Context) map[string]bool {
	fields := make(map[string]bool)
	for _, field := range graphql.CollectAllFields(ctx) {
		fields[field] = true
	}
	return fields
}

// reconcileTeam Trigger team reconcilers for a given team
func (r *Resolver) reconcileTeam(ctx context.Context, correlationID uuid.UUID, team db.Team) {
	reconcilerInput, err := reconcilers.CreateReconcilerInput(ctx, r.database, team)
	if err != nil {
		r.log.Errorf("unable to generate reconcile input for team %q: %s", team.Slug, err)
		return
	}

	r.teamReconciler <- reconcilerInput.WithCorrelationID(correlationID)
}

// reconcileAllTeams Trigger reconcilers for all teams
func (r *Resolver) reconcileAllTeams(ctx context.Context, correlationID uuid.UUID) ([]*model.TeamSync, error) {
	teams, err := r.database.GetTeams(ctx)
	if err != nil {
		return nil, fmt.Errorf("unable to get teams for reconcile loop: %w", err)
	}

	syncEntries := make([]*model.TeamSync, 0, len(teams))
	for _, team := range teams {
		input, err := reconcilers.CreateReconcilerInput(ctx, r.database, *team)
		if err != nil {
			r.log.WithTeamSlug(string(team.Slug)).WithError(err).Error("unable to create reconciler input")
			continue
		}
		r.teamReconciler <- input.WithCorrelationID(correlationID)
		syncEntries = append(syncEntries, &model.TeamSync{
			Team:          team,
			CorrelationID: &correlationID,
		})
	}

	return syncEntries, nil
}

// getTeam helper to get team by slug, or log if err
func (r *Resolver) getTeamBySlugOrLog(ctx context.Context, slug slug.Slug) (*db.Team, error) {
	team, err := r.database.GetTeamBySlug(ctx, slug)
	if err != nil {
		r.log.WithTeamSlug(string(slug)).WithError(err).Errorf("get team")
		return nil, apierror.ErrTeamNotExist
	}

	return team, nil
}

func sqlcRoleFromTeamRole(teamRole model.TeamRole) (sqlc.RoleName, error) {
	switch teamRole {
	case model.TeamRoleMember:
		return sqlc.RoleNameTeammember, nil
	case model.TeamRoleOwner:
		return sqlc.RoleNameTeamowner, nil
	}

	return "", fmt.Errorf("invalid team role: %v", teamRole)
}
