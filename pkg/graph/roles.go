package graph

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.

import (
	"context"

	"github.com/nais/console/pkg/dbmodels"
	"github.com/nais/console/pkg/graph/generated"
	"github.com/nais/console/pkg/graph/model"
)

func (r *mutationResolver) AssignRoleToUser(ctx context.Context, input model.AssignRoleInput) (*dbmodels.RoleBinding, error) {
	rb := &dbmodels.RoleBinding{
		UserID: input.UserID,
		RoleID: input.RoleID,
		TeamID: input.TeamID,
	}

	tx := r.db.WithContext(ctx).Create(rb)

	if tx.Error != nil {
		return nil, tx.Error
	}
	return rb, nil
}

func (r *mutationResolver) RemoveRoleFromUser(ctx context.Context, input model.RemoveRoleInput) (bool, error) {
	rb := &dbmodels.RoleBinding{}
	tx := r.db.WithContext(ctx).First(rb, "user_id = ? AND team_id = ? AND role_id = ?", input.UserID, input.TeamID, input.RoleID).Delete(rb)
	if tx.Error != nil {
		return false, tx.Error
	}
	return true, nil
}

func (r *queryResolver) Roles(ctx context.Context, input *model.QueryRolesInput, sort *model.QueryRolesSortInput) (*model.Roles, error) {
	roles := make([]*dbmodels.Role, 0)
	if sort == nil {
		sort = &model.QueryRolesSortInput{
			Field:     model.RoleSortFieldName,
			Direction: model.SortDirectionAsc,
		}
	}
	pagination, err := r.paginatedQuery(ctx, input, sort, &dbmodels.Role{}, &roles)
	return &model.Roles{
		Pagination: pagination,
		Nodes:      roles,
	}, err
}

func (r *roleBindingResolver) Role(ctx context.Context, obj *dbmodels.RoleBinding) (*dbmodels.Role, error) {
	var role *dbmodels.Role
	err := r.db.Model(&obj).Association("Role").Find(&role)
	if err != nil {
		return nil, err
	}
	return role, nil
}

func (r *roleBindingResolver) User(ctx context.Context, obj *dbmodels.RoleBinding) (*dbmodels.User, error) {
	var user *dbmodels.User
	err := r.db.Model(&obj).Association("User").Find(&user)
	if err != nil {
		return nil, err
	}
	return user, nil
}

func (r *roleBindingResolver) Team(ctx context.Context, obj *dbmodels.RoleBinding) (*dbmodels.Team, error) {
	if obj.TeamID == nil {
		return nil, nil
	}
	var team *dbmodels.Team
	err := r.db.Model(&obj).Association("Team").Find(&team)
	if err != nil {
		return nil, err
	}
	return team, nil
}

// RoleBinding returns generated.RoleBindingResolver implementation.
func (r *Resolver) RoleBinding() generated.RoleBindingResolver { return &roleBindingResolver{r} }

type roleBindingResolver struct{ *Resolver }
