package graph

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/nais/console/pkg/authz"
	"github.com/nais/console/pkg/db"
	"github.com/nais/console/pkg/fixtures"
	"github.com/nais/console/pkg/graph/generated"
	"github.com/nais/console/pkg/graph/model"
	"github.com/nais/console/pkg/serviceaccount"
	"github.com/nais/console/pkg/sqlc"
)

func (r *mutationResolver) CreateServiceAccount(ctx context.Context, input model.CreateServiceAccountInput) (*db.User, error) {
	actor := authz.UserFromContext(ctx)
	err := authz.RequireGlobalAuthorization(actor, sqlc.AuthzNameServiceAccountsCreate)
	if err != nil {
		return nil, err
	}

	name := string(*input.Name)
	if strings.HasPrefix(name, fixtures.NaisServiceAccountPrefix) {
		return nil, fmt.Errorf("'%s' is a reserved prefix", fixtures.NaisServiceAccountPrefix)
	}

	correlationID, err := uuid.NewUUID()
	if err != nil {
		return nil, fmt.Errorf("unable to create correlation ID for audit log: %w", err)
	}

	serviceAccount, err := r.database.AddServiceAccount(ctx, *input.Name, serviceaccount.Email(*input.Name, r.tenantDomain), actor.ID)
	if err != nil {
		return nil, err
	}

	r.auditLogger.Logf(ctx, sqlc.AuditActionGraphqlApiServiceAccountCreate, correlationID, r.systemName, &actor.Email, nil, &serviceAccount.Email, "Service account created")

	return serviceAccount, nil
}

func (r *mutationResolver) UpdateServiceAccount(ctx context.Context, serviceAccountID *uuid.UUID, input model.UpdateServiceAccountInput) (*db.User, error) {
	panic(fmt.Errorf("not implemented"))
}

func (r *mutationResolver) DeleteServiceAccount(ctx context.Context, serviceAccountID *uuid.UUID) (bool, error) {
	actor := authz.UserFromContext(ctx)
	err := authz.RequireAuthorization(actor, sqlc.AuthzNameServiceAccountsDelete, *serviceAccountID)
	if err != nil {
		return false, err
	}

	serviceAccount, err := r.database.GetUserByID(ctx, *serviceAccountID)
	if err != nil {
		return false, err
	}

	if serviceAccount.Name == fixtures.AdminUserName {
		return false, fmt.Errorf("unable to delete admin account")
	}

	if strings.HasPrefix(serviceAccount.Name, fixtures.NaisServiceAccountPrefix) {
		return false, fmt.Errorf("unable to delete static service account")
	}

	correlationID, err := uuid.NewUUID()
	if err != nil {
		return false, err
	}

	err = r.database.DeleteUser(ctx, serviceAccount.ID)
	if err != nil {
		return false, err
	}

	r.auditLogger.Logf(ctx, sqlc.AuditActionGraphqlApiServiceAccountDelete, correlationID, r.systemName, &actor.Email, nil, &serviceAccount.Email, "Service account deleted")

	return true, nil
}

func (r *queryResolver) Users(ctx context.Context) ([]*db.User, error) {
	actor := authz.UserFromContext(ctx)
	err := authz.RequireGlobalAuthorization(actor, sqlc.AuthzNameUsersList)
	if err != nil {
		return nil, err
	}

	return r.database.GetUsers(ctx)
}

func (r *queryResolver) User(ctx context.Context, id *uuid.UUID) (*db.User, error) {
	actor := authz.UserFromContext(ctx)
	err := authz.RequireGlobalAuthorization(actor, sqlc.AuthzNameUsersList)
	if err != nil {
		return nil, err
	}

	return r.database.GetUserByID(ctx, *id)
}

func (r *queryResolver) Me(ctx context.Context) (*db.User, error) {
	return authz.UserFromContext(ctx), nil
}

func (r *userResolver) Teams(ctx context.Context, obj *db.User) ([]*model.UserTeam, error) {
	actor := authz.UserFromContext(ctx)
	err := authz.RequireGlobalAuthorization(actor, sqlc.AuthzNameTeamsList)
	if err != nil {
		return nil, err
	}

	teams, err := r.database.GetUserTeams(ctx, obj.ID)
	if err != nil {
		return nil, err
	}

	userTeams := make([]*model.UserTeam, 0, len(teams))
	for _, team := range teams {
		isOwner, err := r.database.UserIsTeamOwner(ctx, obj.ID, team.ID)
		if err != nil {
			return nil, err
		}

		role := model.TeamRoleMember
		if isOwner {
			role = model.TeamRoleOwner
		}

		userTeams = append(userTeams, &model.UserTeam{
			Team: team,
			Role: role,
		})
	}

	return userTeams, nil
}

func (r *userResolver) IsServiceAccount(ctx context.Context, obj *db.User) (bool, error) {
	actor := authz.UserFromContext(ctx)
	err := authz.RequireAuthorization(actor, sqlc.AuthzNameServiceAccountsRead, obj.ID)
	if err != nil {
		return false, err
	}

	return serviceaccount.IsServiceAccount(*obj, r.tenantDomain), nil
}

func (r *userResolver) Roles(ctx context.Context, obj *db.User) ([]*db.Role, error) {
	actor := authz.UserFromContext(ctx)
	err := authz.RequireAuthorizationOrTargetMatch(actor, sqlc.AuthzNameUsersUpdate, obj.ID)
	if err != nil {
		return nil, err
	}

	return r.database.GetUserRoles(ctx, obj.ID)
}

// User returns generated.UserResolver implementation.
func (r *Resolver) User() generated.UserResolver { return &userResolver{r} }

type userResolver struct{ *Resolver }
