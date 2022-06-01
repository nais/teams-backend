package graph

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.

import (
	"context"

	"github.com/nais/console/pkg/dbmodels"
	"github.com/nais/console/pkg/graph/model"
)

func (r *queryResolver) Systems(ctx context.Context, input *model.QuerySystemsInput, sort *model.QuerySystemsSortInput) (*model.Systems, error) {
	systems := make([]*dbmodels.System, 0)

	if sort == nil {
		sort = &model.QuerySystemsSortInput{
			Field:     model.SystemSortFieldName,
			Direction: model.SortDirectionAsc,
		}
	}
	pagination, err := r.paginatedQuery(ctx, input, sort, &dbmodels.User{}, &systems)

	return &model.Systems{
		Pagination: pagination,
		Nodes:      systems,
	}, err
}
