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

func (r *roleResolver) Authorizations(ctx context.Context, obj *dbmodels.Role) ([]*dbmodels.Authorization, error) {
	authorizations := make([]*dbmodels.Authorization, 0)
	err := r.db.Model(obj).Association("Authorizations").Find(&authorizations)
	if err != nil {
		return nil, err
	}
	return authorizations, nil
}

// Role returns generated.RoleResolver implementation.
func (r *Resolver) Role() generated.RoleResolver { return &roleResolver{r} }

type roleResolver struct{ *Resolver }
