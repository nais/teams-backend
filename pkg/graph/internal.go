package graph

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.
// Code generated by github.com/99designs/gqlgen version v0.17.36

import (
	"context"

	"github.com/nais/teams-backend/pkg/db"
	"github.com/nais/teams-backend/pkg/graph/generated"
	"github.com/nais/teams-backend/pkg/graph/model"
	"github.com/nais/teams-backend/pkg/sqlc"
)

// TeamsInternal is the resolver for the teamsInternal field.
func (r *queryResolver) TeamsInternal(ctx context.Context) (*model.TeamsInternal, error) {
	return &model.TeamsInternal{}, nil
}

// Name is the resolver for the name field.
func (r *roleResolver) Name(ctx context.Context, obj *db.Role) (sqlc.RoleName, error) {
	return obj.RoleName, nil
}

// Roles is the resolver for the roles field.
func (r *teamsInternalResolver) Roles(ctx context.Context, obj *model.TeamsInternal) ([]sqlc.RoleName, error) {
	return sqlc.AllRoleNameValues(), nil
}

// Role returns generated.RoleResolver implementation.
func (r *Resolver) Role() generated.RoleResolver { return &roleResolver{r} }

// TeamsInternal returns generated.TeamsInternalResolver implementation.
func (r *Resolver) TeamsInternal() generated.TeamsInternalResolver { return &teamsInternalResolver{r} }

type roleResolver struct{ *Resolver }
type teamsInternalResolver struct{ *Resolver }
