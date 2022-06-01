package graph

import (
	"context"
	"fmt"

	"github.com/nais/console/pkg/authz"
	"github.com/nais/console/pkg/dbmodels"
	"github.com/nais/console/pkg/graph/model"
	"gorm.io/gorm"
)

// This file will not be regenerated automatically.
//
// It serves as dependency injection for your app, add any dependencies you require here.

type Resolver struct {
	db      *gorm.DB
	trigger chan<- *dbmodels.Team
	console *dbmodels.System
}

func NewResolver(db *gorm.DB, console *dbmodels.System, trigger chan<- *dbmodels.Team) *Resolver {
	return &Resolver{
		db:      db,
		trigger: trigger,
		console: console,
	}
}

// Enables abstracted access to CreatedBy and UpdatedBy for generic database models.
type Model interface {
	GetModel() *dbmodels.Model
}

// Create a new object in the database, attaching the current user in metadata.
func (r *Resolver) createDB(ctx context.Context, obj Model) error {
	m := obj.GetModel()
	user := authz.UserFromContext(ctx)
	m.CreatedBy = user
	m.UpdatedBy = user
	tx := r.db.WithContext(ctx).Create(obj)
	if tx.Error != nil {
		return tx.Error
	}
	return tx.Find(obj).Error
}

// Update an object in the database, attaching the current user in metadata.
func (r *Resolver) updateDB(ctx context.Context, obj Model) error {
	m := obj.GetModel()
	m.UpdatedBy = authz.UserFromContext(ctx)
	tx := r.db.WithContext(ctx).Updates(obj)
	if tx.Error != nil {
		return tx.Error
	}
	if tx.RowsAffected == 0 {
		return fmt.Errorf("no such %T", obj)
	}
	return tx.Find(obj).Error
}

// Run a query to get data from the database. Populates `collection` and returns pagination metadata.
func (r *Resolver) paginatedQuery(ctx context.Context, input model.Query, sort model.QueryOrder, dbmodel interface{}, collection interface{}) (*model.Pagination, error) {
	tx := r.db.WithContext(ctx).Model(dbmodel).Where(input.GetQuery()).Order(sort.GetOrderString())
	pagination, tx := r.withPagination(input.GetPagination(), tx)
	return pagination, tx.Find(collection).Error
}

// Limit a query by its pagination parameters, count number of rows in dataset, and return pagination metadata.
func (r *Resolver) withPagination(pagination *model.PaginationInput, tx *gorm.DB) (*model.Pagination, *gorm.DB) {
	var count int64
	tx = tx.Count(&count).Limit(pagination.Limit).Offset(pagination.Offset)

	return &model.Pagination{
		Results: int(count),
		Offset:  pagination.Offset,
		Limit:   pagination.Limit,
	}, tx
}
