package graph

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.

import (
	"context"

	"github.com/nais/console/pkg/dbmodels"
	"github.com/nais/console/pkg/graph/model"
)

func (r *queryResolver) Systems(ctx context.Context, pagination *model.Pagination, query *model.SystemsQuery, sort *model.SystemsSort) (*model.Systems, error) {
	systems := make([]*dbmodels.System, 0)

	if sort == nil {
		sort = &model.SystemsSort{
			Field:     model.SystemSortFieldName,
			Direction: model.SortDirectionAsc,
		}
	}
	pageInfo, err := r.paginatedQuery(ctx, pagination, query, sort, &dbmodels.User{}, &systems)

	return &model.Systems{
		PageInfo: pageInfo,
		Nodes:    systems,
	}, err
}
