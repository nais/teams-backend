package graph

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.

import (
	"context"

	"github.com/google/uuid"
	"github.com/nais/console/pkg/graph/generated"
	"github.com/nais/console/pkg/models"
)

func (r *queryResolver) Users(ctx context.Context, id *uuid.UUID, email *string) ([]*models.User, error) {
	query := &models.User{
		Model: models.Model{
			ID: id,
		},
		Email: email,
	}
	users := make([]*models.User, 0)
	tx := r.db.WithContext(ctx).Where(query).Find(&users)
	if tx.Error != nil {
		return nil, tx.Error
	}
	return users, nil
}

// Query returns generated.QueryResolver implementation.
func (r *Resolver) Query() generated.QueryResolver { return &queryResolver{r} }

type queryResolver struct{ *Resolver }
