package graph

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/nais/console/pkg/graph/generated"
	"github.com/nais/console/pkg/graph/model"
	"github.com/nais/console/pkg/models"
)

func (r *Resolver) updateOrBust(ctx context.Context, obj interface{}) error {
	tx := r.db.WithContext(ctx).Updates(obj)
	if tx.Error != nil {
		return tx.Error
	}
	if tx.RowsAffected == 0 {
		return fmt.Errorf("no such %T", obj)
	}
	tx.Find(obj)
	return tx.Error
}

func (r *mutationResolver) CreateUser(ctx context.Context, input model.CreateUserInput) (*models.User, error) {
	u := &models.User{
		Email: input.Email,
		Name:  &input.Name,
	}
	tx := r.db.WithContext(ctx).Create(u)
	if tx.Error != nil {
		return nil, tx.Error
	}
	return u, nil
}

func (r *mutationResolver) UpdateUser(ctx context.Context, input model.UpdateUserInput) (*models.User, error) {
	u := &models.User{
		Model: models.Model{
			ID: input.ID,
		},
		Email: input.Email,
		Name:  input.Name,
	}
	err := r.updateOrBust(ctx, u)
	return u, err
}

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

// Mutation returns generated.MutationResolver implementation.
func (r *Resolver) Mutation() generated.MutationResolver { return &mutationResolver{r} }

// Query returns generated.QueryResolver implementation.
func (r *Resolver) Query() generated.QueryResolver { return &queryResolver{r} }

type mutationResolver struct{ *Resolver }
type queryResolver struct{ *Resolver }
