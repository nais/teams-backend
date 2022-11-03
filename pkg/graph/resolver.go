package graph

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/nais/console/pkg/auditlogger"
	"github.com/nais/console/pkg/db"
	"github.com/nais/console/pkg/graph/model"
	"github.com/nais/console/pkg/reconcilers"
	"github.com/nais/console/pkg/sqlc"
	log "github.com/sirupsen/logrus"
)

// This file will not be regenerated automatically.
//
// It serves as dependency injection for your app, add any dependencies you require here.

type Resolver struct {
	database       db.Database
	tenantDomain   string
	teamReconciler chan<- reconcilers.Input
	systemName     sqlc.SystemName
	auditLogger    auditlogger.AuditLogger
}

func NewResolver(database db.Database, tenantDomain string, teamReconciler chan<- reconcilers.Input, auditLogger auditlogger.AuditLogger) *Resolver {
	return &Resolver{
		database:       database,
		tenantDomain:   tenantDomain,
		systemName:     sqlc.SystemNameGraphqlApi,
		teamReconciler: teamReconciler,
		auditLogger:    auditLogger,
	}
}

// reconcileTeam Trigger team reconcilers for a given team
func (r *Resolver) reconcileTeam(ctx context.Context, correlationID uuid.UUID, team db.Team) {
	reconcilerInput, err := reconcilers.CreateReconcilerInput(ctx, r.database, team)
	if err != nil {
		log.Errorf("unable to generate reconcile input for team %q: %s", team.Slug, err)
		return
	}

	r.teamReconciler <- reconcilerInput.WithCorrelationID(correlationID)
}

// reconcileAllTeams Trigger reconcilers for all teams
func (r *Resolver) reconcileAllTeams(ctx context.Context, correlationID uuid.UUID) {
	teams, err := r.database.GetTeams(ctx)
	if err != nil {
		log.Errorf("unable to get teams for reconcile loop: %s", err)
		return
	}

	for _, team := range teams {
		input, err := reconcilers.CreateReconcilerInput(ctx, r.database, *team)
		if err != nil {
			log.Errorf("unable to create input for team %q for reconcile loop: %s", team.Slug, err)
			return
		}

		r.teamReconciler <- input.WithCorrelationID(correlationID)
	}
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
