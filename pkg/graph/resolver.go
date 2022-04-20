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
	tx.Find(obj)
	return tx.Error
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
	tx.Find(obj)
	return tx.Error
}

// Run a query to get data from the database. Populates `collection` and returns pagination metadata.
func (r *Resolver) paginatedQuery(ctx context.Context, input model.Query, dbmodel interface{}, collection interface{}) (*model.Pagination, error) {
	query := input.GetQuery()
	tx := r.db.WithContext(ctx).Model(dbmodel).Where(query)
	pagination := r.withPagination(input, tx)
	tx.Find(collection)
	return pagination, tx.Error
}

// Limit a query by its pagination parameters, count number of rows in dataset, and return pagination metadata.
func (r *Resolver) withPagination(input model.Query, tx *gorm.DB) *model.Pagination {
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
