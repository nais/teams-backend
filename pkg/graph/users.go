package graph

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/nais/console/pkg/authz"
	"github.com/nais/console/pkg/console"
	"github.com/nais/console/pkg/dbmodels"
	"github.com/nais/console/pkg/graph/generated"
	"github.com/nais/console/pkg/graph/model"
	"gorm.io/gorm"
)

func (r *mutationResolver) CreateServiceAccount(ctx context.Context, input model.CreateServiceAccountInput) (*dbmodels.User, error) {
	email := console.ServiceAccountEmail(*input.Name, r.partnerDomain)
	sa := &dbmodels.User{
		Name:  input.Name.String(),
		Email: &email,
	}
	err := r.createTrackedObject(ctx, sa)
	if err != nil {
		return nil, err
	}
	return sa, nil
}

func (r *queryResolver) Users(ctx context.Context, pagination *model.Pagination, query *model.UsersQuery, sort *model.UsersSort) (*model.Users, error) {
	users := make([]*dbmodels.User, 0)

	if sort == nil {
		sort = &model.UsersSort{
			Field:     model.UserSortFieldName,
			Direction: model.SortDirectionAsc,
		}
	}
	pageInfo, err := r.paginatedQuery(ctx, pagination, query, sort, &dbmodels.User{}, &users)

	return &model.Users{
		PageInfo: pageInfo,
		Nodes:    users,
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

func (r *userResolver) RoleBindings(ctx context.Context, obj *dbmodels.User, teamID *uuid.UUID) ([]*dbmodels.RoleBinding, error) {
	roleBindings := make([]*dbmodels.RoleBinding, 0)
	where := &dbmodels.RoleBinding{
		TeamID: teamID,
	}
	err := r.db.WithContext(ctx).Model(obj).Where(where).Association("RoleBindings").Find(&roleBindings)
	if err != nil {
		return nil, err
	}
	return roleBindings, nil
}

func (r *userResolver) HasAPIKey(ctx context.Context, obj *dbmodels.User) (bool, error) {
	apiKey := &dbmodels.ApiKey{}
	err := r.db.WithContext(ctx).Where("user_id = ?", obj.ID).First(&apiKey).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return false, nil
		}
		return false, err
	}

	return true, nil
}

func (r *userResolver) IsServiceAccount(ctx context.Context, obj *dbmodels.User) (bool, error) {
	return console.IsServiceAccount(*obj, r.partnerDomain), nil
}

// User returns generated.UserResolver implementation.
func (r *Resolver) User() generated.UserResolver { return &userResolver{r} }

type userResolver struct{ *Resolver }
