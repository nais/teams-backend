package graph

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.

import (
	"context"

	"github.com/google/uuid"
	"github.com/nais/console/pkg/authz"
	"github.com/nais/console/pkg/dbmodels"
	"github.com/nais/console/pkg/graph/generated"
	"github.com/nais/console/pkg/graph/model"
)

func (r *mutationResolver) CreateUser(ctx context.Context, input model.CreateUserInput) (*dbmodels.User, error) {
	u := &dbmodels.User{
		Email: input.Email,
		Name:  &input.Name,
	}
	err := r.createDB(ctx, u)
	return u, err
}

func (r *mutationResolver) UpdateUser(ctx context.Context, input model.UpdateUserInput) (*dbmodels.User, error) {
	u := &dbmodels.User{
		Model: dbmodels.Model{
			ID: input.ID,
		},
		Email: input.Email,
		Name:  input.Name,
	}
	err := r.updateDB(ctx, u)
	return u, err
}

func (r *queryResolver) Users(ctx context.Context, input *model.QueryUsersInput, sort *model.QueryUsersSortInput) (*model.Users, error) {
	users := make([]*dbmodels.User, 0)

	if sort == nil {
		sort = &model.QueryUsersSortInput{
			Field:     model.UserSortFieldName,
			Direction: model.SortDirectionAsc,
		}
	}
	pagination, err := r.paginatedQuery(ctx, input, sort, &dbmodels.User{}, &users)

	return &model.Users{
		Pagination: pagination,
		Nodes:      users,
	}, err
}

func (r *queryResolver) User(ctx context.Context, id *uuid.UUID) (*dbmodels.User, error) {
	user := &dbmodels.User{}
	tx := r.db.WithContext(ctx).Where("id = ?", id).First(user)
	if tx.Error != nil {
		return nil, tx.Error
	}
	return user, nil
}

func (r *queryResolver) Me(ctx context.Context) (*dbmodels.User, error) {
	return authz.UserFromContext(ctx), nil
}

func (r *userResolver) Teams(ctx context.Context, obj *dbmodels.User) ([]*dbmodels.Team, error) {
	teams := make([]*dbmodels.Team, 0)
	err := r.db.WithContext(ctx).Model(obj).Association("Teams").Find(&teams)
	if err != nil {
		return nil, err
	}
	return teams, nil
}

func (r *userResolver) Roles(ctx context.Context, obj *dbmodels.User, teamID *uuid.UUID) ([]*dbmodels.Role, error) {
	roleBindings := make([]*dbmodels.RoleBinding, 0)
	roles := make([]*dbmodels.Role, 0)

	tx := r.db.WithContext(ctx).Model(&dbmodels.RoleBinding{}).Preload("Role").Find(&roleBindings, "user_id = ? AND team_id = ?", obj.ID, *teamID)
	if tx.Error != nil {
		return nil, tx.Error
	}

	for _, rb := range roleBindings {
		roles = append(roles, rb.Role)
	}

	return roles, nil
}

// User returns generated.UserResolver implementation.
func (r *Resolver) User() generated.UserResolver { return &userResolver{r} }

type userResolver struct{ *Resolver }
