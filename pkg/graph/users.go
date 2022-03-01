package graph

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.

import (
	"context"

	"github.com/nais/console/pkg/graph/model"
	"github.com/nais/console/pkg/models"
)

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

func (r *queryResolver) Users(ctx context.Context, input *model.QueryUserInput) (*model.Users, error) {
	query := input.Query()
	users := make([]*models.User, 0)
	tx := r.db.WithContext(ctx).Where(query).Find(&users)
	if tx.Error != nil {
		return nil, tx.Error
	}
	return &model.Users{
		Meta: &model.Meta{
			NumResults: len(users),
		},
		Nodes: users,
	}, nil
}
