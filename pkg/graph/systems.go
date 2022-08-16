package graph

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.

import (
	"context"

	"github.com/nais/console/pkg/sqlc"
)

func (r *queryResolver) Systems(ctx context.Context) ([]*sqlc.System, error) {
	systems, err := r.queries.GetSystems(ctx)
	if err != nil {
		return nil, err
	}
	return systems, nil
}
