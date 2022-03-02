package graph

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.

import (
	"context"

	"github.com/nais/console/pkg/dbmodels"
	"github.com/nais/console/pkg/graph/model"
	"gorm.io/gorm"
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
	err := r.updateOrBust(ctx, u)
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

// !!! WARNING !!!
// The code below was going to be deleted when updating resolvers. It has been copied here so you have
// one last chance to move it out of harms way if you want. There are two reasons this happens:
//  - When renaming or deleting a resolver the old code will be put in here. You can safely delete
//    it when you're done.
//  - You have helper methods in this file. Move them out to keep these resolver files clean.
func (r *Resolver) withPagination(input model.PaginatedQuery, tx *gorm.DB) *model.Pagination {
	var count int64
	in := input.GetPagination()
	tx.Count(&count)
	tx.Limit(in.Limit)
	tx.Offset(in.Offset)
	return &model.Pagination{
		Results: int(count),
		Offset:  in.Offset,
		Limit:   in.Limit,
	}
}
