package graph

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/nais/console/pkg/authz"
	"github.com/nais/console/pkg/console"
	"github.com/nais/console/pkg/db"
	"github.com/nais/console/pkg/dbmodels"
	"github.com/nais/console/pkg/fixtures"
	"github.com/nais/console/pkg/graph/generated"
	"github.com/nais/console/pkg/graph/model"
	console_reconciler "github.com/nais/console/pkg/reconcilers/console"
	"github.com/nais/console/pkg/roles"
	"gorm.io/gorm"
)

func (r *mutationResolver) CreateServiceAccount(ctx context.Context, input model.CreateServiceAccountInput) (*dbmodels.User, error) {
	actor := authz.UserFromContext(ctx)
	err := roles.RequireGlobalAuthorization(actor, roles.AuthorizationServiceAccountsCreate)
	if err != nil {
		return nil, err
	}

	name := input.Name.String()
	if strings.HasPrefix(name, fixtures.NaisServiceAccountPrefix) {
		return nil, fmt.Errorf("'%s' is a reserved prefix", fixtures.NaisServiceAccountPrefix)
	}

	corr := &dbmodels.Correlation{}
	serviceAccount := &dbmodels.User{
		Name:  name,
		Email: console.ServiceAccountEmail(*input.Name, r.tenantDomain),
	}

	err = r.db.Transaction(func(tx *gorm.DB) error {
		err = tx.Create(corr).Error
		if err != nil {
			return fmt.Errorf("unable to create correlation for audit log")
		}

		err = db.CreateTrackedObject(ctx, tx, serviceAccount)
		if err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	r.auditLogger.Logf(console_reconciler.OpCreateServiceAccount, *corr, *r.system, actor, nil, serviceAccount, "Service account created")

	return serviceAccount, nil
}

func (r *mutationResolver) UpdateServiceAccount(ctx context.Context, serviceAccountID *uuid.UUID, input model.UpdateServiceAccountInput) (*dbmodels.User, error) {
	actor := authz.UserFromContext(ctx)
	err := roles.RequireAuthorization(actor, roles.AuthorizationServiceAccountsUpdate, *serviceAccountID)
	if err != nil {
		return nil, err
	}

	serviceAccount := &dbmodels.User{}
	err = r.db.Where("id = ?", serviceAccountID).First(serviceAccount).Error
	if err != nil {
		return nil, err
	}

	if serviceAccount.Name == fixtures.AdminUserName {
		return nil, fmt.Errorf("unable to update admin account")
	}

	serviceAccount.Name = string(*input.Name)
	serviceAccount.Email = console.ServiceAccountEmail(*input.Name, r.tenantDomain)

	err = db.UpdateTrackedObject(ctx, r.db, serviceAccount)
	if err != nil {
		return nil, err
	}

	return serviceAccount, nil
}

func (r *mutationResolver) DeleteServiceAccount(ctx context.Context, serviceAccountID *uuid.UUID) (bool, error) {
	serviceAccount := &dbmodels.User{}
	err := r.db.Where("id = ?", serviceAccountID).First(serviceAccount).Error
	if err != nil {
		return false, err
	}

	if serviceAccount.Name == fixtures.AdminUserName {
		return false, fmt.Errorf("unable to delete admin account")
	}

	err = r.db.Delete(serviceAccount).Error
	if err != nil {
		return false, err
	}

	return true, nil
}

func (r *queryResolver) Users(ctx context.Context, pagination *model.Pagination, query *model.UsersQuery, sort *model.UsersSort) (*model.Users, error) {
	users := make([]*dbmodels.User, 0)

	if sort == nil {
		sort = &model.UsersSort{
			Field:     model.UserSortFieldName,
			Direction: model.SortDirectionAsc,
		}
	}
	pageInfo, err := r.paginatedQuery(pagination, query, sort, &dbmodels.User{}, &users)

	return &model.Users{
		PageInfo: pageInfo,
		Nodes:    users,
	}, err
}

func (r *queryResolver) User(ctx context.Context, id *uuid.UUID) (*dbmodels.User, error) {
	user := &dbmodels.User{}
	err := r.db.Where("id = ?", id).First(user).Error
	if err != nil {
		return nil, err
	}
	return user, nil
}

func (r *queryResolver) Me(ctx context.Context) (*dbmodels.User, error) {
	return authz.UserFromContext(ctx), nil
}

func (r *userResolver) Teams(ctx context.Context, obj *dbmodels.User) ([]*dbmodels.Team, error) {
	teams := make([]*dbmodels.Team, 0)
	err := r.db.Model(obj).Association("Teams").Find(&teams)
	if err != nil {
		return nil, err
	}
	return teams, nil
}

func (r *userResolver) HasAPIKey(ctx context.Context, obj *dbmodels.User) (bool, error) {
	apiKey := &dbmodels.ApiKey{}
	err := r.db.Where("user_id = ?", obj.ID).First(&apiKey).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return false, nil
		}
		return false, err
	}

	return true, nil
}

func (r *userResolver) IsServiceAccount(ctx context.Context, obj *dbmodels.User) (bool, error) {
	return console.IsServiceAccount(*obj, r.tenantDomain), nil
}

// User returns generated.UserResolver implementation.
func (r *Resolver) User() generated.UserResolver { return &userResolver{r} }

type userResolver struct{ *Resolver }
