package graph

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.

import (
	"context"

	"github.com/nais/console/pkg/dbmodels"
	"github.com/nais/console/pkg/graph/generated"
	"github.com/nais/console/pkg/graph/model"
)

func (r *queryResolver) Roles(ctx context.Context, pagination *model.Pagination, query *model.RolesQuery, sort *model.RolesSort) (*model.Roles, error) {
	roles := make([]*dbmodels.Role, 0)

	if sort == nil {
		sort = &model.RolesSort{
			Field:     model.RoleSortFieldName,
			Direction: model.SortDirectionAsc,
		}
	}
	pageInfo, err := r.paginatedQuery(pagination, query, sort, &dbmodels.Role{}, &roles)

	return &model.Roles{
		PageInfo: pageInfo,
		Nodes:    roles,
	}, err
}

func (r *roleBindingResolver) Role(ctx context.Context, obj *dbmodels.UserRole) (*dbmodels.Role, error) {
	role := &dbmodels.Role{}
	err := r.db.Where("id = ?", obj.RoleID).First(role).Error
	if err != nil {
		return nil, err
	}
	return role, nil
}

func (r *roleBindingResolver) IsGlobal(ctx context.Context, obj *dbmodels.UserRole) (bool, error) {
	return obj.TargetID == nil, nil
}

// RoleBinding returns generated.RoleBindingResolver implementation.
func (r *Resolver) RoleBinding() generated.RoleBindingResolver { return &roleBindingResolver{r} }

type roleBindingResolver struct{ *Resolver }
