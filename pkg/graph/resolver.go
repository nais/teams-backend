package graph

import (
	"context"
	"fmt"

	"github.com/99designs/gqlgen/graphql"
	"github.com/google/uuid"
	"github.com/nais/teams-backend/pkg/auditlogger"
	"github.com/nais/teams-backend/pkg/db"
	"github.com/nais/teams-backend/pkg/deployproxy"
	"github.com/nais/teams-backend/pkg/graph/apierror"
	"github.com/nais/teams-backend/pkg/graph/model"
	"github.com/nais/teams-backend/pkg/logger"
	"github.com/nais/teams-backend/pkg/slug"
	"github.com/nais/teams-backend/pkg/sqlc"
	"github.com/nais/teams-backend/pkg/teamsync"
	"github.com/nais/teams-backend/pkg/types"
	"github.com/nais/teams-backend/pkg/usersync"
)

// This file will not be regenerated automatically.
//
// It serves as dependency injection for your app, add any dependencies you require here.

type Resolver struct {
	teamSyncHandler teamsync.Handler
	database        db.Database
	deployProxy     deployproxy.Proxy
	tenantDomain    string
	userSync        chan<- uuid.UUID
	systemName      types.ComponentName
	auditLogger     auditlogger.AuditLogger
	gcpEnvironments []string
	log             logger.Logger
	userSyncRuns    *usersync.RunsHandler
}

func NewResolver(teamSyncHandler teamsync.Handler, database db.Database, deployProxy deployproxy.Proxy, tenantDomain string, userSync chan<- uuid.UUID, auditLogger auditlogger.AuditLogger, gcpEnvironments []string, log logger.Logger, userSyncRuns *usersync.RunsHandler) *Resolver {
	return &Resolver{
		teamSyncHandler: teamSyncHandler,
		database:        database,
		deployProxy:     deployProxy,
		tenantDomain:    tenantDomain,
		systemName:      types.ComponentNameGraphqlApi,
		auditLogger:     auditLogger,
		gcpEnvironments: gcpEnvironments,
		log:             log.WithComponent(types.ComponentNameGraphqlApi),
		userSync:        userSync,
		userSyncRuns:    userSyncRuns,
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

// addTeamToReconcilerQueue add a team (enclosed in an input) to the reconciler queue
func (r *Resolver) addTeamToReconcilerQueue(input teamsync.Input) error {
	err := r.teamSyncHandler.Schedule(input)
	if err != nil {
		r.log.WithTeamSlug(string(input.TeamSlug)).WithError(err).Errorf("add team to reconciler queue")
		return apierror.Errorf("teams-backend is about to restart, unable to reconcile team: %q", input.TeamSlug)
	}
	return nil
}

// reconcileTeam Trigger team reconcilers for a given team
func (r *Resolver) reconcileTeam(_ context.Context, correlationID uuid.UUID, slug slug.Slug) error {
	input := teamsync.Input{
		TeamSlug:      slug,
		CorrelationID: correlationID,
	}

	return r.addTeamToReconcilerQueue(input)
}

func (r *Resolver) getTeamBySlug(ctx context.Context, slug slug.Slug) (*db.Team, error) {
	team, err := r.database.GetTeamBySlug(ctx, slug)
	if err != nil {
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
