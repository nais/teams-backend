package graph

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/nais/console/pkg/auditlogger"
	"github.com/nais/console/pkg/authz"
	"github.com/nais/console/pkg/db"
	"github.com/nais/console/pkg/graph/dataloader"
	"github.com/nais/console/pkg/graph/generated"
	"github.com/nais/console/pkg/graph/model"
	"github.com/nais/console/pkg/roles"
	"github.com/nais/console/pkg/sqlc"
	"github.com/nais/console/pkg/usersync"
)

// SynchronizeUsers is the resolver for the synchronizeUsers field.
func (r *mutationResolver) SynchronizeUsers(ctx context.Context) (*uuid.UUID, error) {
	actor := authz.ActorFromContext(ctx)
	err := authz.RequireGlobalAuthorization(actor, roles.AuthorizationUsersyncSynchronize)
	if err != nil {
		return nil, err
	}

	correlationID, err := uuid.NewUUID()
	if err != nil {
		return nil, fmt.Errorf("create log correlation ID: %w", err)
	}

	targets := []auditlogger.Target{
		auditlogger.SystemTarget(sqlc.SystemNameUsersync),
	}
	fields := auditlogger.Fields{
		Action:        sqlc.AuditActionGraphqlApiUsersSync,
		Actor:         actor,
		CorrelationID: correlationID,
	}
	r.auditLogger.Logf(ctx, r.database, targets, fields, "Trigger user sync")
	r.userSync <- correlationID

	return &correlationID, nil
}

// Users is the resolver for the users field.
func (r *queryResolver) Users(ctx context.Context) ([]*db.User, error) {
	actor := authz.ActorFromContext(ctx)
	err := authz.RequireGlobalAuthorization(actor, roles.AuthorizationUsersList)
	if err != nil {
		return nil, err
	}

	return r.database.GetUsers(ctx)
}

// User is the resolver for the user field.
func (r *queryResolver) User(ctx context.Context, id *uuid.UUID) (*db.User, error) {
	actor := authz.ActorFromContext(ctx)
	err := authz.RequireGlobalAuthorization(actor, roles.AuthorizationUsersList)
	if err != nil {
		return nil, err
	}

	return dataloader.GetUser(ctx, id)
}

// UserByEmail is the resolver for the userByEmail field.
func (r *queryResolver) UserByEmail(ctx context.Context, email string) (*db.User, error) {
	actor := authz.ActorFromContext(ctx)
	err := authz.RequireGlobalAuthorization(actor, roles.AuthorizationUsersList)
	if err != nil {
		return nil, err
	}

	return r.database.GetUserByEmail(ctx, email)
}

// UserSync is the resolver for the userSync field.
func (r *queryResolver) UserSync(ctx context.Context) ([]*usersync.Run, error) {
	actor := authz.ActorFromContext(ctx)
	err := authz.RequireGlobalAuthorization(actor, roles.AuthorizationUsersyncSynchronize)
	if err != nil {
		return nil, err
	}

	return r.userSyncRuns.GetRuns(), nil
}

// Teams is the resolver for the teams field.
func (r *userResolver) Teams(ctx context.Context, obj *db.User) ([]*model.TeamMembership, error) {
	actor := authz.ActorFromContext(ctx)
	err := authz.RequireGlobalAuthorization(actor, roles.AuthorizationTeamsList)
	if err != nil {
		return nil, err
	}

	userRoles, err := dataloader.GetUserRoles(ctx, obj.ID)
	if err != nil {
		return nil, err
	}

	teams := make([]*model.TeamMembership, 0)
	for _, role := range userRoles {
		if role.TargetTeamSlug == nil {
			continue
		}

		var teamRole model.TeamRole
		switch role.RoleName {
		case sqlc.RoleNameTeammember:
			teamRole = model.TeamRoleMember
		case sqlc.RoleNameTeamowner:
			teamRole = model.TeamRoleOwner
		default:
			continue
		}

		team, err := dataloader.GetTeam(ctx, role.TargetTeamSlug)
		if err != nil {
			return nil, err
		}
		teams = append(teams, &model.TeamMembership{
			Team: team,
			Role: teamRole,
		})

	}

	return teams, nil
}

// Roles is the resolver for the roles field.
func (r *userResolver) Roles(ctx context.Context, obj *db.User) ([]*db.Role, error) {
	actor := authz.ActorFromContext(ctx)
	err := authz.RequireRole(actor, sqlc.RoleNameAdmin)
	if err != nil && actor.User.GetID() != obj.ID {
		return nil, err
	}

	userRoles, err := dataloader.GetUserRoles(ctx, obj.ID)
	if err != nil {
		return nil, err
	}
	ret := make([]*db.Role, 0)
	for _, ur := range userRoles {
		authz, err := roles.Authorizations(ur.RoleName)
		if err != nil {
			return nil, err
		}

		var tsa *uuid.UUID
		ntsa := ur.TargetServiceAccountID
		if ntsa.Valid {
			tsa = &ntsa.UUID
		}

		ret = append(ret, &db.Role{
			Authorizations:         authz,
			RoleName:               ur.RoleName,
			TargetServiceAccountID: tsa,
			TargetTeamSlug:         ur.TargetTeamSlug,
		})
	}

	return ret, nil
}

// LogEntries is the resolver for the logEntries field.
func (r *userSyncRunResolver) LogEntries(ctx context.Context, obj *usersync.Run) ([]*db.AuditLog, error) {
	return r.database.GetAuditLogsForCorrelationID(ctx, obj.CorrelationID())
}

// Status is the resolver for the status field.
func (r *userSyncRunResolver) Status(ctx context.Context, obj *usersync.Run) (model.UserSyncRunStatus, error) {
	switch obj.Status() {
	case usersync.RunSuccess:
		return model.UserSyncRunStatusSuccess, nil
	case usersync.RunFailure:
		return model.UserSyncRunStatusFailure, nil
	default:
		return model.UserSyncRunStatusInProgress, nil
	}
}

// Error is the resolver for the error field.
func (r *userSyncRunResolver) Error(ctx context.Context, obj *usersync.Run) (*string, error) {
	err := obj.Error()
	if err != nil {
		msg := err.Error()
		return &msg, nil
	}

	return nil, nil
}

// User returns generated.UserResolver implementation.
func (r *Resolver) User() generated.UserResolver { return &userResolver{r} }

// UserSyncRun returns generated.UserSyncRunResolver implementation.
func (r *Resolver) UserSyncRun() generated.UserSyncRunResolver { return &userSyncRunResolver{r} }

type (
	userResolver        struct{ *Resolver }
	userSyncRunResolver struct{ *Resolver }
)
