package graph

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.

import (
	"context"

	"github.com/nais/console/pkg/dbmodels"
	"github.com/nais/console/pkg/graph/model"
)

func (r *mutationResolver) CreateUser(ctx context.Context, input model.CreateUserInput) (*dbmodels.User, error) {
	u := &dbmodels.User{
		Email: input.Email,
		Name:  &input.Name,
	}
	tx := r.db.WithContext(ctx).Create(u)
	if tx.Error != nil {
		return nil, tx.Error
	}
	return u, nil
}

func (r *mutationResolver) UpdateUser(ctx context.Context, input model.UpdateUserInput) (*dbmodels.User, error) {
	u := &dbmodels.User{
		Model: dbmodels.Model{
			ID: input.ID,
		},
		Email: input.Email,
		Name:  input.Name,
	}
	actor, err := r.user(ctx)
	if err != nil {
		return nil, err
	}
	u.UpdatedBy = actor
	err = r.updateOrBust(ctx, u)
	return u, err
}

func (r *queryResolver) Users(ctx context.Context, input *model.QueryUserInput) (*model.Users, error) {
	query := input.Query()
	users := make([]*dbmodels.User, 0)
	tx := r.db.WithContext(ctx).Model(&dbmodels.User{}).Where(query)
	pagination := r.withPagination(input, tx)
	tx.Find(&users)
	if tx.Error != nil {
		return nil, tx.Error
	}
	return &model.Users{
		Pagination: pagination,
		Nodes:      users,
	}, nil
}
