package graph

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.

import (
	"context"
	"fmt"

	"github.com/nais/console/pkg/dbmodels"
	"github.com/nais/console/pkg/graph/model"
)

func (r *mutationResolver) CreateRole(ctx context.Context, input model.CreateRoleInput) (*dbmodels.Role, error) {
	u := &dbmodels.Role{
		SystemID:    input.SystemID,
		Resource:    input.Resource,
		AccessLevel: input.AccessLevel,
		Permission:  input.Permission,
	}
	err := r.createDB(ctx, u)
	return u, err
}

func (r *mutationResolver) UpdateRole(ctx context.Context, input model.UpdateRoleInput) (*dbmodels.Role, error) {
	u := &dbmodels.Role{
		Model: dbmodels.Model{
			ID: input.ID,
		},
		SystemID: input.SystemID,
	}
	if input.Resource != nil {
		u.Resource = *input.Resource
	}
	if input.AccessLevel != nil {
		u.AccessLevel = *input.AccessLevel
	}
	if input.Permission != nil {
		u.Permission = *input.Permission
	}
	err := r.updateDB(ctx, u)
	return u, err
}

func (r *mutationResolver) AssignRoleToUser(ctx context.Context, input model.AssignRoleInput) (*dbmodels.User, error) {
	panic(fmt.Errorf("not implemented"))
}

func (r *mutationResolver) AssignRoleToTeam(ctx context.Context, input model.AssignRoleInput) (*dbmodels.Team, error) {
	panic(fmt.Errorf("not implemented"))
}

func (r *queryResolver) Roles(ctx context.Context, input *model.QueryRoleInput) (*model.Roles, error) {
	panic(fmt.Errorf("not implemented"))
}
